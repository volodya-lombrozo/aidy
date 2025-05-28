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
	"github.com/volodya-lombrozo/aidy/internal/executor"
	"github.com/volodya-lombrozo/aidy/internal/git"
	"github.com/volodya-lombrozo/aidy/internal/github"
	"github.com/volodya-lombrozo/aidy/internal/output"
	"golang.org/x/mod/semver"
)

type real struct {
	git    git.Git
	github github.Github
	ai     ai.AI
	output output.Output
	config config.Config
	cache  cache.AidyCache
}

// Create a real aidy instance
// This function initializes the aidy instance with the provided parameters.
// Parameters:
// - summary: whether to use a project summary in AI requests
// - aider: whether to use the aider configuration
// - ailess: whether to use AI or not
func NewAidy(summary bool, aider bool, ailess bool) Aidy {
	var aidy real
	shell := executor.NewReal()
	out := output.NewEditor(shell)
	aidy.output = out
	var err error
	aidy.git, err = git.NewGit(shell)
	if err != nil {
		log.Fatalf("Failed to initialize git: %v", err)
	}
	aidy.CheckGitInstalled()
	aidy.cache = newcache(aidy.git)
	conf := newconf(aider, aidy.git)
	aidy.ai = brain(ailess, summary, conf)
	aidy.github = newgithub(aidy.git, conf, aidy.cache)
	aidy.InitSummary(summary)
	return &aidy
}

func (r *real) Diff() {
	diff, err := r.git.Diff()
	if err != nil {
		log.Fatalf("Failed to get diff: '%v'", err)
	} else {
		fmt.Printf("Diff with the base branch:\n%s\n", diff)
	}
}
func (r *real) Commit() {
	branch, err := r.git.CurrentBranch()
	if err != nil {
		log.Fatalf("Error getting branch name: %v", err)
	}
	_, err = r.git.Run("add", "--all")
	if err != nil {
		log.Fatalf("Error adding changes: %v", err)
	}
	diff, diffErr := r.git.CurrentDiff()
	if diffErr != nil {
		log.Fatalf("Error getting diff: %v", err)
	}
	nissue := inumber(branch)
	msg, cerr := r.ai.CommitMessage(nissue, diff)
	if cerr != nil {
		log.Fatalf("Error generating commit message: %v", cerr)
	}
	_, err = r.git.Run("commit", "-m", msg)
	if err != nil {
		log.Fatalf("Error committing changes: %v", err)
	}
	r.Heal()
}

func (r *real) Issue(task string) {
	summary, _ := r.cache.Summary()
	title, err := r.ai.IssueTitle(task, summary)
	if err != nil {
		log.Fatalf("Error generating title: %v", err)
	}
	body, err := r.ai.IssueBody(task, summary)
	if err != nil {
		log.Fatalf("Error generating body: %v", err)
	}
	labels := r.github.Labels()
	suitable, err := r.ai.IssueLabels(body, labels)
	if err != nil {
		log.Fatalf("Error generating labels: %v", err)
	}
	remote := r.cache.Remote()
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
	err = r.output.Print(cmd)
	if err != nil {
		log.Fatalf("Error during making an issue: %v", err)
	}
}

func escapeBackticks(input string) string {
	return strings.ReplaceAll(input, "`", "\\`")
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

func (r *real) PrintConfig() {
	fmt.Println("Current Configuration:")
	openai, err := r.config.OpenAiKey()
	if err != nil {
		fmt.Printf("Error retrieving OpenAI API key: %v\n", err)
	} else {
		fmt.Printf("OpenAI API Key: %s\n", openai)
	}
	deepseek, err := r.config.DeepseekKey()
	if err != nil {
		fmt.Printf("Error retrieving deepseek API key: %v\n", err)
	} else {
		fmt.Printf("Deepseek API Key: %s\n", deepseek)
	}
	githubKey, err := r.config.GithubKey()
	if err != nil {
		fmt.Printf("Error retrieving GitHub API key: %v\n", err)
	} else {
		fmt.Printf("GitHub API Key: %s\n", githubKey)
	}
	model, err := r.config.Model()
	if err != nil {
		fmt.Printf("Error retrieving model: %v\n", err)
	} else {
		fmt.Printf("Model: %s\n", model)
	}
}

func (r *real) PullRequest() {
	branch, err := r.git.CurrentBranch()
	if err != nil {
		log.Fatalf("Error getting branch name: %v", err)
	}
	diff, err := r.git.Diff()
	if err != nil {
		log.Fatalf("Error getting git diff: %v", err)
	}
	summary, _ := r.cache.Summary()
	nissue := inumber(branch)
	issue := r.github.Description(nissue)
	title, err := r.ai.PrTitle(nissue, diff, issue, summary)
	if err != nil {
		log.Fatalf("Error generating title: %v", err)
	}
	body, err := r.ai.PrBody(nissue, diff, issue, summary)
	if err != nil {
		log.Fatalf("Error generating body: %v", err)
	}
	remote := r.cache.Remote()
	var repo string
	if remote != "" {
		repo = " --repo " + remote
	} else {
		repo = ""
	}
	prtitle := healPRTitle(healQuotes(title), nissue)
	prbody := healPRBody(healQuotes(body), nissue)
	cmd := escapeBackticks(fmt.Sprintf("gh pr create --title \"%s\" --body \"%s\"%s", prtitle, prbody, repo))
	err = r.output.Print(cmd)
	if err != nil {
		log.Fatalf("Error during making a pull-request: %v", err)
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

func (r *real) Append() {
	err := r.git.Append()
	if err != nil {
		log.Fatalf("Error appending to commit: %v", err)
	}
}

func (r *real) Clean() {
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

func (r *real) StartIssue(number string) error {
	if number == "" {
		return fmt.Errorf("error: no issue number provided")
	}
	re := regexp.MustCompile(`\d+`)
	found := re.FindString(number)
	if found == "" {
		return fmt.Errorf("error: invalid issue number '%s'", number)
	}
	descr := r.github.Description(found)
	raw, err := r.ai.SuggestBranch(descr)
	if err != nil {
		return fmt.Errorf("error generating branch name: %v", err)
	}
	branch := branchName(found, raw)
	err = r.git.Checkout(branch)
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

func (r *real) Release(interval string, repo string) error {
	tags, err := r.git.Tags(repo)
	if err != nil {
		return fmt.Errorf("failed to get tags: '%v'", err)
	}
	log.Printf("tags found: %v, size %d\n", tags, len(tags))
	var notes string
	var updated string
	if len(tags) > 0 {
		mtags := clearTags(tags)
		latest := latest(keys(mtags))
		messages, err := r.git.Log(mtags[latest])
		if err != nil {
			return fmt.Errorf("failed to get git log: '%v'", err)
		}
		summary := strings.Join(messages, "\n")
		notes, err = r.ai.ReleaseNotes(summary)
		if err != nil {
			return fmt.Errorf("failed to generate release notes: '%v'", err)
		}
		updated, err = upver(latest, interval)
		if err != nil {
			return fmt.Errorf("failed to update version: '%v'", err)
		}
		if strings.HasPrefix(mtags[latest], "v") {
			updated = "v" + updated
		}
	} else {
		updated = "v0.0.1"
		messages, err := r.git.Log("")
		if err != nil {
			return fmt.Errorf("failed to get git log: '%v'", err)
		}
		summary := strings.Join(messages, "\n")
		notes, err = r.ai.ReleaseNotes(summary)
		if err != nil {
			return fmt.Errorf("failed to generate release notes: '%v'", err)
		}
	}
	command := fmt.Sprintf("git tag --cleanup=verbatim -a \"%s\" -m \"%s\" ", updated, notes)
	return r.output.Print(command)
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

func (r *real) Squash() {
	base, err := r.git.BaseBranch()
	if err != nil {
		log.Fatalf("Error determining base branch: %v", err)
	}
	resetErr := r.git.Reset("refs/heads/" + base)
	if resetErr != nil {
		log.Fatalf("Error executing git reset: %v", err)
	}
	r.Commit()
}

func (r *real) Heal() {
	name, err := r.git.CurrentBranch()
	if err != nil {
		log.Fatalf("Error getting branch name: %v", err)
	}
	issueNumber := inumber(name)
	commitMessage, gitErr := r.git.CommitMessage()
	if gitErr != nil {
		log.Fatalf("Error getting current commit message: %v", err)
	}
	re := regexp.MustCompile(`#\d+`)
	newCommitMessage := re.ReplaceAllString(commitMessage, fmt.Sprintf("#%s", issueNumber))
	err = r.git.Amend(newCommitMessage)
	if err != nil {
		log.Fatalf("Error amending commit message: %v", err)
	}
}

func (r *real) CheckGitInstalled() {
	installed, err := r.git.Installed()
	if err != nil {
		log.Fatalf("Can't understand whether git is installed or not, because of '%v'", err)
	}
	if !installed {
		log.Fatal("git is not installed on the system")
	}
}

func (r *real) InitSummary(required bool) {
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
		_, chash := r.cache.Summary()
		if shash != chash {
			summary, err := r.ai.Summary(readme)
			if err != nil {
				log.Printf("Can't generate summary for README.md, because of '%v'", err)
				return
			}
			r.cache.WithSummary(summary, shash)
			log.Printf("Project '%s' summary was successfully saved\n", shash)
		} else {
			log.Printf("No need to update the project summary '%s'\n", shash)
		}
	}
}

func newgithub(git git.Git, conf config.Config, cache cache.AidyCache) github.Github {
	token, err := conf.GithubKey()
	if err != nil {
		log.Fatalf("error getting github token: %v", err)
	}
	return github.NewGithub("https://api.github.com", git, token, cache)
}

func newconf(aider bool, git git.Git) config.Config {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("error getting home directory: %v", err)
	}
	var conf config.Config
	if aider {
		conf, _ = config.NewAider(fmt.Sprintf("%s/.aider.conf.yml", home))
	} else {
		conf, _ = config.NewCascade(git)
	}
	return conf
}

func brain(ailess bool, sumrequired bool, conf config.Config) ai.AI {
	if ailess {
		return ai.NewMockAI()
	}
	model, err := conf.Model()
	if err != nil {
		log.Fatalf("Can't find GitHub token in configuration")
	}
	var brain ai.AI
	if model == "deepseek-chat" {
		apiKey, err := conf.DeepseekKey()
		if err != nil {
			log.Fatalf("Error getting Deepseek API key: %v", err)
		}
		if apiKey == "" {
			log.Fatalf("Deepseek API key not found in config file")
		} else {
			log.Println("Deepseek key is found")
		}
		brain = ai.NewDeepSeekAI(apiKey, sumrequired)
	} else {
		apiKey, err := conf.OpenAiKey()
		if err != nil {
			log.Fatalf("Error getting OpenAI API key: %v", err)
		}
		if apiKey == "" {
			log.Fatalf("OpenAI API key not found in config file")
		}
		brain = ai.NewOpenAI(apiKey, model, 0.2, sumrequired)
	}
	return brain
}

func newcache(repo git.Git) cache.AidyCache {
	gitcache, err := cache.NewGitCache(".aidy/cache.js", repo)
	if err != nil {
		log.Fatalf("Can't open cache %v", err)
	}
	return cache.NewAidyCache(gitcache)
}
