package main

import (
	"fmt"
	"github.com/volodya-lombrozo/aidy/ai"
	"github.com/volodya-lombrozo/aidy/config"
	"github.com/volodya-lombrozo/aidy/executor"
	"github.com/volodya-lombrozo/aidy/git"
	"log"
	"os"
	"regexp"
	"strings"
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
	yamlConfig, err := config.NewYAMLConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to create YAMLConfig: %v", err)
	}
	apiKey, confErr := yamlConfig.GetOpenAIAPIKey()
	if confErr != nil {
		panic(confErr)
	}
	if apiKey == "" {
		log.Fatalf("OpenAI API key not found in config file")
	}
	shell := &executor.RealExecutor{}
	aiService := ai.NewOpenAI(apiKey, "gpt-4o", 0.2)
	gitService := git.NewRealGit(shell)
	switch command {
	case "help":
		help()
	case "pr", "pull-request":
		pull_request(gitService, aiService)
	case "h", "heal":
		heal(gitService, shell)
	case "ci", "commit":
		commit(gitService, shell)
	case "sq", "squash":
		squash(gitService, shell)
	case "ap", "append":
		appendToCommit(gitService)
	case "i", "issue":
		if len(os.Args) < 3 {
			log.Fatalf("Error: No input provided for issue generation.")
		}
		userInput := os.Args[2]
		issue(userInput, aiService)
	default:
		log.Fatalf("Error: Unknown command '%s'. Use 'aidy help' for usage.\n", command)
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

func squash(gitService git.Git, shell executor.Executor) {
	baseBranch, err := gitService.GetBaseBranchName()
	if err != nil {
		log.Fatalf("Error determining base branch: %v", err)
	}
	_, resetErr := shell.RunCommand("git", "reset", "--soft", baseBranch)
	if resetErr != nil {
		log.Fatalf("Error executing git reset: %v", err)
	}
	commit(gitService, shell)
}

func commit(gitService git.Git, shell executor.Executor) {
	branchName, err := gitService.GetBranchName()
	if err != nil {
		log.Fatalf("Error getting branch name: %v", err)
	}
	issueNumber := extractIssueNumber(branchName)
	prompt := fmt.Sprintf(ai.GenerateCommitPrompt, issueNumber, issueNumber)
	_, shellErr := shell.RunCommand("aider", "--commit", "--commit-prompt", prompt)
	if shellErr != nil {
		log.Fatalf("Error executing aider --commit: %v", err)
	}
	heal(gitService, shell)
}

func pull_request(gitService git.Git, aiService ai.AI) {
	branchName, err := gitService.GetBranchName()
	if err != nil {
		log.Fatalf("Error getting branch name: %v", err)
	}
	diff, err := gitService.GetDiff()
	if err != nil {
		log.Fatalf("Error getting git diff: %v", err)
	}
	title, err := aiService.GenerateTitle(branchName, diff)
	if err != nil {
		log.Fatalf("Error generating title: %v", err)
	}

	body, err := aiService.GenerateBody(branchName, diff)
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
