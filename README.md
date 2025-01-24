# Image sorter

Simply copies files in a directory of when it was created. For images it tries to read the date from the EXIF data, as a fallback the creation date of the file itself will be used. The file of the name will be the time followed by an iteration counter. An optional prefix can be applied to the file names as well.

The reason I created this program is because yet again my father managed to accidentally delete an entire hard drive of pictures. He managed to retrieve them, but the name an path of all the files was lost in the process. This application can be ran to put them back into a probably location using the file's date.

## Usage

```BASH
images_sorter <destination_directory> [source_directory] [file_name_prefix]
```

The source directory is optional, if no second argument is given it will assume the current working directory is the source directory. In addition the file name prefix is also optional.

## Build

Install `go` and `upx`. Run `bash ./build_windows.sh`.
