# File sorter

Simply copies files in a directory of when it was created. For images it tries to read the date from the EXIF data, as a fallback the creation date of the file itself will be used. The file of the name will be the time followed by an iteration counter. An optional prefix can be applied to the file names as well.

The reason I created this program is because yet again my father managed to accidentally deleted an entire hard drive of pictures, again. He managed to retrieve them, but the name an path of all the files was lost in the process. This application can be ran to put them back into a probably location using the file's date.

## Usage

```bash
file_sorter --help
```

## Build

- Install `go` and `upx` then run `bash ./run_build.sh`.
