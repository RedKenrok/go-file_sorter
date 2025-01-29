package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	file "github.com/redkenrok/go-file_sorter/internal/file"
)

var (
	version   = "dev"
	commit    = "none"
	buildDate = "unknown"
)

func main() {
	help := flag.Bool("help", false, "Show detailed help information")
	helpShort := flag.Bool("h", false, "Show detailed help information")
	version := flag.Bool("version", false, "Show program version information")
	versionShort := flag.Bool("v", false, "Show program version information")

	dryRunLong := flag.Bool("dry-run", false, "Perform a dry run without moving or copying files")
	format := flag.String("f", "%year%/%year%-%month%-%day%/%type%/file-%hour%_%minute%_%second%-%index%%ext%", "Path format for sorted files")
	formatLong := flag.String("format", "%year%/%year%-%month%-%day%/%type%/file-%hour%_%minute%_%second%-%index%%ext%", "Path format for sorted files")
	input := flag.String("i", ".", "Source directory")
	inputLong := flag.String("input", ".", "Source directory")
	move := flag.Bool("m", false, "Move files instead of copying")
	moveLong := flag.Bool("move", false, "Move files instead of copying")
	output := flag.String("o", "", "Destination directory")
	outputLong := flag.String("output", "", "Destination directory")
	dryRun := flag.Bool("dr", false, "Perform a dry run without moving or copying files")

	flag.Usage = func() {
		fmt.Println("For usage run: file_sorter --help\n")
	}

	flag.Parse()

	if *help || *helpShort {
		showHelp()
		os.Exit(0)
	}
	if *version || *versionShort {
		showVersion()
		os.Exit(0)
	}

	destDir := *output
	if destDir == "" {
		destDir = *outputLong
	}
	if destDir == "" {
		fmt.Println("Error: Output directory is required\n")
		flag.Usage()
		os.Exit(1)
	}
	destDir, _ = filepath.Abs(destDir)

	sourceDir := *input
	if sourceDir == "" {
		sourceDir = *inputLong
	}
	sourceDir, _ = filepath.Abs(sourceDir)

	pathFormat := *format
	if pathFormat == "" {
		pathFormat = *formatLong
	}

	doDryRun := *dryRun || *dryRunLong
	doMove := *move || *moveLong

	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		fmt.Printf("Error creating destination directory: %v\n", err)
		os.Exit(1)
	}

	count := 0
	err := filepath.Walk(
		sourceDir,
		func(
			path string,
			info fs.FileInfo,
			err error,
		) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			// Skip files in destination directory.
			absPath, _ := filepath.Abs(path)
			if strings.HasPrefix(absPath, destDir) {
				return nil
			}

			// Create new file path using creation date.
			creationDate, err := file.GetFileCreationDate(path)
			if err != nil {
				return fmt.Errorf("error getting creation date for file %s: %w", path, err)
			}
			count++
			newFileName := formatFileName(pathFormat, creationDate, count, path)
			destinationPath := filepath.Join(destDir, newFileName)

			if doDryRun {
				fmt.Printf("Dry run: Would %s file %s to %s\n",
					map[bool]string{true: "move", false: "copy"}[doMove],
					path, destinationPath)
				return nil
			}

			if err := os.MkdirAll(filepath.Dir(destinationPath), os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(destinationPath), err)
			}

			// Choose between move and copy.
			var transferErr error
			if doMove {
				transferErr = os.Rename(path, destinationPath)
			} else {
				transferErr = file.CopyFile(path, destinationPath)
			}

			if transferErr != nil {
				return fmt.Errorf("failed to %s file %s to %s: %w",
					map[bool]string{true: "move", false: "copy"}[doMove],
					path, destinationPath, transferErr)
			}

			fmt.Printf("%s file %s to %s\n",
				map[bool]string{true: "Moved", false: "Copied"}[doMove],
				path, destinationPath)
			return nil
		},
	)

	if err != nil {
		fmt.Printf("Error processing files: %v\n", err)
	}
}

func formatFileName(
	format string,
	creationDate time.Time,
	index int,
	originalPath string,
) string {
	if !strings.Contains(format, "%ext") {
		format += "%ext"
	}

	ext := filepath.Ext(originalPath)
	mimeType := file.GetMimeType(originalPath)
	mimeTypeShort := mimeType[strings.LastIndex(mimeType, "/")+1:]

	replacer := strings.NewReplacer(
		"%year%", creationDate.Format("2006"),
		"%month%", creationDate.Format("01"),
		"%day%", creationDate.Format("02"),
		"%hour%", creationDate.Format("15"),
		"%minute%", creationDate.Format("04"),
		"%second%", creationDate.Format("05"),
		"%index%", strconv.Itoa(index),
		"%type%", mimeTypeShort,
		"%mime-type%", mimeType,
		"%ext%", ext,
	)

	return replacer.Replace(format)
}

func showHelp() {
	fmt.Println("file_sorter: Organize and sort files based on its metadata.")
	fmt.Println("\nUsage: file_sorter [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -dr, --dry-run   Perform a dry run without actually moving or copying files, simply outputs what it would have done.")
	fmt.Println("  -f, --format     File path format (default: %year%/%year%-%month%-%day%/%type%/file-%hour%_%minute%_%second%-%index%%ext%).")
	fmt.Println("  -h, --help       Show detailed help information.")
	fmt.Println("  -i, --input      Input directory (default: current working directory).")
	fmt.Println("  -m, --move       Move files instead of copying, increased performance when on the same disk.")
	fmt.Println("  -o, --output     Output directory (required).")
	fmt.Println("  -v, --version    Show program version information.")
	fmt.Println("\nFormat Placeholders:")
	fmt.Println("  %year%      - 4-digit year")
	fmt.Println("  %month%     - 2-digit month")
	fmt.Println("  %day%       - 2-digit day")
	fmt.Println("  %hour%      - 2-digit hour (24-hour format)")
	fmt.Println("  %minute%    - 2-digit minute")
	fmt.Println("  %second%    - 2-digit second")
	fmt.Println("  %index%     - Incremental file index")
	fmt.Println("  %ext%       - File extension")
	fmt.Println("  %type%      - File type (flac,svg+xml,webm)")
	fmt.Println("  %mime-type% - File's mime-type (audio/flac,image/svg+xml,video/webm)")
	fmt.Println("\nExample:")
	fmt.Println("  file_sorter -i /source -o /destination -f \"%year%/%year%-%month%-%day%/file-%hour%_%minute%-%index%%ext%\"")
}

func showVersion() {
	fmt.Printf("Version: %s\n", version)
	fmt.Printf("Commit: %s\n", commit)
	fmt.Printf("Build Date: %s\n", buildDate)
}
