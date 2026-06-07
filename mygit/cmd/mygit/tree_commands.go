package main

import (
	"fmt"

	"github.com/mrunmayee/mygit/internal/index"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/storage"
	"github.com/mrunmayee/mygit/internal/tree"
	"github.com/spf13/cobra"
)

var writeTreeCmd = &cobra.Command{
	Use:   "write-tree",
	Short: "Create a tree object from the working directory",
	Long:  "Walk the working tree, store blobs and trees, and print the root tree hash.",
	RunE:  runWriteTree,
}

var lsTreeCmd = &cobra.Command{
	Use:   "ls-tree <hash>",
	Short: "List the contents of a tree object",
	Args:  cobra.ExactArgs(1),
	RunE:  runLsTree,
}

func initTreeCommands() {
	rootCmd.AddCommand(writeTreeCmd)
	rootCmd.AddCommand(lsTreeCmd)
}

func runWriteTree(cmd *cobra.Command, args []string) error {
	repo, err := repository.Discover("")
	if err != nil {
		return err
	}

	objectID, err := index.WriteTreeForCommit(repo)
	if err != nil {
		return err
	}

	fmt.Println(objectID)
	return nil
}

func runLsTree(cmd *cobra.Command, args []string) error {
	repo, err := repository.Discover("")
	if err != nil {
		return err
	}

	store := storage.New(repo)
	output, err := tree.ListTree(store, args[0])
	if err != nil {
		return err
	}

	if output != "" {
		fmt.Println(output)
	}
	return nil
}
