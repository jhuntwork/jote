package jote

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/zeebo/blake3"
)

const (
	template = "---\ntitle:\ntags: []\n---\n"
	defPerms = 0o600
	dirPerms = 0o700
)

var (
	errNewJote   = errors.New("unable to create a new Jote")
	errNewTmpdir = errors.New("unable to create a temporary directory")
)

type Entry struct {
	Title string   `json:"title"`
	Tags  []string `json:"tags"`
}

type Jote struct {
	root       string
	repo       *git.Repository
	editor     exec.Cmd
	editorArgs []string
}

func DefaultStoreLocation() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share")
}

// NewJote will create a Jote store in a directory called jote under storePath.
func NewJote(storePath string, editor exec.Cmd) (Jote, error) {
	root := filepath.Join(storePath, "jote")
	repo, err := git.PlainOpen(root)
	if err != nil {
		if repo, err = git.PlainInit(root, false); err != nil {
			return Jote{},
				fmt.Errorf("%v: %w", errNewJote, err)
		}
		// Create a new empty branch named after this system's hostname
		hostname, err := os.Hostname()
		if err != nil {
			return Jote{}, fmt.Errorf("%v: %w", errNewJote, err)
		}
		h := plumbing.NewSymbolicReference(
			plumbing.HEAD,
			plumbing.NewBranchReferenceName(strings.Split(hostname, ".")[0]))
		if err := repo.Storer.SetReference(h); err != nil {
			return Jote{}, fmt.Errorf("%v: %w", errNewJote, err)
		}
	}
	return Jote{
		editor:     editor,
		editorArgs: editor.Args,
		root:       root,
		repo:       repo,
	}, nil
}

func (j *Jote) commit(newname string, oldname string) error {
	message := newname
	wt, err := j.repo.Worktree()
	if err != nil {
		return fmt.Errorf("error running Worktree: %w", err)
	}
	status, err := wt.Status()
	if err != nil {
		return err
	}
	// Either there was no oldname given, or it's the same as the new one
	if oldname == "" || oldname == newname {
		_, err = wt.Add(newname)
		if err != nil {
			return fmt.Errorf("filename: %s, error running Add: %w", newname, err)
		}
	}
	// Oldname was given, but it is an untracked file. Move it manually first, then Add.
	if oldname != "" && status.IsUntracked(oldname) {
		if err := os.Rename(filepath.Join(j.root, oldname), filepath.Join(j.root, newname)); err != nil {
			return fmt.Errorf("could not rename: %w", err)
		}
		_, err = wt.Add(newname)
		if err != nil {
			return fmt.Errorf("filename: %s, error running Add: %w", newname, err)
		}
	}
	// Oldname is different but it is a tracked file, already committed. Use Move.
	if oldname != newname && !status.IsUntracked(oldname) {
		message = fmt.Sprintf("%s -> %s", newname, oldname)
		_, err = wt.Move(oldname, newname)
		if err != nil {
			return fmt.Errorf("filename: %s, error running Move: %w", oldname, err)
		}
	}
	_, err = wt.Commit(message, &git.CommitOptions{})
	return err
}

func (j *Jote) List() error {
	var filedata []string
	err := filepath.Walk(
		j.root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && info.Name() == ".git" {
				return filepath.SkipDir
			}
			if !info.IsDir() {
				filedata = append(filedata, strings.TrimPrefix(path, fmt.Sprintf("%s/", j.root)))
			}
			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("unable to walk the directory tree: %w", err)
	}
	baseName, err := j.selectFileFromList(filedata)
	if err != nil {
		return err
	}
	return j.review(baseName, false)
}

func (j *Jote) selectFileFromList(files []string) (string, error) {
	sort.Sort(sort.Reverse(sort.StringSlice(files)))
	chosen, err := fuzzyfinder.Find(
		files,
		func(i int) string {
			return files[i]
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			content, _ := os.ReadFile(filepath.Join(j.root, files[i]))
			return string(content)
		}),
	)
	if err != nil {
		return "", err
	}
	return files[chosen], nil
}

func (j *Jote) Tags() error {
	tags := make(map[string][]string)
	err := filepath.Walk(
		j.root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && info.Name() == ".git" {
				return filepath.SkipDir
			}
			if !info.IsDir() {
				entry, err := parseToEntry(path)
				if err != nil {
					return err
				}
				for _, tag := range entry.Tags {
					if tags[tag] == nil {
						tags[tag] = []string{}
					}
					tags[tag] = append(tags[tag], strings.TrimPrefix(path, fmt.Sprintf("%s/", j.root)))
				}
			}
			return nil
		},
	)
	if err != nil {
		return fmt.Errorf("unable to walk the directory tree: %w", err)
	}
	tagSlice := make([]string, 0, len(tags))
	for key := range tags {
		tagSlice = append(tagSlice, key)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(tagSlice)))
	chosen, err := fuzzyfinder.Find(
		tagSlice,
		func(i int) string {
			return tagSlice[i]
		},
		fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
			if i == -1 {
				return ""
			}
			return fmt.Sprintf("Files with the tag %s:\n\n%s\n", tagSlice[i], strings.Join(tags[tagSlice[i]], "\n"))
		}),
	)
	if err != nil {
		return err
	}
	tag := tags[tagSlice[chosen]]
	baseName, err := j.selectFileFromList(tag)
	if err != nil {
		return err
	}
	return j.review(baseName, false)
}

func (j *Jote) Add() error {
	baseName := fmt.Sprintf("%d.md", time.Now().Unix())
	filename := filepath.Join(j.root, baseName)
	if err := os.WriteFile(filename, []byte(template), defPerms); err != nil {
		return fmt.Errorf("unable to write: %w", err)
	}
	return j.review(baseName, true)
}

func (j *Jote) review(baseName string, isNew bool) error {
	var oldname string
	filename := filepath.Join(j.root, baseName)
	changed, err := j.edit(filename)
	if err != nil {
		return err
	}
	if !changed {
		if isNew {
			os.Remove(filename)
		}
		return nil
	}
	entry, err := parseToEntry(filename)
	if err != nil {
		return err
	}
	if entry.Title != "" {
		oldname = baseName
		baseName = fmt.Sprintf("%s.md", entry.Title)
		newFullPath := filepath.Join(j.root, baseName)
		if err := os.MkdirAll(filepath.Dir(newFullPath), dirPerms); err != nil {
			return fmt.Errorf("could not ensure directory: %w", err)
		}
	}
	return j.commit(baseName, oldname)
}

func parseToEntry(filename string) (Entry, error) {
	var entry Entry
	fileReader, err := os.Open(filename)
	if err != nil {
		return entry, fmt.Errorf("unable to open file: %w", err)
	}
	defer fileReader.Close()

	_, err = frontmatter.Parse(fileReader, &entry)
	if err != nil {
		return entry, fmt.Errorf("could not parse: %w", err)
	}
	return entry, nil
}

func (j *Jote) edit(filename string) (bool, error) {
	preB3Sum, err := computeB3SumFromFile(filename)
	if err != nil {
		return false, err
	}
	j.editor.Args = append(j.editorArgs, filename) //nolint:gocritic // Not appending result to same slice is expected.
	if err := j.editor.Run(); err != nil {
		return false, fmt.Errorf("error calling editor: %w", err)
	}
	postB3Sum, err := computeB3SumFromFile(filename)
	if err != nil {
		return false, err
	}
	if preB3Sum == postB3Sum {
		return false, nil
	}
	return true, nil
}

func computeB3Sum(f io.Reader) (string, error) {
	var buf []byte
	hash := blake3.New()
	if _, err := io.Copy(hash, f); err != nil {
		return "", err
	}
	sum := hash.Sum(buf)
	return hex.EncodeToString(sum), nil
}

func computeB3SumFromFile(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return computeB3Sum(f)
}
