package main

import (
	"os"

	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/status"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the working tree status",
	Long:  "Compare the working tree against the HEAD commit to find modified, deleted, and untracked files.",
	RunE:  runStatus,
}

func initStatusCommands() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	repo, err := repository.Discover("")
	if err != nil {
		return err
	}

	report, err := status.Compute(repo)
	if err != nil {
		return err
	}

	os.Stdout.WriteString(status.Format(report))
	return nil
}
