package main

import (
	"fmt"
	"os"

	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mygit",
	Short: "A Git clone built from scratch in Go",
	Long:  "mygit reimplements Git's core architecture from first principles.",
}

var initCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "Create an empty mygit repository",
	Long: `Initialize a new mygit repository by creating the standard metadata layout:

  .git/
  .git/objects/
  .git/refs/
  .git/refs/heads/
  .git/refs/tags/
  .git/HEAD`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	initObjectCommands()
	initTreeCommands()
	initCommitCommands()
	initLogCommands()
	initBranchCommands()
	initStatusCommands()
}

func runInit(cmd *cobra.Command, args []string) error {
	opts := repository.InitOptions{}
	if len(args) == 1 {
		opts.WorkTree = args[0]
	}

	repo, err := repository.Init(opts)
	if err != nil {
		return err
	}

	fmt.Printf("Initialized empty mygit repository in %s\n", repo.GitDir)
	return nil
}
