package file

import (
	"path/filepath"
	"strings"
)

func GetMimeType(
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
