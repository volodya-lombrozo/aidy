package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/volodya-lombrozo/aidy/internal/config"
)

var defaultModelIDs = map[string]string{
	"deepseek":  "deepseek-chat",
	"openai":    "gpt-4o",
	"anthropic": "claude-sonnet-4-6",
}

var providers = []string{"deepseek", "openai", "anthropic"}

func newInitCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "init",
		Short: "Create a ~/.aidy.conf.yml configuration file",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("error getting home directory: %v", err)
			}
			path := filepath.Join(home, ".aidy.conf.yml")
			return runInit(cmd.InOrStdin(), cmd.OutOrStdout(), path)
		},
	}
	return command
}

func runInit(in io.Reader, out io.Writer, path string) error {
	reader := bufio.NewReader(in)
	if _, err := os.Stat(path); err == nil {
		overwrite, err := askYesNo(reader, out, fmt.Sprintf("configuration file already exists at '%s', overwrite it?", path))
		if err != nil {
			return err
		}
		if !overwrite {
			fmt.Fprintln(out, "aborted, the existing configuration file was left untouched.")
			return nil
		}
	}
	provider, err := askProvider(reader, out)
	if err != nil {
		return err
	}
	model, err := ask(reader, out, "model id", defaultModelIDs[provider])
	if err != nil {
		return err
	}
	key, err := askRequired(reader, out, fmt.Sprintf("%s API key", provider))
	if err != nil {
		return err
	}
	github, err := ask(reader, out, "GitHub API key (optional, press enter to skip)", "")
	if err != nil {
		return err
	}
	apiKeys := map[string]string{provider: key}
	if github != "" {
		apiKeys["github"] = github
	}
	conf := &config.YamlConfig{
		DefaultModel: provider,
		APIKeys:      apiKeys,
		Models: map[string]map[string]string{
			provider: {
				"provider": provider,
				"model-id": model,
			},
		},
	}
	if err := config.WriteYaml(path, conf); err != nil {
		return fmt.Errorf("error writing configuration file: %v", err)
	}
	fmt.Fprintf(out, "configuration file created at '%s'\n", path)
	return nil
}

func askProvider(reader *bufio.Reader, out io.Writer) (string, error) {
	fmt.Fprintln(out, "choose an AI provider:")
	for i, p := range providers {
		fmt.Fprintf(out, "  (%d) %s\n", i+1, p)
	}
	for {
		fmt.Fprint(out, "enter the number of the provider to use: ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("error reading input: %v", err)
		}
		choice := strings.TrimSpace(line)
		for i, p := range providers {
			if choice == fmt.Sprintf("%d", i+1) || choice == p {
				return p, nil
			}
		}
		fmt.Fprintln(out, "invalid choice, please try again.")
	}
}

func ask(reader *bufio.Reader, out io.Writer, prompt string, def string) (string, error) {
	if def != "" {
		fmt.Fprintf(out, "%s [%s]: ", prompt, def)
	} else {
		fmt.Fprintf(out, "%s: ", prompt)
	}
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("error reading input: %v", err)
	}
	answer := strings.TrimSpace(line)
	if answer == "" {
		return def, nil
	}
	return answer, nil
}

func askRequired(reader *bufio.Reader, out io.Writer, prompt string) (string, error) {
	for {
		answer, err := ask(reader, out, prompt, "")
		if err != nil {
			return "", err
		}
		if answer != "" {
			return answer, nil
		}
		fmt.Fprintln(out, "this value is required, please try again.")
	}
}

func askYesNo(reader *bufio.Reader, out io.Writer, prompt string) (bool, error) {
	fmt.Fprintf(out, "%s [y/N]: ", prompt)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("error reading input: %v", err)
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}
