package cmd

import (
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

type joteCmd struct {
	editor exec.Cmd
}

const (
	description = "\njote jots down notes"
)

func Execute() error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim --nofork"
	}
	editorWithArgs := strings.Split(editor, " ")
	command := exec.Command(editorWithArgs[0], editorWithArgs[1:]...) //nolint:gosec // Expected to execute user editor
	command.Env = os.Environ()
	command.Stderr = os.Stderr
	command.Stdout = os.Stdout
	command.Stdin = os.Stdin
	jc := &joteCmd{
		editor: *command,
	}

	rootCmd := &cobra.Command{
		Short:         description,
		RunE:          jc.new,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	NewCmd := &cobra.Command{
		Use:           "new",
		Short:         "jot down a new note",
		RunE:          jc.new,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	rootCmd.AddCommand(NewCmd)

	LsCmd := &cobra.Command{
		Use:           "ls",
		Short:         "list existing notes",
		RunE:          jc.ls,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	rootCmd.AddCommand(LsCmd)

	TagsCmd := &cobra.Command{
		Use:           "tags",
		Short:         "search notes by tags",
		RunE:          jc.tags,
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	rootCmd.AddCommand(TagsCmd)

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	return rootCmd.Execute()
}
