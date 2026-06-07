package main

import (
	"fmt"
	"os"

	"github.com/mrunmayee/mygit/internal/branch"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/spf13/cobra"
)

var branchCmd = &cobra.Command{
	Use:   "branch [name]",
	Short: "List or create branches",
	Long: `List local branches, or create a new branch at the current HEAD commit.

Examples:
  mygit branch
  mygit branch feature`,
	Args: cobra.MaximumNArgs(1),
	RunE: runBranch,
}

var checkoutCmd = &cobra.Command{
	Use:   "checkout <branch-or-commit>",
	Short: "Switch branches or enter detached HEAD",
	Args:  cobra.ExactArgs(1),
	RunE:  runCheckout,
}

func initBranchCommands() {
	rootCmd.AddCommand(branchCmd)
	rootCmd.AddCommand(checkoutCmd)
}

func runBranch(cmd *cobra.Command, args []string) error {
	repo, err := repository.Discover("")
	if err != nil {
		return err
	}

	if len(args) == 1 {
		if err := branch.Create(repo, args[0]); err != nil {
			return err
		}
		return nil
	}

	branches, err := branch.List(repo)
	if err != nil {
		return err
	}

	output := branch.FormatList(branches)
	if output != "" {
		fmt.Println(output)
	}
	return nil
}

func runCheckout(cmd *cobra.Command, args []string) error {
	repo, err := repository.Discover("")
	if err != nil {
		return err
	}

	target := args[0]
	if err := branch.Checkout(repo, target); err != nil {
		return err
	}

	head, err := repo.ReadHEAD()
	if err != nil {
		return err
	}

	if len(head) >= 4 && head[:4] == "ref:" {
		name := target
		fmt.Fprintf(os.Stdout, "Switched to branch '%s'\n", name)
		return nil
	}

	fmt.Fprintf(os.Stdout, "Note: switching to '%s'.\n", shortHash(target))
	fmt.Fprintln(os.Stdout, "You are in 'detached HEAD' state.")
	return nil
}
