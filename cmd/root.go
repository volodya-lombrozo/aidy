package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/volodya-lombrozo/aidy/internal/aidy"
	"github.com/volodya-lombrozo/aidy/internal/config"
	"github.com/volodya-lombrozo/aidy/internal/metrics"
)

type Context struct {
	Assistant aidy.Aidy
}

func Execute() error {
	usage := Usage()
	return Tracked(NewRootCmd(Real, usage), usage)
}

func Tracked(root *cobra.Command, usage metrics.Metrics) error {
	start := time.Now()
	executed, err := root.ExecuteC()
	if tracked(executed) {
		if rerr := usage.Record(metrics.NewRun(executed.Name(), start, err)); rerr != nil {
			_, _ = fmt.Fprintf(root.ErrOrStderr(), "failed to record usage metrics: %v\n", rerr)
		}
	}
	return err
}

func Usage() metrics.Metrics {
	home, err := os.UserHomeDir()
	if err != nil {
		return metrics.NewDisabled(metrics.NewMock())
	}
	file := metrics.NewFile(filepath.Join(home, ".aidy", "metrics.json"))
	conf, err := config.NewCascadeInDirs(os.Getwd, gitRoot, os.UserHomeDir)
	if err != nil {
		return file
	}
	if enabled, err := conf.Metrics(); err == nil && !enabled {
		return metrics.NewDisabled(file)
	}
	return file
}

func gitRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func trackable(root *cobra.Command) []string {
	var names []string
	for _, command := range root.Commands() {
		if tracked(command) {
			names = append(names, command.Name())
		}
	}
	return names
}

func tracked(command *cobra.Command) bool {
	return command != nil &&
		command.HasParent() &&
		command.Parent() == command.Root() &&
		command.IsAvailableCommand() &&
		command.Name() != "completion"
}

func Real(summary, aider, ailess, silent, debug bool, language string) aidy.Aidy {
	return aidy.NewAidy(summary, aider, ailess, silent, debug, language)
}

func NewRootCmd(create func(bool, bool, bool, bool, bool, string) aidy.Aidy, usage metrics.Metrics) *cobra.Command {
	var ctx Context
	var ailess bool
	var aider bool
	var summary bool
	var silent bool
	var debug bool
	var language string
	root := &cobra.Command{
		Use:     "aidy",
		Short:   "aidy - ai-powered github cli helper",
		Long:    "Aidy assists you with generating commit messages, pull requests, issues, and releases",
		Version: Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmd.Root().SilenceUsage = true
			cmd.Root().SilenceErrors = true
			ctx.Assistant = create(summary, aider, ailess, silent, debug, language)
		},
	}
	root.PersistentFlags().BoolVarP(&ailess, "no-ai", "n", false, "don't use AI")
	root.PersistentFlags().BoolVarP(&summary, "summary", "s", false, "use a project summary in AI requests")
	root.PersistentFlags().BoolVar(&aider, "aider", false, "use aider configuration")
	root.PersistentFlags().BoolVarP(&silent, "quiet", "q", false, "be silent, don't print logs")
	root.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "print debug logs")
	root.PersistentFlags().StringVarP(&language, "language", "l", "en", "language for AI-generated text (e.g. fr, de, ja)")
	root.AddCommand(
		newInitCmd(),
		newCommitCmd(&ctx),
		newIssueCmd(&ctx),
		newReleaseCmd(&ctx),
		newPrCmd(&ctx),
		newMrCmd(&ctx),
		newHealCmd(&ctx),
		newSquashCmd(&ctx),
		newAppendCmd(&ctx),
		newConfigCmd(&ctx),
		newCleanCmd(&ctx),
		newStartCmd(&ctx),
		newDiffCmd(&ctx),
		newLastCmd(&ctx),
		newVersionCmd(),
		newStatsCmd(usage),
	)
	return root
}
