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
	git     git.Git
	github  github.Github
	ai      ai.AI
	editor  output.Output
	config  config.Config
	cache   cache.AidyCache
	printer output.Output
	in      *os.File
}

// Create a real aidy instance
// This function initializes the aidy instance with the provided parameters.
// Parameters:
// - summary: whether to use a project summary in AI requests
// - aider: whether to use the aider configuration
// - ailess: whether to use AI or not
func NewAidy(summary bool, aider bool, ailess bool) Aidy {
	var aidy real
	aidy.in = os.Stdin
	shell := executor.NewReal()
	aidy.editor = output.NewEditor(shell)
	aidy.printer = output.NewPrinter()
	var err error
	aidy.git, err = git.NewGit(shell)
	if err != nil {
		log.Fatalf("failed to initialize git: %v", err)
	}
	err = aidy.CheckGitInstalled()
	if err != nil {
		log.Fatalf("failed to check git installation: %v", err)
	}
	aidy.cache = newcache(aidy.git)
	conf := newconf(aider, aidy.git)
	if aidy.ai, err = Brain(ailess, summary, conf); err != nil {
		log.Fatalf("failed to initialize AI: %v", err)
	}
	aidy.github = newgithub(aidy.git, conf, aidy.cache)
	aidy.config = conf
	if err = aidy.InitSummary(summary, "README.md"); err != nil {
		log.Printf("warning: failed to initialize project summary: %v", err)
	}
	if err = aidy.SetTarget(); err != nil {
		log.Fatalf("failed to set target repository: %v", err)
	}
	return &aidy
}

func (r real) SetTarget() error {
	target := r.cache.Remote()
	if target != "" {
		r.print(fmt.Sprintf("target repository is set to: %s\n", target))
	} else {
		r.print("can't find target")
		out, err := r.git.Run("remote", "-v")
		if err != nil {
			return fmt.Errorf("error running git remote command: %v", err)
		}
		lines := strings.Split(out, "\n")
		r.print(fmt.Sprintf("found %s remote repositories:\n", out))
		re := regexp.MustCompile(`(?:git@github\.com:|https://github\.com/)([^/]+/.+?)(?:\.git)?$`)
		unique := make(map[string]struct{})
		for _, line := range lines {
			matches := re.FindStringSubmatch(line)
			if len(matches) == 2 {
				unique[string(matches[1])] = struct{}{}
			}
		}
		var repos []string
		for repo := range unique {
			repos = append(repos, repo)
		}
		sort.Strings(repos)
		if len(repos) == 1 {
			r.print(fmt.Sprintf("found one remote repository: %s\n", repos[0]))
			r.cache.WithRemote(repos[0])
		} else if len(repos) < 1 {
			return fmt.Errorf("no remote repositories found, please set one")
		} else {
			r.print("where are you going to send prs and issues: ")
			for i, repo := range repos {
				r.print(fmt.Sprintf("(%d): %s\n", i+1, repo))
			}
			var choice int
			r.print("enter the number of the repository to use: ")
			_, err := fmt.Fscan(r.in, &choice)
			if err != nil || choice < 1 || choice > len(repos) {
				return fmt.Errorf("invalid choice: %v", err)
			}
			repo := repos[choice-1]
			r.cache.WithRemote(repo)
		}
	}
	return nil
}

func (r *real) Diff() error {
	diff, err := r.git.Diff()
	if err != nil {
		return fmt.Errorf("failed to get diff: '%v'", err)
	} else {
		r.print(fmt.Sprintf("Diff with the base branch:\n%s\n", diff))
		return nil
	}
}

func (r *real) Commit() error {
	branch, err := r.git.CurrentBranch()
	if err != nil {
		return fmt.Errorf("error getting branch name: %v", err)
	}
	_, err = r.git.Run("add", "--all")
	if err != nil {
		return fmt.Errorf("error adding changes: %v", err)
	}
	diff, diffErr := r.git.CurrentDiff()
	if diffErr != nil {
		return fmt.Errorf("error getting current diff: %v", diffErr)
	}
	nissue := inumber(branch)
	msg, cerr := r.ai.CommitMessage(nissue, diff)
	if cerr != nil {
		return fmt.Errorf("error generating commit message: %v", cerr)
	}
	_, err = r.git.Run("commit", "-m", msg)
	if err != nil {
		return fmt.Errorf("error committing changes: %v", err)
	}
	return r.Heal()
}

func (r *real) Issue(task string) error {
	summary, _ := r.cache.Summary()
	title, err := r.ai.IssueTitle(task, summary)
	if err != nil {
		return fmt.Errorf("error generating title: %v", err)
	}
	body, err := r.ai.IssueBody(task, summary)
	if err != nil {
		return fmt.Errorf("error generating body: %v", err)
	}
	labels, err := r.github.Labels()
	if err != nil {
		return fmt.Errorf("error retrieving labels: %v", err)
	}
	suitable, err := r.ai.IssueLabels(body, labels)
	if err != nil {
		return fmt.Errorf("error generating suitable labels: %v", err)
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
	return r.editor.Print(cmd)
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

func (r *real) PrintConfig() error {
	r.print("aidy configuration:")
	openai, err := r.config.OpenAiKey()
	if err != nil {
		r.print(fmt.Sprintf("error retrieving openai api key: %v", err))
	} else {
		r.print(fmt.Sprintf("openai api key: %s", mask(openai)))
	}
	deepseek, err := r.config.DeepseekKey()
	if err != nil {
		r.print(fmt.Sprintf("error retrieving deepseek api key: %v", err))
	} else {
		r.print(fmt.Sprintf("deepseek api key: %s", mask(deepseek)))
	}
	gh, err := r.config.GithubKey()
	if err != nil {
		r.print(fmt.Sprintf("error retrieving github api key: %v", err))
	} else {
		r.print(fmt.Sprintf("github api key: %s", mask(gh)))
	}
	model, err := r.config.Model()
	if err != nil {
		r.print(fmt.Sprintf("error retrieving model: %v\n", err))
	} else {
		r.print(fmt.Sprintf("model: %s\n", model))
	}
	return err
}

func mask(key string) string {
	if len(key) <= 4 {
		return strings.Repeat("*", len(key))
	}
	return strings.Repeat("*", len(key)-4) + key[len(key)-4:]
}

func (r *real) print(msg string) {
	err := r.printer.Print(msg)
	if err != nil {
		log.Fatalf("Error printing message: %v", err)
	}
}

func (r *real) PullRequest() error {
	branch, err := r.git.CurrentBranch()
	if err != nil {
		return fmt.Errorf("error getting branch name: %v", err)
	}
	diff, err := r.git.Diff()
	if err != nil {
		return fmt.Errorf("error getting git diff: %v", err)
	}
	summary, _ := r.cache.Summary()
	nissue := inumber(branch)
	issue, err := r.github.Description(nissue)
	if err != nil {
		issue = "not-found"
		log.Printf("warning: issue description not found for issue #%s, using default value", nissue)
	}
	title, err := r.ai.PrTitle(nissue, diff, issue, summary)
	if err != nil {
		return fmt.Errorf("error generating pull request title: %v", err)
	}
	body, err := r.ai.PrBody(nissue, diff, issue, summary)
	if err != nil {
		return fmt.Errorf("error generating pull request body: %v", err)
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
	return r.editor.Print(cmd)
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
	descr, err := r.github.Description(found)
	if err != nil {
		return fmt.Errorf("error retrieving issue description: %v", err)
	}
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
	return r.editor.Print(command)
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
	err = r.git.Reset("refs/heads/" + base)
	if err != nil {
		log.Fatalf("Error executing git reset: %v", err)
	}
	err = r.Commit()
	if err != nil {
		log.Fatalf("Error committing changes: %v", err)
	}
}

func (r *real) Heal() error {
	name, err := r.git.CurrentBranch()
	if err != nil {
		return fmt.Errorf("error getting branch name: %v", err)
	}
	inumber := inumber(name)
	message, gitErr := r.git.CommitMessage()
	if gitErr != nil {
		return fmt.Errorf("error getting current commit message: %v", gitErr)
	}
	re := regexp.MustCompile(`#\d+`)
	updated := re.ReplaceAllString(message, fmt.Sprintf("#%s", inumber))
	err = r.git.Amend(updated)
	return err
}

func (r *real) CheckGitInstalled() error {
	installed, err := r.git.Installed()
	if err != nil {
		return fmt.Errorf("can't determine whether git is installed or not, because of '%v'", err)
	}
	if !installed {
		return fmt.Errorf("git is not installed on the system")
	}
	return nil
}

func (r *real) InitSummary(required bool, file string) error {
	if required {
		log.Println("Undertstanding the project summary")
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("can't read README.md file, because of '%v'", err)
		}
		readme := string(content)
		hash := fnv.New64a()
		hash.Write([]byte(readme))
		shash := fmt.Sprintf("%x", hash.Sum64())
		_, chash := r.cache.Summary()
		if shash != chash {
			summary, err := r.ai.Summary(readme)
			if err != nil {
				return fmt.Errorf("error generating summary for README.md: %v", err)
			}
			r.cache.WithSummary(summary, shash)
			r.print(fmt.Sprintf("project summary was successfully saved with hash '%s'\n", shash))
		} else {
			r.print(fmt.Sprintf("project summary is already saved with hash '%s'\n", shash))
		}
	}
	return nil
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

func Brain(ailess bool, summary bool, conf config.Config) (ai.AI, error) {
	if ailess {
		return ai.NewMockAI(), nil
	}
	provider, err := conf.Provider()
	if err != nil {
		return nil, fmt.Errorf("error getting AI provider from configuration: %v", err)
	}
	token, err := conf.Token()
	if err != nil {
		return nil, fmt.Errorf("error getting AI token: %v", err)
	}
	if token == "" {
		return nil, fmt.Errorf("AI token not found in configuration")
	}
	model, err := conf.Model()
	if err != nil {
		return nil, fmt.Errorf("error getting AI model from configuration: %v", err)
	}
	var brain ai.AI
	switch provider {
	case "deepseek":
		brain = ai.NewDeepSeek(token, summary)
	case "openai":
		brain = ai.NewOpenAI(token, model, 0.2, summary)
	default:
		return nil, fmt.Errorf("unknown AI provider '%s' specified in configuration", provider)
	}
	return brain, nil
}

func newcache(repo git.Git) cache.AidyCache {
	gitcache, err := cache.NewGitCache(".aidy/cache.js", repo)
	if err != nil {
		log.Fatalf("Can't open cache %v", err)
	}
	return cache.NewAidyCache(gitcache)
}
