**`z`** jumps straight to the single best match.

- `z proj` goes to the highest-ranked directory matching "proj". It's fast, but if the match
  is wrong you end up in the wrong place. 
- You can pass several words to narrow it: `z dot ghost` goes to the best match containing both, like 
`~/dotfiles/ghostty`. 
- With no argument, `z` goes to your home directory. 

**`zi`** opens an fzf picker. 

- `zi proj` shows all matching directories, ranked, and you pick one with the arrow keys 
and Enter. 
- With no argument it lists everything in the zoxide database. 
- It's the better choice when several directories match and you aren't sure which one `z` would choose.
  Use `z` when the name is distinctive and you're confident about the result. Use `zi` when it's 
  ambiguous or you want to see the options first.

One thing to know: zoxide's database starts empty and only fills as you `cd` around. Until you've visited a directory at least once, neither command will find it.
