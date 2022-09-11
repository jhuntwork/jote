# Jote

Jote is a simple Markdown notes manager.

## Why?

I wanted a simple way of organizing my notes, without being required to use a
new editor, or being bound to a specific editor. By design, such a method
should at minium be:

- Frictionless to use, especially when adding a new note
- Usable by any editor (that can be invoked by the command line)
- Automatically backed by git
- Able to support any hierarchy
- Searchable via custom tags

## Installing

For now, the easiest way to install is:

```sh
go install github.com/jhuntwork/jote/cmd/jote@latest
```

## Usage

### New notes

To create a new note, simply run `jote new` or just `jote`. If you have the
`EDITOR` environment variable set, that command will be invoked to open a new
file. The `EDITOR` command should block until the file is closed. For example,
the default command is `vim --nofork`, which means vim launches in the
foreground and will block until you close the editor. A similar command for
VS Code would be `code -w`.

On first run, `jote` will initialize a git repository in `~/.local/share/jote`.
Because all further actions there are git actions, you can interact with it as
you would any other git repository.

New files will contain the following template:

```md
---
title:
tags: []
---
```

The section between the `---` lines is considered front matter and is parsed as
yaml. If you provide a title, the file will be saved as `[title].md`, otherwise
the file will be named with a unix timestamp of the current time.

`title` is interpreted as a relative path in the git repository. This means that
notes can support a structured hierarchy. For example:

```yaml
title: recipes/Asian Food/Spicy Thai Noodles
```

The resulting file would be named
`~/.local/share/jote/recipes/Asian Food/Spicy Thai Noodles.md`.

`tags` is a list of strings, so it can either be specified like this:

```yaml
tags: [one, two, "another tag"]
```

or like this:

```yaml
tags:
  - one
  - two
  - another tag
```

Everything following the front matter section is the actual Markdown content.

### Listing all notes

Run `jote ls`. This will provide a `fzf`-like interface (via
[go-fuzzyfinder](https://github.com/ktr0731/go-fuzzyfinder)) to list all notes.
Selecting a note will open it for editing.

Changes made to pre-existing notes via the jote interface will result in a new
git commit tracking that change. Changing the title will move the file to the
new location.

### Searching by tags

Run `jote tags`. This provides a `fzf`-like interface of all known tags, showing
the associated documents for each tag. Select a tag and the interface will
switch to a list of those documents, allowing you to select one for editing.

## Future features

- An opinionated way for syncing to a git remote, allowing for editing across
  machines. For now, this can be managed through manual git commands in the
  repository.
- Outputting results of tag searches to stdout as a list, for consumption by
  other CLI tools.
- More advanced searches by content.
- Support a 'publish' keyword in the front matter. This could then be used by
  plugins to format and send documents to configured consumers.
