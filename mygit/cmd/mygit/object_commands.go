package main

import (
	"fmt"
	"os"

	"github.com/mrunmayee/mygit/internal/objects"
	"github.com/mrunmayee/mygit/internal/repository"
	"github.com/mrunmayee/mygit/internal/storage"
	"github.com/spf13/cobra"
)

var hashObjectWrite bool

var hashObjectCmd = &cobra.Command{
	Use:   "hash-object <file>",
	Short: "Compute object ID and optionally store a file as a blob",
	Args:  cobra.ExactArgs(1),
	RunE:  runHashObject,
}

var catFileType bool

var catFileCmd = &cobra.Command{
	Use:   "cat-file <hash>",
	Short: "Read and display an object from the object database",
	Args:  cobra.ExactArgs(1),
	RunE:  runCatFile,
}

func initObjectCommands() {
	hashObjectCmd.Flags().BoolVarP(&hashObjectWrite, "write", "w", false, "write object to database")
	catFileCmd.Flags().BoolVarP(&catFileType, "type", "t", false, "show object type only")
	rootCmd.AddCommand(hashObjectCmd)
	rootCmd.AddCommand(catFileCmd)
}

func runHashObject(cmd *cobra.Command, args []string) error {
	repo, err := repository.Discover("")
	if err != nil {
		return err
	}

	objectID, err := storage.HashObject(storage.HashObjectOptions{
		Repo:     repo,
		FilePath: args[0],
		Write:    hashObjectWrite,
	})
	if err != nil {
		return err
	}

	fmt.Println(objectID)
	return nil
}

func runCatFile(cmd *cobra.Command, args []string) error {
	repo, err := repository.Discover("")
	if err != nil {
		return err
	}

	store := storage.New(repo)
	obj, err := store.Read(args[0])
	if err != nil {
		return err
	}

	if catFileType {
		fmt.Println(obj.Type)
		return nil
	}

	switch obj.Type {
	case objects.TypeBlob:
		os.Stdout.Write(obj.Content)
	case objects.TypeCommit:
		os.Stdout.Write(obj.Content)
	default:
		return fmt.Errorf("%w: %s", objects.ErrUnsupportedType, obj.Type)
	}

	return nil
}
