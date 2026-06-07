package main

import (
	"github.com/mrunmayee/mygit/internal/index"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <path> [path...]",
	Short: "Stage files for commit",
	Long: `Add file contents to the staging area (index).

Examples:
  mygit add file.txt
  mygit add .`,
	Args: cobra.MinimumNArgs(1),
	RunE: runAdd,
}

func initAddCommands() {
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	repo, err := repository.Discover("")
	if err != nil {
		return err
	}

	for _, path := range args {
		if err := index.Add(repo, path); err != nil {
			return err
		}
	}
	return nil
}
