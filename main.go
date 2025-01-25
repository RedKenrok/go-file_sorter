package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	exif "github.com/dsoprea/go-exif-knife"
)

func main() {
	help := flag.Bool("help", false, "Show detailed help information")
	helpShort := flag.Bool("h", false, "Show detailed help information")

	output := flag.String("o", "", "Destination directory")
	outputLong := flag.String("output", "", "Destination directory")
	input := flag.String("i", ".", "Source directory")
	inputLong := flag.String("input", ".", "Source directory")
	format := flag.String("f", "%year%/%year%-%month%-%day%/%type%/file-%hour%_%minute%_%second%-%index%%ext%", "Path format for sorted files")
	formatLong := flag.String("format", "%year%/%year%-%month%-%day%/%type%/file-%hour%_%minute%_%second%-%index%%ext%", "Path format for sorted files")

	flag.Usage = func() {
		fmt.Println("Usage: file_sorter [options]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *help || *helpShort {
		showHelp()
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

	sourceDir := *input
	if sourceDir == "" {
		sourceDir = *inputLong
	}

	pathFormat := *format
	if pathFormat == "" {
		pathFormat = *formatLong
	}

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

			creationDate, err := getFileCreationDate(path)
			if err != nil {
				return fmt.Errorf("error getting creation date for file %s: %w", path, err)
			}

			count++
			newFileName := formatFileName(pathFormat, creationDate, count, path)
			destinationPath := filepath.Join(destDir, newFileName)

			// Ensure destination directory exists
			if err := os.MkdirAll(filepath.Dir(destinationPath), os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(destinationPath), err)
			}

			if err := copyFile(path, destinationPath); err != nil {
				return fmt.Errorf("failed to move file %s to %s: %w", path, destinationPath, err)
			}

			fmt.Printf("Moved file %s to %s\n", path, destinationPath)
			return nil
		},
	)

	if err != nil {
		fmt.Printf("Error processing files: %v\n", err)
	}
}

func showHelp() {
	fmt.Println("file_sorter: Organize and sort image files based on metadata")
	fmt.Println("\nUsage: file_sorter [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  -o, --output     Destination directory (required)")
	fmt.Println("  -i, --input      Source directory (default: current directory)")
	fmt.Println("  -f, --format     File path format (default: %year%/%year%-%month%-%day%/%type%/file-%hour%_%minute%_%second%-%index%%ext%)")
	fmt.Println("  -h, --help       Show this help message")
	fmt.Println("\nFormat Placeholders:")
	fmt.Println("  %year%    - 4-digit year")
	fmt.Println("  %month%   - 2-digit month")
	fmt.Println("  %day%     - 2-digit day")
	fmt.Println("  %hour%    - 2-digit hour (24-hour format)")
	fmt.Println("  %minute%  - 2-digit minute")
	fmt.Println("  %second%  - 2-digit second")
	fmt.Println("  %index%   - Incremental file index")
	fmt.Println("  %ext%     - File extension")
	fmt.Println("  %type%    - File type (image/video/audio/other)")
	fmt.Println("\nExample:")
	fmt.Println("  file_sorter -i /source -o /destination -f \"%year%/photos-%month%-%day%/image-%hour%_%minute%-%index%%ext%\"")
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
	mimeType := getMimeType(originalPath)
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

// Attempts to extract the creation date of a file.
func getFileCreationDate(
	filePath string,
) (
	time.Time,
	error,
) {
	file, err := os.Open(filePath)
	if err != nil {
		return time.Time{}, err
	}
	defer file.Close()

	mediaContext, err := exif.GetExif(filePath)
	if err == nil {
		time, err := getFileCreationDateFromExif(mediaContext)
		if err == nil {
			return time, nil
		}
	}

	// Fallback to file info modification time.
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return time.Time{}, err
	}

	return fileInfo.ModTime(), nil
}

func getFileCreationDateFromExif(
	mediaContext *exif.MediaContext,
) (
	time.Time,
	error,
) {
	if mediaContext == nil || mediaContext.RootIfd == nil {
		return time.Time{}, fmt.Errorf("no root IFD found")
	}

	dateTimeTags, err := mediaContext.RootIfd.FindTagWithName("DateTime")
	if err != nil {
		return time.Time{}, err
	}

	if len(dateTimeTags) > 0 {
		dateTimeValue, err := dateTimeTags[0].Value()
		if err != nil {
			return time.Time{}, err
		}

		dateTimeStr, ok := dateTimeValue.(string)
		if !ok {
			return time.Time{}, fmt.Errorf("invalid datetime format")
		}

		return time.Parse("2006:01:02 15:04:05", dateTimeStr)
	}

	return time.Time{}, fmt.Errorf("no datetime tag found")
}

func copyFile(
	path string,
	destinationPath string,
) error {
	sourceFile, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(destinationPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}
	return nil
}

func getMimeType(
	path string,
) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))

	switch ext {
	// Images
	case "jpg", "jpeg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "gif":
		return "image/gif"
	case "webp":
		return "image/webp"
	case "bmp":
		return "image/bmp"
	case "tiff", "tif":
		return "image/tiff"
	case "svg":
		return "image/svg+xml"
	case "heic", "heif":
		return "image/heic"
	case "raw":
		return "image/x-raw"
	case "avif":
		return "image/avif"
		// Raw image files from different cameras
	case "cr2", "cr3": // Canon
		return "image/x-canon-cr2"
	case "nef", "nrw": // Nikon
		return "image/x-nikon-nef"
	case "arw", "srf", "sr2": // Sony
		return "image/x-sony-arw"
	case "orf": // Olympus
		return "image/x-olympus-orf"
	case "rw2": // Panasonic
		return "image/x-panasonic-rw2"
	case "dng": // Adobe Digital Negative
		return "image/x-adobe-dng"

		// Video
	case "mp4":
		return "video/mp4"
	case "avi":
		return "video/x-msvideo"
	case "mov":
		return "video/quicktime"
	case "mkv":
		return "video/x-matroska"
	case "webm":
		return "video/webm"

		// Audio
	case "mp3":
		return "audio/mpeg"
	case "wav":
		return "audio/wav"
	case "flac":
		return "audio/flac"
	case "ogg":
		return "audio/ogg"
	case "m4a":
		return "audio/mp4"
	case "wma":
		return "audio/x-ms-wma"
	case "aac":
		return "audio/aac"
	case "mid", "midi":
		return "audio/midi"
	case "opus":
		return "audio/opus"

		// Documents and Markup
	case "pdf":
		return "application/pdf"
	case "doc":
		return "application/msword"
	case "docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "xls":
		return "application/vnd.ms-excel"
	case "xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "md", "markdown":
		return "text/markdown"
	case "rst":
		return "text/x-rst"
	case "txt":
		return "text/plain"
	case "rtf":
		return "application/rtf"
	case "html":
		return "text/html"

		// Code files
	case "js":
		return "application/javascript"
	case "py":
		return "text/x-python"
	case "go":
		return "text/x-go"
	case "rs":
		return "text/x-rust"
	case "cs":
		return "text/x-csharp"
	case "java":
		return "text/x-java-source"
	case "css":
		return "text/css"
	case "c":
		return "text/x-c"
	case "cpp":
		return "text/x-c++"

		// Configuration
	case "json":
		return "application/json"
	case "yaml", "yml":
		return "application/x-yaml"
	case "toml":
		return "application/toml"
	case "ini":
		return "text/ini"
	case "xml":
		return "application/xml"

		// Archives
	case "zip":
		return "application/zip"
	case "rar":
		return "application/x-rar-compressed"
	case "tar":
		return "application/x-tar"
	case "gz":
		return "application/gzip"

	default:
		return "application/octet-stream"
	}
}
