package main

import (
    "fmt"
    "log"
    "github.com/volodya-lombrozo/aidy/ai"
    "github.com/volodya-lombrozo/aidy/git"
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "os"
    "os/exec"
    "bytes"
    "strings"
    "regexp"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Error: No command provided. Use 'aidy help' for usage.")
        os.Exit(1)
    }
    command := os.Args[1]
    switch command {
    case "pr":
        handlePR()
    case "help":
        handleHelp()
    case "heal":
        handleHeal()
    case "ci", "commit":
        handleCommit()
    default:
        fmt.Printf("Error: Unknown command '%s'. Use 'aidy help' for usage.\n", command)
        os.Exit(1)
    }
}

func handleCommit() {
    // Execute aider --commit
    err := exec.Command("aider", "--commit").Run()
    if err != nil {
        log.Fatalf("Error executing aider --commit: %v", err)
    }
    // Execute aidy heal
    handleHeal()
}

func handlePR() {
    // Read API key from configuration file
    homeDir, err := os.UserHomeDir()
    if err != nil {
        log.Fatalf("Error getting home directory: %v", err)
    }
    configPath := fmt.Sprintf("%s/.aidy.conf.yml", homeDir)
    configData, err := ioutil.ReadFile(configPath)
    if err != nil {
        log.Fatalf("Error reading config file: %v", err)
    }

    var config struct {
        OpenAIAPIKey string `yaml:"openai-api-key"`
    }

    err = yaml.Unmarshal(configData, &config)
    if err != nil {
        log.Fatalf("Error parsing config file: %v", err)
    }

    apiKey := config.OpenAIAPIKey
    if apiKey == "" {
        log.Fatalf("OpenAI API key not found in config file")
    }
    gitService := &git.RealGit{}

    branchName, err := gitService.GetBranchName()
    if err != nil {
        log.Fatalf("Error getting branch name: %v", err)
    }

    diff, err := gitService.GetDiff("main")
    if err != nil {
        log.Fatalf("Error getting git diff: %v", err)
    }
    // Use the OpenAI implementation
    aiService := ai.NewOpenAI(apiKey, "gpt-4o", 0.3)

    title, err := aiService.GenerateTitle(branchName, diff)
    if err != nil {
        log.Fatalf("Error generating title: %v", err)
    }

    body, err := aiService.GenerateBody(branchName, diff)
    if err != nil {
        log.Fatalf("Error generating body: %v", err)
    }

    fmt.Printf("Generated PR Command:\ngh pr create --title \"%s\" --body \"%s\"\n", title, body)
}

func handleHelp() {
    fmt.Println("Usage:")
    fmt.Println("  aidy pr   - Generate a pull request using AI-generated title and body.")
    fmt.Println("  aidy help - Show this help message.")
}

func handleHeal() {
    // Read API key from configuration file
    homeDir, err := os.UserHomeDir()
    if err != nil {
        log.Fatalf("Error getting home directory: %v", err)
    }
    configPath := fmt.Sprintf("%s/.aidy.conf.yml", homeDir)
    configData, err := ioutil.ReadFile(configPath)
    if err != nil {
        log.Fatalf("Error reading config file: %v", err)
    }

    var config struct {
        OpenAIAPIKey string `yaml:"openai-api-key"`
    }

    err = yaml.Unmarshal(configData, &config)
    if err != nil {
        log.Fatalf("Error parsing config file: %v", err)
    }

    apiKey := config.OpenAIAPIKey
    if apiKey == "" {
        log.Fatalf("OpenAI API key not found in config file")
    }
    gitService := &git.RealGit{}

    branchName, err := gitService.GetBranchName()
    if err != nil {
        log.Fatalf("Error getting branch name: %v", err)
    }

    // Extract issue number from branch name
    issueNumber := extractIssueNumber(branchName)

    // Get the current commit message
    cmd := exec.Command("git", "log", "-1", "--pretty=%B")
    var out bytes.Buffer
    cmd.Stdout = &out
    err = cmd.Run()
    if err != nil {
        log.Fatalf("Error getting current commit message: %v", err)
    }
    commitMessage := out.String()
    re := regexp.MustCompile(`#\d+`)
    newCommitMessage := re.ReplaceAllString(commitMessage, fmt.Sprintf("#%s", issueNumber))
    fmt.Printf("Executing command: 'git commit --amend -m \"%s\"''", newCommitMessage)
    cmd = exec.Command("git", "commit", "--amend", "-m", newCommitMessage)
    err = cmd.Run()
    if err != nil {
        log.Fatalf("Error amending commit message: %v", err)
    }
    fmt.Println("Commit message healed successfully.")
}


func extractIssueNumber(branchName string) string {
    // Assuming the branch name format is "<issue-number>_<description>"
    parts := strings.Split(branchName, "_")
    if len(parts) > 0 {
        return parts[0]
    }
    return "unknown"
}
