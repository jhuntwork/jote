package cmd

import (
	"github.com/jhuntwork/jote"
	"github.com/spf13/cobra"
)

func (jc *joteCmd) ls(_ *cobra.Command, _ []string) error {
	js, err := jote.NewJote(jote.DefaultStoreLocation(), jc.editor)
	if err != nil {
		return err
	}
	return js.List()
}
