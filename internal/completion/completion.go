package completion

// GenerateBash returns a bash completion script for pixshift.
func GenerateBash() string {
	return `_pixshift() {
    local cur prev opts formats completions
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    formats="jpg jpeg png gif webp tiff bmp heic avif"
    completions="bash zsh fish"

    case "${prev}" in
        -f|--format)
            COMPREPLY=( $(compgen -W "${formats}" -- "${cur}") )
            return 0
            ;;
        --completion)
            COMPREPLY=( $(compgen -W "${completions}" -- "${cur}") )
            return 0
            ;;
        -q|--quality|-j|--jobs|--width|--height|--max-dim)
            return 0
            ;;
        -o|--output)
            COMPREPLY=( $(compgen -d -- "${cur}") )
            return 0
            ;;
        -c|--config)
            COMPREPLY=( $(compgen -f -- "${cur}") )
            return 0
            ;;
        --template)
            return 0
            ;;
    esac

    if [[ "${cur}" == --* ]]; then
        opts="--format --quality --jobs --output --recursive --preserve-metadata --watch --config --overwrite --dry-run --verbose --version --help --width --height --max-dim --strip-metadata --template --completion"
        COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
        return 0
    fi

    if [[ "${cur}" == -* ]]; then
        opts="-f -q -j -o -r -m -w -c -v -V -h -s --format --quality --jobs --output --recursive --preserve-metadata --watch --config --overwrite --dry-run --verbose --version --help --width --height --max-dim --strip-metadata --template --completion"
        COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
        return 0
    fi

    COMPREPLY=( $(compgen -f -- "${cur}") )
}

complete -F _pixshift pixshift
`
}

// GenerateZsh returns a zsh completion script for pixshift.
func GenerateZsh() string {
	return `#compdef pixshift

_pixshift() {
    local -a formats completions
    formats=(jpg jpeg png gif webp tiff bmp heic avif)
    completions=(bash zsh fish)

    _arguments \
        '(-f --format)'{-f,--format}'[output format]:format:(${formats})' \
        '(-q --quality)'{-q,--quality}'[quality level]:quality:' \
        '(-j --jobs)'{-j,--jobs}'[number of parallel jobs]:jobs:' \
        '(-o --output)'{-o,--output}'[output directory]:directory:_directories' \
        '(-r --recursive)'{-r,--recursive}'[process directories recursively]' \
        '(-m --preserve-metadata)'{-m,--preserve-metadata}'[preserve image metadata]' \
        '(-w --watch)'{-w,--watch}'[watch for file changes]' \
        '(-c --config)'{-c,--config}'[config file path]:config:_files' \
        '--overwrite[overwrite existing files]' \
        '--dry-run[show what would be done without doing it]' \
        '(-v --verbose)'{-v,--verbose}'[enable verbose output]' \
        '(-V --version)'{-V,--version}'[show version]' \
        '(-h --help)'{-h,--help}'[show help]' \
        '--width[resize width]:width:' \
        '--height[resize height]:height:' \
        '--max-dim[maximum dimension]:max-dim:' \
        '(-s --strip-metadata)'{-s,--strip-metadata}'[strip image metadata]' \
        '--template[output filename template]:template:' \
        '--completion[generate shell completion]:shell:(${completions})' \
        '*:input files:_files'
}

_pixshift "$@"
`
}

// GenerateFish returns a fish completion script for pixshift.
func GenerateFish() string {
	return `# Fish completions for pixshift

# Disable file completions by default, re-enable for positional args
complete -c pixshift -f

# Format flag
complete -c pixshift -s f -l format -x -d 'Output format' -a 'jpg jpeg png gif webp tiff bmp heic avif'

# Quality flag
complete -c pixshift -s q -l quality -x -d 'Quality level'

# Jobs flag
complete -c pixshift -s j -l jobs -x -d 'Number of parallel jobs'

# Output directory flag
complete -c pixshift -s o -l output -r -F -d 'Output directory'

# Recursive flag
complete -c pixshift -s r -l recursive -d 'Process directories recursively'

# Preserve metadata flag
complete -c pixshift -s m -l preserve-metadata -d 'Preserve image metadata'

# Watch flag
complete -c pixshift -s w -l watch -d 'Watch for file changes'

# Config flag
complete -c pixshift -s c -l config -r -F -d 'Config file path'

# Overwrite flag
complete -c pixshift -l overwrite -d 'Overwrite existing files'

# Dry run flag
complete -c pixshift -l dry-run -d 'Show what would be done without doing it'

# Verbose flag
complete -c pixshift -s v -l verbose -d 'Enable verbose output'

# Version flag
complete -c pixshift -s V -l version -d 'Show version'

# Help flag
complete -c pixshift -s h -l help -d 'Show help'

# Width flag
complete -c pixshift -l width -x -d 'Resize width'

# Height flag
complete -c pixshift -l height -x -d 'Resize height'

# Max dimension flag
complete -c pixshift -l max-dim -x -d 'Maximum dimension'

# Strip metadata flag
complete -c pixshift -s s -l strip-metadata -d 'Strip image metadata'

# Template flag
complete -c pixshift -l template -x -d 'Output filename template'

# Completion flag
complete -c pixshift -l completion -x -d 'Generate shell completion' -a 'bash zsh fish'

# Positional arguments: file completion
complete -c pixshift -a '(__fish_complete_path)'
`
}
