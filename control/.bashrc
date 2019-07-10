

alias l="tree --dirsfirst -ChFLa 1"
alias d="du -chd 1 | sort -h"
alias dc="docker-compose"

export PS1="\[\033[1;31m\]embly \[\e[0m\]\$(date +%M%S) \u: \w \n$ "

export HISTFILESIZE=72000
export HISTSIZE=$HISTFILESIZE
export HISTCONTROL=ignoreboth:erasedups
shopt -s histappend
export PROMPT_COMMAND="history -a; history -c; history -r; $PROMPT_COMMAND"

export PATH=$PATH:/go/src/github.com/maxmcd/embly/tools
