package main

import (
	"os"

	"github.com/mrunmayee/mygit/internal/commit"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/spf13/cobra"
)

var logMaxCount int

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show commit history",
	Long:  "Walk the commit graph from HEAD, following parent links newest-first.",
	RunE:  runLog,
}

func initLogCommands() {
	logCmd.Flags().IntVarP(&logMaxCount, "max-count", "n", 0, "limit number of commits")
	rootCmd.AddCommand(logCmd)
}

func runLog(cmd *cobra.Command, args []string) error {
	repo, err := repository.Discover("")
	if err != nil {
		return err
	}

	entries, err := commit.Log(repo, commit.LogOptions{MaxCount: logMaxCount})
	if err != nil {
		return err
	}

	output := commit.FormatLog(entries)
	if output != "" {
		os.Stdout.WriteString(output)
	}
	return nil
}
