# zipbomb
![Build Status](https://github.com/hupe1980/zipbomb/workflows/build/badge.svg) 
[![Go Reference](https://pkg.go.dev/badge/github.com/hupe1980/zipbomb.svg)](https://pkg.go.dev/github.com/hupe1980/zipbomb)
> Tool that creates zip bombs.

:warning: This is for educational purpose. Don’t try it on live clients/servers!

## Installing
You can install the pre-compiled binary in several different ways

### homebrew tap:
```bash
brew tap hupe1980/zipbomb
brew install zipbomb
```
### scoop:
```bash
scoop bucket add zipbomb https://github.com/hupe1980/zipbomb-bucket.git
scoop install zipbomb
```

### deb/rpm/apk:
Download the .deb, .rpm or .apk from the [releases page](https://github.com/hupe1980/zipbomb/releases) and install them with the appropriate tools.

### manually:
Download the pre-compiled binaries from the [releases page](https://github.com/hupe1980/zipbomb/releases) and copy to the desired location.


## How to use
```
Usage:
  zipbomb [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  overlap     Create non-recursive zipbomb

Flags:
  -h, --help            help for zipbomb
  -o, --output string   output filename (default "bomb.zip")
  -v, --version         version for zipbomb

Use "zipbomb [command] --help" for more information about a command.
```

### Overlap
Create non-recursive zipbomb that achieves a high compression ratio by overlapping files inside the zip container
```
Usage:
  zipbomb overlap [flags]

Flags:
      --alphabet string         alphabet for generating filenames (default "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
      --extension string        extension for generating filenames
  -h, --help                    help for overlap
  -B, --kernel-bytes bytesHex   kernel bytes (default 42)
  -R, --kernel-repeats int      kernel repeats (default 1048576)
  -N, --num-files int           number of files (default 100)
      --verify                  verify zip archive

Global Flags:
  -o, --output string   output filename (default "bomb.zip")
```

## License
[MIT](LICENCE)