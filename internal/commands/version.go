package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

// These variables are set at build time via ldflags:
// go build -ldflags "-X github.com/ramtinJ95/EgenSkriven/internal/commands.Version=1.0.0"
var (
	Version   = "dev"
	BuildDate = "unknown"
	GitCommit = "unknown"
)

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Long:  `Display version, build date, and runtime information.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter()

			info := map[string]string{
				"version":    Version,
				"build_date": BuildDate,
				"git_commit": GitCommit,
				"go_version": runtime.Version(),
				"os":         runtime.GOOS,
				"arch":       runtime.GOARCH,
			}

			if out.JSON {
				return json.NewEncoder(os.Stdout).Encode(info)
			}

			fmt.Printf("EgenSkriven %s\n", Version)
			fmt.Printf("Build date: %s\n", BuildDate)
			fmt.Printf("Git commit: %s\n", GitCommit)
			fmt.Printf("Go version: %s\n", runtime.Version())
			fmt.Printf("OS/Arch:    %s/%s\n", runtime.GOOS, runtime.GOARCH)

			return nil
		},
	}

	return cmd
}
