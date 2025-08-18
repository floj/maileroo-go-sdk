package maileroo

import (
	"encoding/base64"
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var extToMime = map[string]string{
	"png":   "image/png",
	"jpg":   "image/jpeg",
	"jpeg":  "image/jpeg",
	"gif":   "image/gif",
	"bmp":   "image/bmp",
	"webp":  "image/webp",
	"svg":   "image/svg+xml",
	"tiff":  "image/tiff",
	"ico":   "image/x-icon",
	"pdf":   "application/pdf",
	"doc":   "application/msword",
	"docx":  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"xls":   "application/vnd.ms-excel",
	"xlsx":  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	"ppt":   "application/vnd.ms-powerpoint",
	"pptx":  "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	"odt":   "application/vnd.oasis.opendocument.text",
	"ods":   "application/vnd.oasis.opendocument.spreadsheet",
	"odp":   "application/vnd.oasis.opendocument.presentation",
	"rtf":   "application/rtf",
	"txt":   "text/plain",
	"csv":   "text/csv",
	"tsv":   "text/tab-separated-values",
	"json":  "application/json",
	"xml":   "application/xml",
	"html":  "text/html",
	"htm":   "text/html",
	"md":    "text/markdown",
	"zip":   "application/zip",
	"tar":   "application/x-tar",
	"gz":    "application/gzip",
	"tgz":   "application/gzip",
	"rar":   "application/vnd.rar",
	"7z":    "application/x-7z-compressed",
	"mp3":   "audio/mpeg",
	"wav":   "audio/wav",
	"ogg":   "audio/ogg",
	"m4a":   "audio/mp4",
	"flac":  "audio/flac",
	"aac":   "audio/aac",
	"mp4":   "video/mp4",
	"webm":  "video/webm",
	"mov":   "video/quicktime",
	"avi":   "video/x-msvideo",
	"mkv":   "video/x-matroska",
	"flv":   "video/x-flv",
	"wmv":   "video/x-ms-wmv",
	"m4v":   "video/x-m4v",
	"woff":  "font/woff",
	"woff2": "font/woff2",
	"ttf":   "font/ttf",
	"otf":   "font/otf",
	"eot":   "application/vnd.ms-fontobject",
}

type Attachment struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	Content     string `json:"content"`
	Inline      bool   `json:"inline"`
}

func NewAttachment(fileName, contentB64 string, contentType string, inline bool) (*Attachment, error) {

	if strings.TrimSpace(fileName) == "" {
		return nil, errors.New("file_name is required")
	}

	if strings.TrimSpace(contentB64) == "" {
		return nil, errors.New("content must be a non-empty base64 string")
	}

	if _, err := base64.StdEncoding.DecodeString(contentB64); err != nil {
		return nil, errors.New("invalid base64 content provided")
	}

	ct := contentType

	if strings.TrimSpace(ct) == "" {
		ct = "application/octet-stream"
	}

	return &Attachment{
		FileName:    fileName,
		ContentType: ct,
		Content:     contentB64,
		Inline:      inline,
	}, nil

}

func AttachmentFromContent(fileName string, content []byte, contentType string, inline bool) (*Attachment, error) {

	if strings.TrimSpace(fileName) == "" {
		return nil, errors.New("file_name is required")
	}

	ct := contentType

	if strings.TrimSpace(ct) == "" {

		if d := detectMimeFromBuffer(content); d != "" {
			ct = d
		} else {
			ct = "application/octet-stream"
		}

	}

	b64 := base64.StdEncoding.EncodeToString(content)

	return &Attachment{
		FileName:    fileName,
		ContentType: ct,
		Content:     b64,
		Inline:      inline,
	}, nil

}

func AttachmentFromBase64Content(fileName, contentB64, contentType string, inline bool) (*Attachment, error) {

	raw, err := base64.StdEncoding.DecodeString(contentB64)

	if err != nil {
		return nil, errors.New("invalid base64 content provided")
	}

	ct := contentType

	if strings.TrimSpace(ct) == "" {

		if d := detectMimeFromBuffer(raw); d != "" {
			ct = d
		} else {
			ct = "application/octet-stream"
		}

	}

	return NewAttachment(fileName, contentB64, ct, inline)

}

func AttachmentFromStream(fileName string, r io.Reader, contentType string, inline bool) (*Attachment, error) {

	if r == nil {
		return nil, errors.New("stream must be a valid, non-nil reader")
	}

	data, err := io.ReadAll(r)

	if err != nil {
		return nil, errors.New("failed to read from stream")
	}

	return AttachmentFromContent(fileName, data, contentType, inline)

}

func AttachmentFromFile(path string, contentType string, inline bool) (*Attachment, error) {

	if strings.TrimSpace(path) == "" {
		return nil, errors.New("path must be a readable file")
	}

	info, err := os.Stat(path)

	if err != nil || info.IsDir() {
		return nil, errors.New("path must be a readable file")
	}

	data, err := os.ReadFile(path)

	if err != nil {
		return nil, errors.New("failed to read file: " + path)
	}

	fileName := filepath.Base(path)
	ct := contentType

	if strings.TrimSpace(ct) == "" {

		if d := detectMimeFromPath(path); d != "" {
			ct = d
		} else if d2 := detectMimeFromBuffer(data); d2 != "" {
			ct = d2
		} else {
			ct = "application/octet-stream"
		}

	}

	b64 := base64.StdEncoding.EncodeToString(data)

	return &Attachment{
		FileName:    fileName,
		ContentType: ct,
		Content:     b64,
		Inline:      inline,
	}, nil

}

func (a *Attachment) ToMap() map[string]any {

	ct := a.ContentType

	if strings.TrimSpace(ct) == "" {
		ct = "application/octet-stream"
	}

	return map[string]any{
		"file_name":    a.FileName,
		"content_type": ct,
		"content":      a.Content,
		"inline":       a.Inline,
	}

}

func detectMimeFromPath(path string) string {

	if mt := detectMimeFromExtension(path); mt != "" {
		return mt
	}

	ext := strings.ToLower(filepath.Ext(path))

	if ext != "" {

		if mt := mime.TypeByExtension(ext); mt != "" {

			if i := strings.IndexByte(mt, ';'); i > 0 {
				return strings.TrimSpace(mt[:i])
			}

			return mt

		}

	}

	return ""

}

func detectMimeFromExtension(path string) string {

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))

	if ext == "" {
		return ""
	}

	if mt, ok := extToMime[ext]; ok {
		return mt
	}

	return ""

}

func detectMimeFromBuffer(buf []byte) string {

	mt := http.DetectContentType(buf)

	if mt == "" {
		return ""
	}

	if i := strings.IndexByte(mt, ';'); i > 0 {
		return strings.TrimSpace(mt[:i])
	}

	return mt

}

func (a *Attachment) validate() error {

	if strings.TrimSpace(a.FileName) == "" {
		return errors.New("attachment.file_name is required")
	}

	if strings.TrimSpace(a.Content) == "" {
		return errors.New("attachment.content_base64 must be a non-empty base64 string")
	}

	if strings.TrimSpace(a.ContentType) == "" {
		return errors.New("attachment.content_type is required")
	}

	return nil

}
