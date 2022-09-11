package jote

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
)

func setupJote() (string, Jote, error) {
	tmpdir, err := os.MkdirTemp("", "jote-test-*")
	if err != nil {
		return "", Jote{}, errNewTmpdir
	}
	js, err := NewJote(tmpdir, exec.Cmd{})
	if err != nil {
		os.RemoveAll(tmpdir)
		return "", Jote{}, errNewJote
	}
	return tmpdir, js, nil
}

//nolint:funlen
func Test_edit(t *testing.T) {
	t.Parallel()
	t.Run("should return false if there were no changes made to the file", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)
		tmpdir, js, err := setupJote()
		if err != nil {
			assert.FailNow(err.Error())
		}
		defer os.RemoveAll(tmpdir)
		filename := filepath.Join(tmpdir, "edittest")
		if err := os.WriteFile(filename, []byte(""), 0o644); err != nil { //nolint:gosec
			assert.FailNow("unable to create testfile")
		}
		testEdit := filepath.Join(tmpdir, "test-edit")
		data := []byte("#!/bin/sh\necho true\n")
		if err := os.WriteFile(testEdit, data, 0o755); err != nil { //nolint:gosec
			assert.FailNow("unable to create test-editor command")
		}
		js.editor.Path = testEdit
		changed, err := js.edit(filename)
		assert.NoError(err)
		assert.False(changed)
	})
	t.Run("should return true if there were changes made to the file", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)
		tmpdir, js, err := setupJote()
		if err != nil {
			assert.FailNow(err.Error())
		}
		defer os.RemoveAll(tmpdir)
		filename := filepath.Join(tmpdir, "edittest")
		if err := os.WriteFile(filename, []byte(""), 0o644); err != nil { //nolint:gosec
			assert.FailNow("unable to create testfile")
		}
		testEdit := filepath.Join(tmpdir, "test-edit")
		data := []byte(fmt.Sprintf("#!/bin/sh\nprintf 'changed\n' > %s\n", filename))
		if err := os.WriteFile(testEdit, data, 0o755); err != nil { //nolint:gosec
			assert.FailNow("unable to create test-editor command")
		}
		js.editor.Path = testEdit
		changed, err := js.edit(filename)
		assert.NoError(err)
		assert.True(changed)
	})
	t.Run("should error if the file cannot be found", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)
		tmpdir, js, err := setupJote()
		if err != nil {
			assert.FailNow(err.Error())
		}
		defer os.RemoveAll(tmpdir)
		filename := filepath.Join(tmpdir, "fakefile")
		_, err = js.edit(filename)
		assert.Error(err)
		assert.Contains(err.Error(), "fakefile: no such file or directory")
	})
	t.Run("should error if the editor cannot be called", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)
		tmpdir, js, err := setupJote()
		if err != nil {
			assert.FailNow(err.Error())
		}
		defer os.RemoveAll(tmpdir)
		filename := filepath.Join(tmpdir, "edittest")
		if err := os.WriteFile(filename, []byte(""), 0o644); err != nil { //nolint:gosec
			assert.FailNow("unable to create testfile")
		}
		editorname := filepath.Join(tmpdir, "nosuchfile")
		js.editor.Path = editorname
		_, err = js.edit(filename)
		assert.Error(err)
		assert.Contains(err.Error(), "nosuchfile: no such file or directory")
	})
}

func TestNewJote(t *testing.T) {
	t.Parallel()
	t.Run("should have a predictable root location", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)
		tmpdir, js, err := setupJote()
		if err != nil {
			assert.FailNow(err.Error())
		}
		defer os.RemoveAll(tmpdir)
		root := filepath.Join(tmpdir, "jote")
		assert.NoError(err)
		assert.Equal(root, js.root)
	})
	t.Run("should be git-backed", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)
		tmpdir, js, err := setupJote()
		if err != nil {
			assert.FailNow(err.Error())
		}
		defer os.RemoveAll(tmpdir)
		assert.NoError(err)
		if ok := assert.NotNil(js.repo); !ok {
			return
		}
		assert.IsType(&git.Repository{}, js.repo)
	})
	t.Run("should create the root location if it doesn't exist", func(t *testing.T) {
		t.Parallel()
		assert := assert.New(t)
		tmpdir, js, err := setupJote()
		if err != nil {
			assert.FailNow(err.Error())
		}
		defer os.RemoveAll(tmpdir)
		assert.NoError(err)
		gitRoot := filepath.Join(js.root, ".git")
		finfo, err := os.Stat(gitRoot)
		if err != nil {
			if os.IsNotExist(err) {
				assert.FailNowf("root was not created", "%s does not exist", gitRoot)
			}
			assert.FailNow("encountered an unexpected stat error")
		}
		assert.True(finfo.IsDir())
	})
}
