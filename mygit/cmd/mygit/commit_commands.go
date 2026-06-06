package main

import (
	"fmt"

	"github.com/mrunmayee/mygit/internal/commit"
	"github.com/mrunmayee/mygit/internal/refs"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/spf13/cobra"
)

var commitMessage string

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Record changes to the repository",
	Long:  "Create a commit object from the current working tree and advance the current branch.",
	RunE:  runCommit,
}

func initCommitCommands() {
	commitCmd.Flags().StringVarP(&commitMessage, "message", "m", "", "commit message (required)")
	_ = commitCmd.MarkFlagRequired("message")
	rootCmd.AddCommand(commitCmd)
}

func runCommit(cmd *cobra.Command, args []string) error {
	repo, err := repository.Discover("")
	if err != nil {
		return err
	}

	objectID, err := commit.Create(repo, commit.CreateOptions{
		Message: commitMessage,
	})
	if err != nil {
		return err
	}

	fmt.Printf("[%s %s] %s\n", headBranchLabel(repo), shortHash(objectID), commitMessage)
	return nil
}

func headBranchLabel(repo *repository.Repository) string {
	head, err := refs.ReadHEAD(repo)
	if err != nil || head.Branch == "" {
		return "main"
	}
	return head.Branch
}

func shortHash(hash string) string {
	if len(hash) >= 7 {
		return hash[:7]
	}
	return hash
}
