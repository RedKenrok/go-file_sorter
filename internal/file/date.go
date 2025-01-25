package file

import (
	"fmt"
	"os"
	"time"

	exif "github.com/dsoprea/go-exif-knife"
)

// Attempts to extract the creation date of a file.
func GetFileCreationDate(
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
