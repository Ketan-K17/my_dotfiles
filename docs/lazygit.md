`lg` is an alias I added for `lazygit`, a terminal UI for git. Instead of typing git commands, you 
get a full-screen interface where you move around with the keyboard. Run it inside any git repo:

- **Files panel:** `space` stages or unstages a file, `a` stages everything, and `c` commits. 
- **Branches, commits, stash:** Switch panes with `1`-`5` or `Tab`, press `enter` to dig in, and 
press `space` to check out a branch. 
- **Push, pull and rebase:** `P` pushes, `p` pulls, and `r` rebases. 
- **Diffs:** They show on the right, and git-delta from your Brewfile can make them 
prettier. 
- **Help:** `?` shows the keybindings for whichever pane you're in. 
- **Quit:** `q`.

It does nothing outside a git repo, and it won't help with `~/dotfiles` until you've made a
first commit, though it works there as a way to stage and commit. Lazygit is meant to replace 
most day-to-day `git add`, `commit`, `checkout` and `rebase` typing.
