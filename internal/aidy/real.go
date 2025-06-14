package aidy

import (
	"fmt"
	"hash/fnv"
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
	"github.com/volodya-lombrozo/aidy/internal/log"
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
	logger  log.Logger
	in      *os.File
}

// Create a real aidy instance
// This function initializes the aidy instance with the provided parameters.
// Parameters:
// - summary: whether to use a project summary in AI requests
// - aider: whether to use the aider configuration
// - ailess: whether to use AI or not
// - silent: whether to suppress output
// - debug: whether to enable debug logging
func NewAidy(summary bool, aider bool, ailess bool, silent bool, debug bool) Aidy {
	var aidy real
	aidy.in = os.Stdin
	InitLogger(silent, debug)
	aidy.logger = log.Get()
	shell := executor.NewReal()
	aidy.editor = output.NewEditor(shell)
	aidy.printer = output.NewPrinter()
	var err error
	if aidy.git, err = git.NewGit(shell); err != nil {
		aidy.logger.Error("failed to initialize git: %v", err)
		os.Exit(1)
	}
	if err = aidy.CheckGitInstalled(); err != nil {
		aidy.logger.Error("git is not installed or not found: %v", err)
		os.Exit(1)
	}
	if aidy.cache, err = NewCache(aidy.git, ".aidy/cache.js"); err != nil {
		aidy.logger.Error("failed to initialize cache: %v", err)
		os.Exit(1)
	}
	if aidy.config, err = NewConf(aider, aidy.git); err != nil {
		aidy.logger.Error("failed to initialize configuration: %v", err)
		os.Exit(1)
	}
	if aidy.ai, err = Brain(ailess, summary, aidy.config); err != nil {
		aidy.logger.Error("failed to initialize AI: %v", err)
		os.Exit(1)
	}
	if aidy.github, err = NewGitHub(aidy.git, aidy.config, aidy.cache); err != nil {
		aidy.logger.Error("failed to initialize GitHub client: %v", err)
		os.Exit(1)
	}
	if err = aidy.InitSummary(summary, "README.md"); err != nil {
		aidy.logger.Warn("failed to initialize project summary: %v", err)
	}
	if err = aidy.SetTarget(); err != nil {
		aidy.logger.Error("failed to set target repository: %v", err)
		os.Exit(1)
	}
	return &aidy
}

func InitLogger(silent bool, debug bool) {
	var logger log.Logger
	if debug {
		logger = log.NewShort(log.NewZerolog(os.Stdout, "debug"))
	} else {
		logger = log.NewShort(log.NewZerolog(os.Stdout, "info"))
	}
	if silent {
		logger = log.NewSilent()
	}
	log.Set(logger)
}

func (r *real) SetTarget() error {
	target := r.cache.Remote()
	if target != "" {
		r.logger.Debug("target repository is set to: %s", target)
	} else {
		r.logger.Warn("target repository is not set, trying to find one")
		out, err := r.git.Run("remote", "-v")
		if err != nil {
			return fmt.Errorf("error running git remote command: %v", err)
		}
		lines := strings.Split(out, "\n")
		r.logger.Debug("found %d remote repositories:\n%s", len(lines), out)
		re := regexp.MustCompile(`(?:git@github\.com:|https://github\.com/)([^/]+/.+?)(?:\.git)?$`)
		unique := make(map[string]struct{})
		for _, line := range lines {
			fields := strings.Fields(line)
			var address string
			if len(fields) < 2 {
				address = line
			} else {
				address = fields[1]
			}
			matches := re.FindStringSubmatch(address)
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
			r.logger.Debug("found one remote repository: %s", repos[0])
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
		r.print(fmt.Sprintf("diff with the base branch:\n%s\n", diff))
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
	r.logger.Info("generating commit message...")
	msg, cerr := r.ai.CommitMessage(nissue, diff)
	if cerr != nil {
		return fmt.Errorf("error generating commit message: %v", cerr)
	}
	_, err = r.git.Run("commit", "-m", msg)
	if err != nil {
		return fmt.Errorf("error committing changes: %v", err)
	}
	r.logger.Info("commit was created with message: '%s'", msg)
	return r.Heal()
}

func (r *real) Issue(task string) error {
	summary, _ := r.cache.Summary()
	r.logger.Info("generating issue title...")
	title, err := r.ai.IssueTitle(task, summary)
	if err != nil {
		return fmt.Errorf("error generating title: %v", err)
	}
	r.logger.Info("generating issue body...")
	body, err := r.ai.IssueBody(task, summary)
	if err != nil {
		return fmt.Errorf("error generating body: %v", err)
	}
	r.logger.Info("retrieving suitable labels...")
	labels, err := r.github.Labels()
	if err != nil {
		return fmt.Errorf("error retrieving labels: %v", err)
	}
	r.logger.Info("applying suitable labels for the issue...")
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
	provider, err := r.config.Provider()
	if err != nil {
		r.print(fmt.Sprintf("error retrieving AI provider: %v", err))
	} else {
		r.print(fmt.Sprintf("AI provider: %s", provider))
	}
	model, err := r.config.Model()
	if err != nil {
		r.print(fmt.Sprintf("error retrieving model: %v", err))
	} else {
		r.print(fmt.Sprintf("model: %s", model))
	}
	token, err := r.config.Token()
	if err != nil {
		r.print(fmt.Sprintf("error retrieving AI token: %v", err))
	} else {
		r.print(fmt.Sprintf("AI API token: %s", mask(token)))
	}
	gh, err := r.config.GithubKey()
	if err != nil {
		r.print(fmt.Sprintf("error retrieving GitHub API key: %v", err))
	} else {
		r.print(fmt.Sprintf("GitHub API key: %s", mask(gh)))
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
		r.logger.Error("error printing message: %v", err)
		os.Exit(1)
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
	r.logger.Info("retrieving the description for issue #%s...", nissue)
	issue, err := r.github.Description(nissue)
	if err != nil {
		issue = "not-found"
		r.logger.Warn("issue description not found for issue #%s because of %v, using default value", nissue, err)
	}
	r.logger.Info("generating pull request title...")
	title, err := r.ai.PrTitle(nissue, diff, issue, summary)
	if err != nil {
		return fmt.Errorf("error generating pull request title: %v", err)
	}
	r.logger.Info("generating pull request body...")
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
		r.logger.Error("error appending to commit: %v", err)
		os.Exit(1)
	}
	r.logger.Info("changes were appended to the last commit")
}

func (r *real) Clean() {
	dir, err := os.Getwd()
	if err != nil {
		r.logger.Error("Can't understand what is the current directory, '%v'", err)
		os.Exit(1)
	}
	cache := filepath.Join(dir, ".aidy")
	err = os.RemoveAll(cache)
	if err != nil {
		r.logger.Error("Can't clear '.aidy' directory, '%v'", err)
		os.Exit(1)
	}
	r.logger.Info("'.aidy' directory was cleared")
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
	r.logger.Info("retrieving the description for issue #%s...", found)
	descr, err := r.github.Description(found)
	if err != nil {
		return fmt.Errorf("error retrieving issue description: %v", err)
	}
	r.logger.Info("generating branch name for issue #%s...", found)
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
	r.logger.Debug("found %d tags: %v", len(tags), tags)
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
		r.logger.Info("generating release notes...")
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
		r.logger.Info("latest tag is '%s', updating to '%s'", latest, updated)
	} else {
		updated = "v0.0.1"
		messages, err := r.git.Log("")
		if err != nil {
			return fmt.Errorf("failed to get git log: '%v'", err)
		}
		summary := strings.Join(messages, "\n")
		r.logger.Info("generating release notes for the first release...")
		notes, err = r.ai.ReleaseNotes(summary)
		if err != nil {
			return fmt.Errorf("failed to generate release notes: '%v'", err)
		}
		r.logger.Info("no tags found, creating the first release with version '%s'", updated)
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
		r.logger.Error("Error determining base branch: %v", err)
		os.Exit(1)
	}
	err = r.git.Reset("refs/heads/" + base)
	if err != nil {
		r.logger.Error("Error executing git reset: %v", err)
		os.Exit(1)
	}
	err = r.Commit()
	if err != nil {
		r.logger.Error("Error committing changes: %v", err)
		os.Exit(1)
	}
	r.logger.Info("changes were squashed")
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
		r.logger.Debug("summary is required, checking README.md file")
		r.logger.Info("understanding the project summary")
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
			r.logger.Info("summary was generated and saved to the cache with hash '%s', using it", shash)
		} else {
			r.logger.Info("summary is already saved in the cache with hash '%s', using it", shash)
		}
	}
	return nil
}

func NewGitHub(git git.Git, conf config.Config, cache cache.AidyCache) (github.Github, error) {
	token, err := conf.GithubKey()
	if err != nil {
		return nil, fmt.Errorf("error getting github token from configuration: %v", err)
	}
	return github.NewGithub("https://api.github.com", git, token, cache), nil
}

func NewConf(aider bool, git git.Git) (config.Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting home directory: %v", err)
	}
	var conf config.Config
	if aider {
		conf, _ = config.NewAider(fmt.Sprintf("%s/.aider.conf.yml", home))
	} else {
		conf, _ = config.NewCascade(git)
	}
	return conf, nil
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

func NewCache(repo git.Git, path string) (cache.AidyCache, error) {
	ch, err := cache.NewGitCache(path, repo)
	if err != nil {
		return nil, fmt.Errorf("can't open cache: %v", err)
	}
	return cache.NewAidyCache(ch), nil
}
