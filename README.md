# File sorter

Simply copies files in a directory of when it was created. For images it tries to read the date from the EXIF data, as a fallback the creation date of the file itself will be used. The file of the name will be the time followed by an iteration counter. An optional prefix can be applied to the file names as well.

It's simple, but does what it needs to do.

## Usage

```bash
file_sorter --help
```

## Build

- Install `go` and `upx` then run `bash ./run_build.sh`.
