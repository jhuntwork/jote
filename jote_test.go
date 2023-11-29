package jote_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jhuntwork/jote"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJote(t *testing.T) {
	t.Parallel()
	t.Run("should get a Store", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)
		tmpdir, err := os.MkdirTemp("", "jote-test-*")
		if err != nil {
			assert.FailNow("unable to create a temporary directory")
		}
		defer os.RemoveAll(tmpdir)
		js, err := jote.NewJote(tmpdir, exec.Cmd{})
		assert.NotNil(js)
		require.NoError(t, err)
	})
}

func TestAdd(t *testing.T) {
	t.Parallel()
	t.Run("should launch the command specified by the command", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)
		tmpdir, err := os.MkdirTemp("", "jote-test-*")
		if err != nil {
			assert.FailNow("unable to create a temporary directory")
		}
		defer os.RemoveAll(tmpdir)

		testout := filepath.Join(tmpdir, "testout")
		data := []byte(fmt.Sprintf("#!/bin/sh\nprintf 'Called with: %%s' \"$*\" > %s\n", testout))
		filename := filepath.Join(tmpdir, "test-editor")
		if err := os.WriteFile(filename, data, 0o755); err != nil { //nolint:gosec
			assert.FailNow("unable to create test-editor command")
		}

		cmd := exec.Command(filename, "arg1", "arg2")
		js, err := jote.NewJote(tmpdir, *cmd)
		if err != nil {
			assert.FailNow("encountered an error creating Store")
		}

		if err := js.Add(); err != nil {
			assert.FailNowf("encountered an unexpected error when Adding: %s", err.Error())
		}
		output, err := os.ReadFile(testout)
		require.NoError(t, err)
		assert.Contains(string(output), "Called with: arg1 arg2")
	})
}

func TestDefaultStoreLocation(t *testing.T) {
	t.Parallel()
	t.Run("should return a predictable location", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)
		loc := jote.DefaultStoreLocation()
		home, _ := os.UserHomeDir()
		assert.Equal(filepath.Join(home, ".local", "share"), loc)
	})
}
