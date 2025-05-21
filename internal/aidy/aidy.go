package aidy

import (
	"fmt"
	"hash/fnv"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	msemver "github.com/Masterminds/semver/v3"
	"github.com/volodya-lombrozo/aidy/internal/ai"
	"github.com/volodya-lombrozo/aidy/internal/cache"
	"github.com/volodya-lombrozo/aidy/internal/config"
	"github.com/volodya-lombrozo/aidy/internal/git"
	"github.com/volodya-lombrozo/aidy/internal/github"
	"github.com/volodya-lombrozo/aidy/internal/output"
	"golang.org/x/mod/semver"
)

func Release(step string, repo string, gs git.Git, brain ai.AI, out output.Output) error {
	tags, err := gs.Tags(repo)
	if err != nil {
		return fmt.Errorf("failed to get tags: '%v'", err)
	}
	log.Printf("tags found: %v, size %d\n", tags, len(tags))
	var notes string
	var updated string
	if len(tags) > 0 {
		mtags := clearTags(tags)
		latest := latest(keys(mtags))
		messages, err := gs.Log(mtags[latest])
		if err != nil {
			return fmt.Errorf("failed to get git log: '%v'", err)
		}
		summary := strings.Join(messages, "\n")
		notes, err = brain.ReleaseNotes(summary)
		if err != nil {
			return fmt.Errorf("failed to generate release notes: '%v'", err)
		}
		updated, err = upver(latest, step)
		if err != nil {
			return fmt.Errorf("failed to update version: '%v'", err)
		}
		if strings.HasPrefix(mtags[latest], "v") {
			updated = "v" + updated
		}
	} else {
		updated = "v0.0.1"
		messages, err := gs.Log("")
		if err != nil {
			return fmt.Errorf("failed to get git log: '%v'", err)
		}
		summary := strings.Join(messages, "\n")
		notes, err = brain.ReleaseNotes(summary)
		if err != nil {
			return fmt.Errorf("failed to generate release notes: '%v'", err)
		}
	}
	command := fmt.Sprintf("git tag --cleanup=verbatim -a \"%s\" -m \"%s\" ", updated, notes)
	return out.Print(command)
}

func keys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func upver(ver string, step string) (string, error) {
	mver, err := msemver.NewVersion(ver)
	if err != nil {
		return "", fmt.Errorf("failed to parse version: %v", err)
	}
	var newver msemver.Version
	switch step {
	case "patch":
		newver = mver.IncPatch()
	case "minor":
		newver = mver.IncMinor()
	case "major":
		newver = mver.IncMajor()
	default:
		return "", fmt.Errorf("unknown version step: '%s'", step)
	}
	return newver.String(), nil
}

func clearTags(tags []string) map[string]string {
	if len(tags) == 0 {
		return nil
	}
	res := make(map[string]string)
	for i := range tags {
		if tags[i] == "" {
			continue
		}
		if strings.HasPrefix(tags[i], "v") {
			res[tags[i]] = tags[i]
		} else {
			res["v"+tags[i]] = tags[i]
		}
	}
	return res
}

func latest(tags []string) string {
	sort.Slice(tags, func(i, j int) bool {
		return semver.Compare(tags[i], tags[j]) < 0
	})
	return tags[len(tags)-1]
}

func Diff(gs git.Git) {
	diff, err := gs.Diff()
	if err != nil {
		log.Fatalf("Failed to get diff: '%v'", err)
	} else {
		fmt.Printf("Diff with the base branch:\n%s\n", diff)
	}
}

func Pconfig(cfg config.Config) {
	fmt.Println("Current Configuration:")
	apiKey, err := cfg.GetOpenAIAPIKey()
	if err != nil {
		fmt.Printf("Error retrieving OpenAI API key: %v\n", err)
	} else {
		fmt.Printf("OpenAI API Key: %s\n", apiKey)
	}

	deepseek, err := cfg.GetDeepseekAPIKey()
	if err != nil {
		fmt.Printf("Error retrieving deepseek API key: %v\n", err)
	} else {
		fmt.Printf("Deepseek API Key: %s\n", deepseek)
	}

	githubKey, err := cfg.GetGithubAPIKey()
	if err != nil {
		fmt.Printf("Error retrieving GitHub API key: %v\n", err)
	} else {
		fmt.Printf("GitHub API Key: %s\n", githubKey)
	}

	model, err := cfg.GetModel()
	if err != nil {
		fmt.Printf("Error retrieving model: %v\n", err)
	} else {
		fmt.Printf("Model: %s\n", model)
	}

}

func help() {
	fmt.Println("Usage:")
	fmt.Println("  aidy pr   - Generate a pull request using AI-generated title and body.")
	fmt.Println("  aidy help - Show this help message.")
}

func Issue(task string, brain ai.AI, gh github.Github, ch cache.AidyCache, out output.Output) {
	summary, _ := ch.Summary()
	title, err := brain.IssueTitle(task, summary)
	if err != nil {
		log.Fatalf("Error generating title: %v", err)
	}
	body, err := brain.IssueBody(task, summary)
	if err != nil {
		log.Fatalf("Error generating body: %v", err)
	}
	labels := gh.Labels()
	suitable, err := brain.IssueLabels(body, labels)
	if err != nil {
		log.Fatalf("Error generating labels: %v", err)
	}
	remote := ch.Remote()
	var repo string
	if remote != "" {
		repo = " --repo " + remote
	} else {
		repo = ""
	}
	var cmd string
	if len(suitable) > 0 {
		cmd = fmt.Sprintf("\n%s", escapeBackticks(fmt.Sprintf("gh issue create --title \"%s\" --body \"%s\" --label \"%s\"", healQuotes(title), healQuotes(body), strings.Join(suitable, ","))))
	} else {
		cmd = fmt.Sprintf("\n%s", escapeBackticks(fmt.Sprintf("gh issue create --title \"%s\" --body \"%s\"", healQuotes(title), healQuotes(body))))
	}
	cmd = fmt.Sprintf("%s%s\n", cmd, repo)
	err = out.Print(cmd)
	if err != nil {
		log.Fatalf("Error during making an issue: %v", err)
	}
}

func Squash(gs git.Git, brain ai.AI) {
	base, err := gs.BaseBranch()
	if err != nil {
		log.Fatalf("Error determining base branch: %v", err)
	}
	resetErr := gs.Reset("refs/heads/" + base)
	if resetErr != nil {
		log.Fatalf("Error executing git reset: %v", err)
	}
	Commit(gs, brain)
}

func Commit(gs git.Git, brain ai.AI) {
	branch, err := gs.CurrentBranch()
	if err != nil {
		log.Fatalf("Error getting branch name: %v", err)
	}
	_, err = gs.Run("add", "--all")
	if err != nil {
		log.Fatalf("Error adding changes: %v", err)
	}
	diff, diffErr := gs.CurrentDiff()
	if diffErr != nil {
		log.Fatalf("Error getting diff: %v", err)
	}
	nissue := inumber(branch)
	msg, cerr := brain.CommitMessage(nissue, diff)
	if cerr != nil {
		log.Fatalf("Error generating commit message: %v", cerr)
	}
	_, err = gs.Run("commit", "-m", msg)
	if err != nil {
		log.Fatalf("Error committing changes: %v", err)
	}
	Heal(gs)
}

func PullRequest(gs git.Git, brain ai.AI, gh github.Github, ch cache.AidyCache, out output.Output) {
	branch, err := gs.CurrentBranch()
	if err != nil {
		log.Fatalf("Error getting branch name: %v", err)
	}
	diff, err := gs.Diff()
	if err != nil {
		log.Fatalf("Error getting git diff: %v", err)
	}
	summary, _ := ch.Summary()
	nissue := inumber(branch)
	issue := gh.Description(nissue)
	title, err := brain.PrTitle(nissue, diff, issue, summary)
	if err != nil {
		log.Fatalf("Error generating title: %v", err)
	}
	body, err := brain.PrBody(nissue, diff, issue, summary)
	if err != nil {
		log.Fatalf("Error generating body: %v", err)
	}
	remote := ch.Remote()
	var repo string
	if remote != "" {
		repo = " --repo " + remote
	} else {
		repo = ""
	}
	prtitle := healPRTitle(healQuotes(title), nissue)
	prbody := healPRBody(healQuotes(body), nissue)
	cmd := escapeBackticks(fmt.Sprintf("gh pr create --title \"%s\" --body \"%s\"%s", prtitle, prbody, repo))
	err = out.Print(cmd)
	if err != nil {
		log.Fatalf("Error during making a pull-request: %v", err)
	}
}

func Heal(gitService git.Git) {
	name, err := gitService.CurrentBranch()
	if err != nil {
		log.Fatalf("Error getting branch name: %v", err)
	}
	issueNumber := inumber(name)
	commitMessage, gitErr := gitService.CommitMessage()
	if gitErr != nil {
		log.Fatalf("Error getting current commit message: %v", err)
	}
	re := regexp.MustCompile(`#\d+`)
	newCommitMessage := re.ReplaceAllString(commitMessage, fmt.Sprintf("#%s", issueNumber))
	err = gitService.Amend(newCommitMessage)
	if err != nil {
		log.Fatalf("Error amending commit message: %v", err)
	}
}

func healQuotes(text string) string {
	clean := healQuote('`', text)
	clean = healQuote('\'', clean)
	clean = healQuote('"', clean)
	return clean
}

func healQuote(open rune, text string) string {
	stack := []int{}
	runes := []rune(text)
	size := len(runes)
	for i := range size {
		if runes[i] == open {
			if len(stack) > 0 {
				prev := stack[len(stack)-1]
				if i == size-1 && prev == 0 {
					return string(runes[1 : size-1])
				} else {
					stack = stack[:len(stack)-1]
				}
			} else {
				stack = append(stack, i)
			}
		}
	}
	return text
}

func healPRBody(body string, issue string) string {
	re := regexp.MustCompile(`(Closes|closes|related to|Related to)\s?:?\s?#(\d+)`)
	replaced := re.ReplaceAllStringFunc(body, func(m string) string {
		groups := re.FindStringSubmatch(m)
		if len(groups) == 3 {
			verb := groups[1]
			return fmt.Sprintf("%s #%s", verb, issue)
		}
		return m
	})
	return replaced
}

func healPRTitle(text string, issue string) string {
	re := regexp.MustCompile(`(fix|feat|build|chore|ci|docs|style|refactor|perf|test)\(#(\d+)\)`)
	replaced := re.ReplaceAllStringFunc(text, func(m string) string {
		groups := re.FindStringSubmatch(m)
		if len(groups) == 3 {
			commitType := groups[1]
			return fmt.Sprintf("%s(#%s)", commitType, issue)
		}
		return m
	})
	return replaced
}

func AppendCommit(gs git.Git) {
	err := gs.Append()
	if err != nil {
		log.Fatalf("Error appending to commit: %v", err)
	}
}

func inumber(branch string) string {
	if branch == "" {
		return "unknown"
	}
	if strings.Contains(branch, "/") {
		parts := strings.Split(branch, "/")
		branch = parts[len(parts)-1]
	}
	re, err := regexp.Compile(`\d+`)
	if err != nil {
		panic(err)
	}
	found := re.FindString(branch)
	if found == "" {
		return branch
	}
	return found
}

func escapeBackticks(input string) string {
	return strings.ReplaceAll(input, "`", "\\`")
}

func Clean() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Can't understand what is the current directory, '%v'", err)
	}
	cache := filepath.Join(dir, ".aidy")
	err = os.RemoveAll(cache)
	if err != nil {
		log.Fatalf("Can't clear '.aidy' directory, '%v'", err)
	}
}

func StartIssue(number string, brain ai.AI, gs git.Git, gh github.Github) error {
	if number == "" {
		return fmt.Errorf("error: no issue number provided")
	}
	re := regexp.MustCompile(`\d+`)
	found := re.FindString(number)
	if found == "" {
		return fmt.Errorf("error: invalid issue number '%s'", number)
	}
	descr := gh.Description(found)
	raw, err := brain.SuggestBranch(descr)
	if err != nil {
		return fmt.Errorf("error generating branch name: %v", err)
	}
	branch := branchName(found, raw)
	err = gs.Checkout(branch)
	if err != nil {
		return fmt.Errorf("error checking out branch '%s': %v", branch, err)
	}
	return nil
}

func branchName(number string, suggested string) string {
	suggested = strings.ReplaceAll(suggested, " ", "-")
	suggested = strings.ReplaceAll(suggested, "_", "-")
	suggested = strings.ReplaceAll(suggested, "/", "-")
	return fmt.Sprintf("%s-%s", number, suggested)
}

func InitSummary(required bool, aiService ai.AI, ch cache.AidyCache) {
	if required {
		log.Println("Undertstanding the project summary")
		content, err := os.ReadFile("README.md")
		if err != nil {
			log.Printf("Can't retrieve content of README.md, because of '%v'", err)
			return
		}
		readme := string(content)
		hash := fnv.New64a()
		hash.Write([]byte(readme))
		shash := fmt.Sprintf("%x", hash.Sum64())
		_, chash := ch.Summary()
		if shash != chash {
			summary, err := aiService.Summary(readme)
			if err != nil {
				log.Printf("Can't generate summary for README.md, because of '%v'", err)
				return
			}
			ch.WithSummary(summary, shash)
			log.Printf("Project '%s' summary was successfully saved\n", shash)
		} else {
			log.Printf("No need to update the project summary '%s'\n", shash)
		}
	}
}

func CheckGitInstalled(gitService git.Git) {
	installed, err := gitService.Installed()
	if err != nil {
		log.Fatalf("Can't understand whether git is installed or not, because of '%v'", err)
	}
	if !installed {
		log.Fatal("git is not installed on the system")
	}
}
