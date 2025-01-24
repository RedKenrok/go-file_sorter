package main

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	exif "github.com/dsoprea/go-exif-knife"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: image_sorter <destination_directory> [source_directory] [file_name_prefix]")
		os.Exit(1)
	}

	destDir := os.Args[1]
	sourceDir := "."
	if len(os.Args) > 2 {
		sourceDir = os.Args[2]
	}

	prefix := ""
	if len(os.Args) > 3 {
		prefix = fmt.Sprintf("%s-", os.Args[3])
	}

	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		fmt.Printf("Error creating destination directory: %v\n", err)
		os.Exit(1)
	}

	count := 0
	err := filepath.Walk(sourceDir, func(path string, info fs.FileInfo, err error) error {
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

		// Format date into year/month/day structure
		dirStructure := filepath.Join(destDir, creationDate.Format("2006"), creationDate.Format("01"), creationDate.Format("02"))

		if err := os.MkdirAll(dirStructure, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirStructure, err)
		}

		// Rename the file with prefix, creation time, and count.
		count++
		newFileName := fmt.Sprintf("%s%s-%d%s", prefix, creationDate.Format("15_04_05"), count, filepath.Ext(path))
		destinationPath := filepath.Join(dirStructure, newFileName)

		if err := copyFile(path, destinationPath); err != nil {
			return fmt.Errorf("failed to move file %s to %s: %w", path, destinationPath, err)
		}

		fmt.Printf("Moved file %s to %s\n", path, destinationPath)
		return nil
	})

	if err != nil {
		fmt.Printf("Error processing files: %v\n", err)
	}
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
