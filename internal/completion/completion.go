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
        --preset)
            COMPREPLY=( $(compgen -W "web thumbnail print archive" -- "${cur}") )
            return 0
            ;;
        --crop-gravity)
            COMPREPLY=( $(compgen -W "center north south east west" -- "${cur}") )
            return 0
            ;;
        --watermark-pos)
            COMPREPLY=( $(compgen -W "bottom-right bottom-left top-right top-left center" -- "${cur}") )
            return 0
            ;;
        --interpolation)
            COMPREPLY=( $(compgen -W "nearest bilinear catmullrom" -- "${cur}") )
            return 0
            ;;
        --png-compression)
            COMPREPLY=( $(compgen -W "0 1 2 3" -- "${cur}") )
            return 0
            ;;
        -q|--quality|-j|--jobs|--width|--height|--max-dim)
            return 0
            ;;
        --sepia|--brightness|--contrast|--blur|--watermark-size|--watermark-opacity|--dedup-threshold|--contact-cols|--contact-size|--webp-method)
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
        --template|--crop|--crop-ratio|--watermark|--watermark-color|--watermark-bg)
            return 0
            ;;
    esac

    if [[ "${cur}" == --* ]]; then
        opts="--format --quality --jobs --output --recursive --preserve-metadata --watch --config --overwrite --dry-run --verbose --version --help --width --height --max-dim --strip-metadata --template --completion --auto-rotate --crop --crop-ratio --crop-gravity --watermark --watermark-pos --watermark-opacity --preset --backup --json --tree --dedup --dedup-threshold --ssim --contact-sheet --contact-cols --contact-size --grayscale --sepia --brightness --contrast --sharpen --blur --invert --progressive --png-compression --webp-method --lossless --watermark-size --watermark-color --watermark-bg --interpolation"
        COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
        return 0
    fi

    if [[ "${cur}" == -* ]]; then
        opts="-f -q -j -o -r -m -w -c -v -V -h -s --format --quality --jobs --output --recursive --preserve-metadata --watch --config --overwrite --dry-run --verbose --version --help --width --height --max-dim --strip-metadata --template --completion --auto-rotate --crop --crop-ratio --crop-gravity --watermark --watermark-pos --watermark-opacity --preset --backup --json --tree --dedup --dedup-threshold --ssim --contact-sheet --contact-cols --contact-size --grayscale --sepia --brightness --contrast --sharpen --blur --invert --progressive --png-compression --webp-method --lossless --watermark-size --watermark-color --watermark-bg --interpolation"
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
    local -a formats completions presets gravities positions interpolations
    formats=(jpg jpeg png gif webp tiff bmp heic avif)
    completions=(bash zsh fish)
    presets=(web thumbnail print archive)
    gravities=(center north south east west)
    positions=(bottom-right bottom-left top-right top-left center)
    interpolations=(nearest bilinear catmullrom)

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
        '--auto-rotate[auto-rotate images based on EXIF orientation]' \
        '--crop[crop to WxH dimensions]:dimensions:' \
        '--crop-ratio[crop to aspect ratio (e.g. 16\:9)]:ratio:' \
        '--crop-gravity[crop anchor point]:gravity:(${gravities})' \
        '--watermark[add text watermark]:text:' \
        '--watermark-pos[watermark position]:position:(${positions})' \
        '--watermark-opacity[watermark opacity (0.0-1.0)]:opacity:' \
        '--watermark-size[watermark font size]:size:' \
        '--watermark-color[watermark text color]:color:' \
        '--watermark-bg[watermark background color]:color:' \
        '--preset[apply a built-in preset]:preset:(${presets})' \
        '--backup[create backup before overwriting]' \
        '--json[output results as JSON]' \
        '--tree[display directory tree of image files]' \
        '--dedup[find duplicate images using perceptual hashing]' \
        '--dedup-threshold[hamming distance threshold for dedup]:threshold:' \
        '--ssim[compute structural similarity between two images]:file1:_files:file2:_files' \
        '--contact-sheet[generate a contact sheet]' \
        '--contact-cols[number of columns in contact sheet]:columns:' \
        '--contact-size[thumbnail size in contact sheet]:size:' \
        '--grayscale[convert image to grayscale]' \
        '--sepia[apply sepia tone filter]:intensity:' \
        '--brightness[adjust brightness]:value:' \
        '--contrast[adjust contrast]:value:' \
        '--sharpen[apply sharpen filter]' \
        '--blur[apply blur filter]:radius:' \
        '--invert[invert image colors]' \
        '--progressive[enable progressive encoding]' \
        '--png-compression[PNG compression level (0-3)]:level:(0 1 2 3)' \
        '--webp-method[WebP compression method (0-6)]:method:' \
        '--lossless[enable lossless encoding]' \
        '--interpolation[resize interpolation method]:method:(${interpolations})' \
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

# v0.3.0 flags

# Auto-rotate flag
complete -c pixshift -l auto-rotate -d 'Auto-rotate images based on EXIF orientation'

# Crop flag
complete -c pixshift -l crop -x -d 'Crop to WxH dimensions (e.g. 800x600)'

# Crop ratio flag
complete -c pixshift -l crop-ratio -x -d 'Crop to aspect ratio (e.g. 16:9)'

# Crop gravity flag
complete -c pixshift -l crop-gravity -x -d 'Crop anchor point' -a 'center north south east west'

# Watermark flag
complete -c pixshift -l watermark -x -d 'Add text watermark'

# Watermark position flag
complete -c pixshift -l watermark-pos -x -d 'Watermark position' -a 'bottom-right bottom-left top-right top-left center'

# Watermark opacity flag
complete -c pixshift -l watermark-opacity -x -d 'Watermark opacity (0.0-1.0)'

# Preset flag
complete -c pixshift -l preset -x -d 'Apply a built-in preset' -a 'web thumbnail print archive'

# Backup flag
complete -c pixshift -l backup -d 'Create backup before overwriting'

# JSON output flag
complete -c pixshift -l json -d 'Output results as JSON'

# Tree mode flag
complete -c pixshift -l tree -d 'Display directory tree of image files'

# Dedup flag
complete -c pixshift -l dedup -d 'Find duplicate images using perceptual hashing'

# Dedup threshold flag
complete -c pixshift -l dedup-threshold -x -d 'Hamming distance threshold for dedup'

# SSIM flag
complete -c pixshift -l ssim -r -F -d 'Compute structural similarity between two images'

# Contact sheet flag
complete -c pixshift -l contact-sheet -d 'Generate a contact sheet'

# Contact columns flag
complete -c pixshift -l contact-cols -x -d 'Number of columns in contact sheet'

# Contact size flag
complete -c pixshift -l contact-size -x -d 'Thumbnail size in contact sheet'

# v0.4.0 flags

# Grayscale flag
complete -c pixshift -l grayscale -d 'Convert image to grayscale'

# Sepia flag
complete -c pixshift -l sepia -x -d 'Apply sepia tone filter (intensity)'

# Brightness flag
complete -c pixshift -l brightness -x -d 'Adjust brightness'

# Contrast flag
complete -c pixshift -l contrast -x -d 'Adjust contrast'

# Sharpen flag
complete -c pixshift -l sharpen -d 'Apply sharpen filter'

# Blur flag
complete -c pixshift -l blur -x -d 'Apply blur filter (radius)'

# Invert flag
complete -c pixshift -l invert -d 'Invert image colors'

# Progressive flag
complete -c pixshift -l progressive -d 'Enable progressive encoding'

# PNG compression flag
complete -c pixshift -l png-compression -x -d 'PNG compression level (0-3)' -a '0 1 2 3'

# WebP method flag
complete -c pixshift -l webp-method -x -d 'WebP compression method (0-6)'

# Lossless flag
complete -c pixshift -l lossless -d 'Enable lossless encoding'

# Watermark size flag
complete -c pixshift -l watermark-size -x -d 'Watermark font size'

# Watermark color flag
complete -c pixshift -l watermark-color -x -d 'Watermark text color'

# Watermark background flag
complete -c pixshift -l watermark-bg -x -d 'Watermark background color'

# Interpolation flag
complete -c pixshift -l interpolation -x -d 'Resize interpolation method' -a 'nearest bilinear catmullrom'

# Positional arguments: file completion
complete -c pixshift -a '(__fish_complete_path)'
`
}
