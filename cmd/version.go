package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Version is set by goreleaser via -X github.com/volodya-lombrozo/aidy/cmd.Version.
// `go install` does not pass that ldflag, so resolvedVersion falls back to the
// module version from the build info (issue #298).
var Version = "dev"

func resolvedVersion() string {
	return resolveVersion(Version, debug.ReadBuildInfo)
}

func resolveVersion(version string, info func() (*debug.BuildInfo, bool)) string {
	if version != "" && version != "dev" {
		return version
	}
	if version == "" {
		version = "dev"
	}
	bi, ok := info()
	if !ok || bi == nil {
		return version
	}
	mod := bi.Main.Version
	if mod != "" && mod != "(devel)" {
		return mod
	}
	return version
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the current aidy version",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), resolvedVersion())
			return err
		},
	}
}
