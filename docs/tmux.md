# tmux

A terminal multiplexer: it keeps your shells running in the background, lets you split one terminal window into many, and lets you detach and reattach later (e.g. after closing the terminal or SSH dropping).

Config lives in `tmux.conf` (currently only cursor settings, so everything below uses tmux defaults).

## Concepts

tmux has a three-level hierarchy:

```
server
└── session        a named workspace (e.g. "work", "notes")
    └── window     like a tab; fills the whole screen
        └── pane   a split inside a window; each runs its own shell
```

| Term | What it is | Browser/IDE analogy |
|------|------------|---------------------|
| **Session** | A collection of windows. Survives when you detach or close the terminal. | A whole workspace/project |
| **Window** | One full-screen view inside a session. Listed in the status bar at the bottom. tmux calls these windows, but they behave like **tabs**. | A tab |
| **Pane** | A rectangular split of a window, each running its own program. | A split editor view |
| **Prefix** | The key chord you press *before* any tmux command. Default: `Ctrl-b`. | — |

So when you say "tabs" in tmux, you mean **windows**.

### The prefix key

Almost every command is: press `Ctrl-b`, release, then press the command key.
Below, `prefix` means `Ctrl-b`. So `prefix c` = press `Ctrl-b`, let go, press `c`.

## Sessions

From the shell:

```sh
tmux                      # start a new unnamed session
tmux new -s work          # start a session named "work"
tmux ls                   # list sessions
tmux attach -t work       # reattach to "work" (alias: tmux a -t work)
tmux kill-session -t work # kill a session
tmux kill-server          # kill everything
```

Inside tmux:

| Keys | Action |
|------|--------|
| `prefix d` | **Detach** (session keeps running in the background) |
| `prefix s` | Pick a session from a list |
| `prefix $` | Rename current session |
| `prefix (` / `prefix )` | Previous / next session |

## Windows (tabs)

| Keys | Action |
|------|--------|
| `prefix c` | **Create** a new window |
| `prefix ,` | Rename current window |
| `prefix n` / `prefix p` | Next / previous window |
| `prefix 0`–`9` | Jump to window by number |
| `prefix w` | Pick a window from a list (across sessions) |
| `prefix l` | Toggle back to the last window |
| `prefix &` | **Close** current window (asks for confirmation) |

Or just type `exit` (or `Ctrl-d`) in the shell; when the last pane of a window exits, the window closes.

## Panes (splits)

| Keys | Action |
|------|--------|
| `prefix %` | Split **left/right** (vertical divider) |
| `prefix "` | Split **top/bottom** (horizontal divider) |
| `prefix ←↑↓→` | Move to the pane in that direction |
| `prefix o` | Cycle to the next pane |
| `prefix ;` | Toggle back to the last pane |
| `prefix q` | Flash pane numbers; press a number to jump |
| `prefix x` | **Close** current pane (asks for confirmation) |
| `prefix z` | **Zoom** a pane to full window; press again to restore |
| `prefix {` / `prefix }` | Swap pane with previous / next |
| `prefix Space` | Cycle through preset layouts |
| `prefix !` | Break the pane out into its own window |

Resize: hold `prefix` then press `Ctrl-←↑↓→` (or `prefix` then `:resize-pane -L 10` etc.). On macOS, `Ctrl-arrow` may be grabbed by Mission Control; use the command form if so.

Closing a pane: `prefix x`, or `exit` / `Ctrl-d` in it. Closing the last pane closes the window; closing the last window ends the session.

## Copy mode (scrolling)

Mouse wheel doesn't scroll by default. Instead:

| Keys | Action |
|------|--------|
| `prefix [` | Enter copy/scroll mode |
| `↑↓` / `PgUp` `PgDn` | Scroll |
| `/` or `?` | Search forward / backward |
| `q` | Exit copy mode |

## Command prompt

`prefix :` opens a command line, e.g. `:new-window`, `:split-window -h`, `:kill-pane`. Everything above is a shortcut for one of these commands.

## Misc

| Keys | Action |
|------|--------|
| `prefix ?` | List all key bindings |
| `prefix t` | Show a clock |
| `prefix r` | (only if you add it) reload config |

Reload the config without restarting:

```sh
tmux source-file ~/.tmux.conf
```

## Common tweaks (not yet in `tmux.conf`)

```tmux
set -g mouse on                      # click panes, drag borders, scroll wheel
set -g base-index 1                  # number windows from 1
setw -g pane-base-index 1
bind | split-window -h -c "#{pane_current_path}"   # nicer split keys, keep cwd
bind - split-window -v -c "#{pane_current_path}"
bind r source-file ~/.tmux.conf \; display "reloaded"
```

## Quick start

1. `tmux new -s work`
2. `prefix %` to split, work in both panes (`prefix ←/→` to move)
3. `prefix c` for a second window, `prefix n`/`p` to switch
4. `prefix d` to detach, close the terminal if you like
5. Later: `tmux a -t work` and everything is still there
