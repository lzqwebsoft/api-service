package api

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"api-service/handler"
	"api-service/models"
	"api-service/service"
)

// DownloadHandler handles public download redirection and tracking
type DownloadHandler struct {
	*handler.Router
	*BaseHandler
	fileService service.FileService
}

// NewDownloadHandler creates a new DownloadHandler instance
func NewDownloadHandler(base *BaseHandler, fileService service.FileService) *DownloadHandler {
	h := &DownloadHandler{
		BaseHandler: base,
		fileService: fileService,
	}
	h.Router = handler.NewRouter(h)
	return h
}

// InitRoutes registers public download routes
func (h *DownloadHandler) InitRoutes() []handler.Route {
	return []handler.Route{
		{Method: http.MethodGet, Path: "/api/v1/download", Handler: h.handleDownload},
	}
}

// handleDownload processes download tracking and redirects (302) or serves local files directly
func (h *DownloadHandler) handleDownload(w http.ResponseWriter, r *http.Request) {
	fileIDStr := r.URL.Query().Get("file_id")
	if fileIDStr == "" {
		fileIDStr = r.URL.Query().Get("id")
	}
	appID := r.URL.Query().Get("app_id")
	channel := r.URL.Query().Get("channel")

	clientIP := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")
	referer := r.Header.Get("Referer")

	var file *models.AppFile
	var err error

	if fileIDStr != "" {
		fileID, errConv := strconv.Atoi(fileIDStr)
		if errConv != nil || fileID <= 0 {
			h.JSONError(w, r, "无效的文件ID", 400)
			return
		}
		file, err = h.fileService.ProcessDownload(r.Context(), fileID, clientIP, userAgent, referer, channel)
		if err != nil {
			h.JSONError(w, r, "下载文件不存在或已被禁用: "+err.Error(), 404)
			return
		}
	} else if appID != "" {
		file, err = h.fileService.ProcessLatestDownload(r.Context(), appID, clientIP, userAgent, referer, channel)
		if err != nil {
			h.JSONError(w, r, "应用无可用下载版本: "+err.Error(), 404)
			return
		}
	} else {
		h.JSONError(w, r, "请提供 file_id 或 app_id 参数", 400)
		return
	}

	targetFileUrl := file.DownloadURL

	// Set cache control headers to prevent browsers from caching download responses
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Option 1: External URL / CDN URL -> 302 HTTP Redirect
	if strings.HasPrefix(targetFileUrl, "http://") || strings.HasPrefix(targetFileUrl, "https://") {
		http.Redirect(w, r, targetFileUrl, http.StatusFound)
		return
	}

	// Option 2: Local file on server (runtimes/uploads)
	localPath := targetFileUrl
	if after, ok := strings.CutPrefix(localPath, "local://"); ok {
		localPath = after
	}

	// Validate path security to prevent directory traversal
	cleanPath := filepath.Clean(localPath)
	expectedPrefix := filepath.Clean("runtimes/uploads")
	if !strings.HasPrefix(cleanPath, expectedPrefix) && !strings.HasPrefix(cleanPath, filepath.Clean("runtime/upload")) {
		h.JSONError(w, r, "非法的文件路径访问", 403)
		return
	}

	fileInfo, statErr := os.Stat(cleanPath)
	if os.IsNotExist(statErr) || fileInfo.IsDir() {
		h.JSONError(w, r, "本地文件不存在或已被删除", 404)
		return
	}

	// Serve local file for direct download with Content-Disposition header
	downloadFileName := file.FileName
	if downloadFileName == "" {
		downloadFileName = fileInfo.Name()
	}
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", url.QueryEscape(downloadFileName)))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	http.ServeFile(w, r, cleanPath)
}

// getClientIP extracts real client IP considering proxy headers
func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = r.RemoteAddr
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}
	}
	if strings.Contains(ip, ",") {
		ip = strings.Split(ip, ",")[0]
	}
	return strings.TrimSpace(ip)
}
