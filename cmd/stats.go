package cmd

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/volodya-lombrozo/aidy/internal/metrics"
)

func newStatsCmd(usage metrics.Metrics) *cobra.Command {
	command := &cobra.Command{
		Use:   "stats",
		Short: "Print how often each aidy command was used",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			runs, err := usage.Runs()
			if err != nil {
				return err
			}
			return printStats(cmd.OutOrStdout(), metrics.Summarize(runs, trackable(cmd.Root())))
		},
	}
	return command
}

func printStats(out io.Writer, summaries []metrics.Summary) error {
	table := tabwriter.NewWriter(out, 0, 0, 3, ' ', 0)
	if _, err := fmt.Fprintln(table, "command\truns\tfailed\tlast used"); err != nil {
		return err
	}
	for _, summary := range summaries {
		if _, err := fmt.Fprintf(table, "%s\t%d\t%s\t%s\n", summary.Command, summary.Runs, failed(summary), lastUsed(summary)); err != nil {
			return err
		}
	}
	return table.Flush()
}

func failed(summary metrics.Summary) string {
	if summary.Runs == 0 {
		return "-"
	}
	return fmt.Sprintf("%d", summary.Failed)
}

func lastUsed(summary metrics.Summary) string {
	if summary.Runs == 0 {
		return "never"
	}
	return summary.Last.Local().Format("2006-01-02")
}
