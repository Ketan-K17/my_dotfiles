# sn-journal

Write a daily diary entry in Neovim; every `:w` lands in a Standard Notes tag.

- `:Journal` in Neovim opens today's entry (the day rolls over at 4 AM), titled like `Sunday, November 1, '26`. `bin/journal` does the same from a shell.
- Each save pushes the whole file in the background; Neovim's text wins.
- Closing the buffer (or quitting) deletes the local copy once Standard Notes has it, and clears registers holding text yanked from it. If a push failed the copy is kept and the next `:Journal` reopens it.
- No swap or undo files are written for the entry.

## Setup

    brew install go jonhadfield/tap/sn-cli
    sn session --add          # stores the session in the macOS Keychain
    go build -o bin/sn-journal .

Load it as a local plugin with lazy.nvim: `{ dir = "~/Developer/sn-journal", cmd = "Journal" }`.

`vim.g.sn_journal_tag` or `SN_JOURNAL_TAG` sets the tag (exact title, must already exist).
