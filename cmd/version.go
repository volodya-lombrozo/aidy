package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

const devVersion = "dev"

var Version = devVersion

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the current aidy version",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), currentVersion())
			return err
		},
	}
}

func currentVersion() string {
	return resolveVersion(Version, debug.ReadBuildInfo)
}

func resolveVersion(stamped string, buildInfo func() (*debug.BuildInfo, bool)) string {
	if stamped != devVersion {
		return stamped
	}
	info, ok := buildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return devVersion
	}
	return info.Main.Version
}
