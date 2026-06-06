package file

import (
	"errors"
	"net/http"
	"strings"

	"dotblue/internal/domains/identity"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func UploadHandler(r *ghttp.Request) {
	userID := identity.GetUserId(r)
	groupID := identity.GetCurrentEnterpriseId(r)
	if userID == "" || groupID == "" {
		r.Response.WriteStatus(http.StatusUnauthorized, "User context not found")
		return
	}

	upload := r.GetUploadFile("file")
	if upload == nil {
		r.Response.WriteStatus(http.StatusBadRequest, "file is required")
		return
	}
	fileReader, err := upload.Open()
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, "failed to read upload")
		return
	}
	defer fileReader.Close()

	readSeeker, ok := fileReader.(interface {
		Read([]byte) (int, error)
		Seek(offset int64, whence int) (int64, error)
	})
	if !ok {
		r.Response.WriteStatus(http.StatusBadRequest, "upload stream is not seekable")
		return
	}

	record, err := defaultService.Upload(r.Context(), UploadInput{
		UserID:         userID,
		GroupID:        groupID,
		ConversationID: strings.TrimSpace(r.Get("conversationId").String()),
		OriginalName:   upload.Filename,
		MimeType:       upload.Header.Get("Content-Type"),
		SizeBytes:      upload.Size,
		Kind:           r.Get("kind").String(),
		Content:        readSeeker,
	})
	if err != nil {
		g.Log().Errorf(
			r.Context(),
			"file upload failed user=%s group=%s conversation=%s filename=%q size=%d kind=%s: %v",
			userID,
			groupID,
			strings.TrimSpace(r.Get("conversationId").String()),
			upload.Filename,
			upload.Size,
			r.Get("kind").String(),
			err,
		)
		writeFileError(r, err)
		return
	}

	r.Response.WriteJson(toPublic(record))
}

func GetHandler(r *ghttp.Request) {
	userID := identity.GetUserId(r)
	groupID := identity.GetCurrentEnterpriseId(r)
	if userID == "" || groupID == "" {
		r.Response.WriteStatus(http.StatusUnauthorized, "User context not found")
		return
	}
	record, err := defaultService.GetPublicForUser(r.Context(), r.Get("id").String(), userID, groupID)
	if err != nil {
		writeFileError(r, err)
		return
	}
	r.Response.WriteJson(record)
}

func PreviewHandler(r *ghttp.Request) {
	openAndServe(r, false)
}

func DownloadHandler(r *ghttp.Request) {
	openAndServe(r, true)
}

func openAndServe(r *ghttp.Request, download bool) {
	userID := identity.GetUserId(r)
	groupID := identity.GetCurrentEnterpriseId(r)
	if userID == "" || groupID == "" {
		r.Response.WriteStatus(http.StatusUnauthorized, "User context not found")
		return
	}
	opened, err := defaultService.OpenForUser(r.Context(), r.Get("id").String(), userID, groupID)
	if err != nil {
		writeFileError(r, err)
		return
	}
	defer opened.Content.Close()

	if download {
		r.Response.Header().Set("Content-Disposition", `attachment; filename="`+sanitizeDownloadFilename(opened.File.OriginName)+`"`)
	}
	if opened.File.MimeType != "" {
		r.Response.Header().Set("Content-Type", opened.File.MimeType)
	}
	r.Response.ServeContent(opened.File.OriginName, opened.File.CreatedAt, opened.Content)
}

func sanitizeDownloadFilename(name string) string {
	name = strings.ReplaceAll(name, `"`, "")
	name = strings.ReplaceAll(name, "\r", "")
	name = strings.ReplaceAll(name, "\n", "")
	if strings.TrimSpace(name) == "" {
		return "download"
	}
	return name
}

func writeFileError(r *ghttp.Request, err error) {
	switch {
	case errors.Is(err, ErrFileNotFound):
		r.Response.WriteStatus(http.StatusNotFound, "File not found")
	case errors.Is(err, ErrFileAccessDenied):
		r.Response.WriteStatus(http.StatusForbidden, "File access denied")
	case errors.Is(err, ErrFileTooLarge):
		r.Response.WriteStatus(http.StatusRequestEntityTooLarge, "File too large")
	case errors.Is(err, ErrInvalidFileType), errors.Is(err, ErrInvalidFileUpload):
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
	default:
		r.Response.WriteStatus(http.StatusInternalServerError, err.Error())
	}
}
