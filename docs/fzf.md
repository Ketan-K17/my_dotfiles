**`fzf`** is a fuzzy finder. It takes a list of lines (files, history, directories, anything), lets you type a few characters to filter it, and returns the line you pick.

## Shell keybindings

Set up in `zshrc` with `source <(fzf --zsh)`.

- **Ctrl-R** searches your command history. Type part of an old command, pick it with Enter, and it lands on the prompt so you can edit it before running.
- **Ctrl-T** searches files and pastes the chosen path onto the command line. It searches `~/Desktop` and `~/Downloads` (see `FZF_CTRL_T_COMMAND`).
- **Alt-C** searches directories and `cd`s into the one you pick.

Inside the picker:

- Arrow keys or Ctrl-J / Ctrl-K move the selection.
- Enter accepts, Esc or Ctrl-C cancels.
- Tab marks an item in multi-select mode (e.g. `vim **<Tab>`), Shift-Tab unmarks.

## Search syntax

- `foo bar` matches lines containing both words, in any order.
- `'foo` is an exact match, no fuzziness.
- `^foo` matches at the start of the line, `foo$` at the end.
- `!foo` excludes lines containing foo.
- `foo | bar` matches either.

## Fuzzy completion with `**`

Type `**` then Tab after a command to get a picker for that context:

- `vim **<Tab>` picks files.
- `cd **<Tab>` picks directories.
- `kill -9 <Tab>` picks a process.
- `ssh **<Tab>` picks a host.

## Piping into fzf

fzf works on anything that prints lines:

```sh
git branch | fzf                    # pick a branch
git checkout $(git branch | fzf)    # ...and switch to it
ps aux | fzf                        # pick a process
fd -e md | fzf --preview 'bat {}'   # pick a file with a preview pane
```

Useful flags: `--preview '<cmd>'` shows a preview pane, `-m` allows multi-select, `--height 40%` keeps it inline instead of full screen.

## Where it's used in these dotfiles

- `zi` (zoxide) opens an fzf picker, see `z_zi.md`.
- The `FZF_*_COMMAND` variables in `zshrc` use `fd` instead of `find`, so hidden files are included and `.git` is skipped.
- Vi mode is set before fzf is sourced, otherwise it wipes fzf's keybindings.
