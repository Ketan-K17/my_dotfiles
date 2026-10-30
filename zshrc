# Vi mode (set before fzf so fzf's keybindings aren't wiped)
bindkey -v
export KEYTIMEOUT=1

# History
HISTFILE=~/.zsh_history
HISTSIZE=50000
SAVEHIST=50000
setopt SHARE_HISTORY HIST_IGNORE_ALL_DUPS HIST_IGNORE_SPACE

# Completion
autoload -Uz compinit && compinit

# Editor
export EDITOR=nvim
export VISUAL=nvim

# Aliases
alias ls='eza --icons'
alias ll='eza -l --icons --git'
alias la='eza -la --icons --git'
alias cat='bat --paging=never'
alias v='nvim'
alias lg='lazygit'

# fzf: Ctrl-R history, Ctrl-T files, Alt-C directories
source <(fzf --zsh)
export FZF_DEFAULT_COMMAND='fd --type f --hidden --exclude .git'
export FZF_CTRL_T_COMMAND="$FZF_DEFAULT_COMMAND"
export FZF_ALT_C_COMMAND='fd --type d --hidden --exclude .git'

# zoxide: `z proj` jumps to a frequent directory
eval "$(zoxide init zsh)"

# Prompt
eval "$(starship init zsh)"

# Plugins (syntax-highlighting must stay last)
source /opt/homebrew/share/zsh-autosuggestions/zsh-autosuggestions.zsh
source /opt/homebrew/share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh
