#!/bin/bash
# gf bash completion

_gf() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    opts="init status branch add commit merge finish mr help"

    case "${prev}" in
        gf)
            COMPREPLY=($(compgen -W "${opts}" -- ${cur}))
            return 0
            ;;
        -s|-b|-a|-c|-m|-f|-r)
            COMPREPLY=($(compgen -f -- ${cur}))
            return 0
            ;;
        *)
            ;;
    esac
}

complete -F _gf gf