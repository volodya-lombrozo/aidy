package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/volodya-lombrozo/aidy/ai"
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

	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting home directory: %v", err)
	}
	configPath := fmt.Sprintf("%s/.aidy.conf.yml", homeDir)
	yamlConfig := config.NewConf(configPath)
	apiKey, err := yamlConfig.GetOpenAIAPIKey()
	if err != nil {
		log.Fatalf("Error getting OpenAI API key: %v", err)
	}
	if apiKey == "" {
		log.Fatalf("OpenAI API key not found in config file")
	}
	githubKey, err := yamlConfig.GetGithubAPIKey()
	if err != nil {
		log.Printf("Can't find GitHub token in configuration")
		githubKey = ""
	}
    model, err := yamlConfig.GetModel()
	if err != nil {
		log.Fatalf("Can't find GitHub token in configuration")
	}
	shell := &executor.RealExecutor{}
	aiService := ai.NewOpenAI(apiKey, model, 0.2)
	gitService := git.NewRealGit(shell)
	gh := github.NewRealGithub("https://api.github.com", gitService, githubKey)
	switch command {
	case "help":
		help()
	case "pr", "pull-request":
		pull_request(gitService, aiService, gh)
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
		issue(userInput, aiService)
	case "conf", "config":
		printConfig(yamlConfig)
    default:
		log.Fatalf("Error: Unknown command '%s'. Use 'aidy help' for usage.\n", command)
	}
}

func printConfig(cfg config.Config) {
	fmt.Println("Current Configuration:")
	apiKey, err := cfg.GetOpenAIAPIKey()
	if err != nil {
		fmt.Printf("Error retrieving OpenAI API key: %v\n", err)
	} else {
		fmt.Printf("OpenAI API Key: %s\n", apiKey)
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
func issue(userInput string, aiService ai.AI) {
	title, err := aiService.GenerateIssueTitle(userInput)
	if err != nil {
		log.Fatalf("Error generating title: %v", err)
	}
	body, err := aiService.GenerateIssueBody(userInput)
	if err != nil {
		log.Fatalf("Error generating body: %v", err)
	}
	fmt.Printf("\n%s\n", escapeBackticks(fmt.Sprintf("gh issue create --title \"%s\" --body \"%s\"", title, body)))
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

func pull_request(gitService git.Git, aiService ai.AI, gh github.Github) {
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
	fmt.Printf("\n%s\n", escapeBackticks(fmt.Sprintf("gh pr create --title \"%s\" --body \"%s\"", title, body)))
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
