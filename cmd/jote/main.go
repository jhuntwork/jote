package main

import (
	"fmt"
	"os"

	"github.com/jhuntwork/jote/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Stderr.WriteString(fmt.Sprintf("%s\n", err.Error()))
		os.Exit(1)
	}
}
