package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/volodya-lombrozo/aidy/ai"
	"github.com/volodya-lombrozo/aidy/cache"
	"github.com/volodya-lombrozo/aidy/config"
	"github.com/volodya-lombrozo/aidy/executor"
	"github.com/volodya-lombrozo/aidy/git"
	"github.com/volodya-lombrozo/aidy/github"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Error: No command provided. Use 'aidy help' for usage.")
	}
	command := os.Args[1]
	shell := &executor.RealExecutor{}

	gitcache, err := cache.NewGitCache(".aidy/cache.js")
	if err != nil {
		log.Fatalf("Can't open cache %v", err)
	}
	ch := cache.NewAidyCache(gitcache)
	yamlConfig := readConfiguration()
	githubKey, err := yamlConfig.GetGithubAPIKey()
	if err != nil {
		log.Printf("Can't find GitHub token in configuration")
		githubKey = ""
	}
	model, err := yamlConfig.GetModel()
	if err != nil {
		log.Fatalf("Can't find GitHub token in configuration")
	}
	var aiService ai.AI
	if model == "deepseek-chat" {
		apiKey, err := yamlConfig.GetDeepseekAPIKey()
		if err != nil {
			log.Fatalf("Error getting Deepseek API key: %v", err)
		}
		if apiKey == "" {
			log.Fatalf("Deepseek API key not found in config file")
		} else {
			log.Println("Deepseek key is found")
		}
		aiService = ai.NewDeepSeekAI(apiKey)
	} else {
		apiKey, err := yamlConfig.GetOpenAIAPIKey()
		if err != nil {
			log.Fatalf("Error getting OpenAI API key: %v", err)
		}
		if apiKey == "" {
			log.Fatalf("OpenAI API key not found in config file")
		}
		aiService = ai.NewOpenAI(apiKey, model, 0.2)
	}
	gitService := git.NewRealGit(shell)
	gh := github.NewRealGithub("https://api.github.com", gitService, githubKey, ch)
	target := ch.Remote()
	if target != "" {
		log.Printf("Target repo is: %s\n", target)
	} else {
		log.Println("Can't find target")
		repos := gh.Remotes()
		if len(repos) == 1 {
			ch.WithRemote(repos[0])
		} else if len(repos) < 1 {
			log.Fatal("I can't find remore repositories :(")
		} else {
			fmt.Println("Where are you going to send PRs and Issues: ")
			for i, repo := range repos {
				fmt.Printf("(%d): %s\n", i+1, repo)
			}
			var choice int
			fmt.Print("Enter the number of the repository to use: ")
			_, err := fmt.Scan(&choice)
			if err != nil || choice < 1 || choice > len(repos) {
				panic("Invalid choice")
			}
			selectedRepo := repos[choice-1]
			ch.WithRemote(selectedRepo)
		}
	}
	switch command {
	case "help":
		help()
	case "pr", "pull-request":
		pull_request(gitService, aiService, gh, ch)
	case "h", "heal":
		heal(gitService, shell)
	case "ci", "commit":
		noAI := len(os.Args) > 2 && os.Args[2] == "-n"
		commit(gitService, shell, noAI, aiService)
	case "sq", "squash":
		squash(gitService, shell, aiService)
	case "ap", "append":
		appendToCommit(gitService)
	case "i", "issue":
		if len(os.Args) < 3 {
			log.Fatalf("Error: No input provided for issue generation.")
		}
		userInput := os.Args[2]
		issue(userInput, aiService, gh, ch)
	case "conf", "config":
		printConfig(yamlConfig)
	case "clean":
		cleanCache()
	default:
		log.Fatalf("Error: Unknown command '%s'. Use 'aidy help' for usage.\n", command)
	}
}

func readConfiguration() config.Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}
	var conf config.Config
	aider := false
	for _, arg := range os.Args {
		if arg == "--aider" {
			aider = true
		}
	}
	if aider {
		configPath := fmt.Sprintf("%s/.aider.conf.yml", homeDir)
		conf = config.NewAiderConf(configPath)
	} else {
		configPath := fmt.Sprintf("%s/.aidy.conf.yml", homeDir)
		conf = config.NewConf(configPath)
	}
	return conf
}

func printConfig(cfg config.Config) {
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

// This method implements the 'issue' command.
// It creates a `gh` issue command.
// For example `gh issue create --title "Issue title" --body "Issue body"`
func issue(userInput string, aiService ai.AI, gh github.Github, ch cache.AidyCache) {
	title, err := aiService.GenerateIssueTitle(userInput)
	if err != nil {
		log.Fatalf("Error generating title: %v", err)
	}
	body, err := aiService.GenerateIssueBody(userInput)
	if err != nil {
		log.Fatalf("Error generating body: %v", err)
	}
	labels := gh.Labels()
	suitable, err := aiService.GenerateIssueLabels(body, labels)
	if err != nil {
		log.Fatalf("Error generating labels: %v", err)
	}
	remote := ch.Remote()
	var repo string
	if remote != "" {
		repo = " --repo=" + remote
	} else {
		repo = ""
	}
	var cmd string
	if len(suitable) > 0 {
		cmd = fmt.Sprintf("\n%s", escapeBackticks(fmt.Sprintf("gh issue create --title \"%s\" --body \"%s\" --label \"%s\"", healQoutes(title), healQoutes(body), strings.Join(suitable, ","))))
	} else {
		cmd = fmt.Sprintf("\n%s", escapeBackticks(fmt.Sprintf("gh issue create --title \"%s\" --body \"%s\"", healQoutes(title), healQoutes(body))))
	}
	fmt.Printf("%s%s\n", cmd, repo)
}

func squash(gitService git.Git, shell executor.Executor, aiService ai.AI) {
	baseBranch, err := gitService.GetBaseBranchName()
	if err != nil {
		log.Fatalf("Error determining base branch: %v", err)
	}
	_, resetErr := shell.RunCommand("git", "reset", "--soft", baseBranch)
	if resetErr != nil {
		log.Fatalf("Error executing git reset: %v", err)
	}
	commit(gitService, shell, false, aiService)
}

// Comment
func commit(gitService git.Git, shell executor.Executor, noAI bool, aiService ai.AI) {
	if noAI {
		err := gitService.CommitChanges()
		if err != nil {
			log.Fatalf("Error committing changes: %v", err)
		}
	} else {
		branchName, err := gitService.GetBranchName()
		if err != nil {
			log.Fatalf("Error getting branch name: %v", err)
		}
		_, addErr := shell.RunCommand("git", "add", "--all")
		if addErr != nil {
			log.Fatalf("Error adding git files: %v", err)
		}
		diff, diffErr := gitService.GetCurrentDiff()
		if diffErr != nil {
			log.Fatalf("Error getting diff: %v", err)
		}
		msg, cerr := aiService.GenerateCommitMessage(branchName, diff)
		if cerr != nil {
			log.Fatalf("Error generating commit message: %v", cerr)
		}
		if err := gitService.CommitChanges(msg); err != nil {
			log.Fatalf("Error committing changes: %v", err)
		}
	}
	heal(gitService, shell)
}

func pull_request(gitService git.Git, aiService ai.AI, gh github.Github, ch cache.AidyCache) {
	branchName, err := gitService.GetBranchName()
	if err != nil {
		log.Fatalf("Error getting branch name: %v", err)
	}
	diff, err := gitService.GetDiff()
	if err != nil {
		log.Fatalf("Error getting git diff: %v", err)
	}
	issue := gh.IssueDescription(extractIssueNumber(branchName))
	title, err := aiService.GenerateTitle(branchName, diff, issue)
	if err != nil {
		log.Fatalf("Error generating title: %v", err)
	}
	body, err := aiService.GenerateBody(branchName, diff, issue)
	if err != nil {
		log.Fatalf("Error generating body: %v", err)
	}
	remote := ch.Remote()
	var repo string
	if remote != "" {
		repo = " --repo=" + remote
	} else {
		repo = ""
	}
	fmt.Printf("\n%s%s\n", escapeBackticks(fmt.Sprintf("gh pr create --title \"%s\" --body \"%s\"", healQoutes(title), healQoutes(body))), repo)
}

func heal(gitService git.Git, shell executor.Executor) {
	branchName, err := gitService.GetBranchName()
	if err != nil {
		log.Fatalf("Error getting branch name: %v", err)
	}
	issueNumber := extractIssueNumber(branchName)
	commitMessage, gitErr := gitService.GetCurrentCommitMessage()
	if gitErr != nil {
		log.Fatalf("Error getting current commit message: %v", err)
	}
	re := regexp.MustCompile(`#\d+`)
	newCommitMessage := re.ReplaceAllString(commitMessage, fmt.Sprintf("#%s", issueNumber))
	_, err = shell.RunCommand("git", "commit", "--amend", "-m", newCommitMessage)
	if err != nil {
		log.Fatalf("Error amending commit message: %v", err)
	}
}

func healQoutes(text string) string {
	return strings.Trim(text, `"'`)
}

func appendToCommit(gitService git.Git) {
	err := gitService.AppendToCommit()
	if err != nil {
		log.Fatalf("Error appending to commit: %v", err)
	}
}

func extractIssueNumber(branchName string) string {
	// Assuming the branch name format is "<issue-number>_<description>"
	parts := strings.Split(branchName, "_")
	if len(parts) > 0 && branchName != "" {
		return parts[0]
	}
	return "unknown"
}

func escapeBackticks(input string) string {
	return strings.ReplaceAll(input, "`", "\\`")
}

func cleanCache() {
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
