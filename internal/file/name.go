package file

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

func FormatName(
	format string,
	creationDate time.Time,
	index int,
	originalPath string,
) string {
	if !strings.Contains(format, "%ext%") {
		format += "%ext%"
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
		"%index%", fmt.Sprintf("%d", index),
		"%type%", mimeTypeShort,
		"%mime-type%", mimeType,
		"%ext%", ext,
	)

	return replacer.Replace(format)
}
