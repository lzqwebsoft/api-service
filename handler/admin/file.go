package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"api-service/handler"
	"api-service/models"
	"api-service/service"
)

// FileHandler handles admin operations for app file releases and download logs
type FileHandler struct {
	*handler.Router
	*BaseHandler
	fileService service.FileService
	appService  service.AppService
	adminAuth   func(http.Handler) http.Handler
}

// NewFileHandler creates a new FileHandler instance
func NewFileHandler(base *BaseHandler, fileService service.FileService, appService service.AppService, adminAuth func(http.Handler) http.Handler) *FileHandler {
	h := &FileHandler{
		BaseHandler: base,
		fileService: fileService,
		appService:  appService,
		adminAuth:   adminAuth,
	}
	h.Router = handler.NewRouter(h)
	return h
}

// InitRoutes registers admin file routes
func (h *FileHandler) InitRoutes() []handler.Route {
	mw := []func(http.Handler) http.Handler{h.adminAuth}
	return []handler.Route{
		{Method: http.MethodGet, Path: "/admin/files", Handler: h.handleListFiles, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/files/create", Handler: h.handleCreateFile, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/files/update", Handler: h.handleUpdateFile, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/files/delete", Handler: h.handleDeleteFile, Middlewares: mw},
		{Method: http.MethodPost, Path: "/admin/files/upload", Handler: h.handleUploadFile, Middlewares: mw},
		{Method: http.MethodGet, Path: "/admin/files/logs", Handler: h.handleListLogs, Middlewares: mw},
	}
}

// handleListFiles returns paginated app file list
func (h *FileHandler) handleListFiles(w http.ResponseWriter, r *http.Request) {
	currentStr := r.URL.Query().Get("current")
	sizeStr := r.URL.Query().Get("size")
	appRecordIDStr := r.URL.Query().Get("app_record_id")

	current := 1
	if c, err := strconv.Atoi(currentStr); err == nil && c > 0 {
		current = c
	}
	size := 20
	if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
		size = s
	}

	appRecordID := 0
	if id, err := strconv.Atoi(appRecordIDStr); err == nil && id > 0 {
		appRecordID = id
	}

	limit := size
	offset := (current - 1) * size

	list, total, err := h.fileService.ListFiles(r.Context(), appRecordID, limit, offset)
	if err != nil {
		h.SendError(w, r, 500, "加载文件列表失败: "+err.Error())
		return
	}

	res := map[string]interface{}{
		"list":  list,
		"total": total,
	}
	h.SendSuccess(w, r, "获取成功", res)
}

// handleCreateFile creates a new file release record
func (h *FileHandler) handleCreateFile(w http.ResponseWriter, r *http.Request) {
	var file models.AppFile
	if err := json.NewDecoder(r.Body).Decode(&file); err != nil {
		h.SendError(w, r, 400, "请求参数解析失败")
		return
	}

	if file.AppRecordID == 0 {
		h.SendError(w, r, 400, "请选择关联应用")
		return
	}
	if file.FileName == "" {
		h.SendError(w, r, 400, "请输入文件名")
		return
	}
	if file.DownloadURL == "" {
		h.SendError(w, r, 400, "请输入下载链接")
		return
	}

	file.IsActive = true
	if err := h.fileService.CreateFile(r.Context(), &file); err != nil {
		h.SendError(w, r, 500, "创建文件发布失败: "+err.Error())
		return
	}

	h.SendSuccess(w, r, "新建文件发布成功", file)
}

// handleUpdateFile updates an existing file release record
func (h *FileHandler) handleUpdateFile(w http.ResponseWriter, r *http.Request) {
	var file models.AppFile
	if err := json.NewDecoder(r.Body).Decode(&file); err != nil {
		h.SendError(w, r, 400, "请求参数解析失败")
		return
	}

	if file.ID == 0 {
		h.SendError(w, r, 400, "缺少文件ID")
		return
	}

	if err := h.fileService.UpdateFile(r.Context(), &file); err != nil {
		h.SendError(w, r, 500, "更新文件发布失败: "+err.Error())
		return
	}

	h.SendSuccess(w, r, "更新文件发布成功", file)
}

// handleDeleteFile deletes a file release record
func (h *FileHandler) handleDeleteFile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = r.ParseForm()
		if id, err := strconv.Atoi(r.FormValue("id")); err == nil {
			req.ID = id
		}
	}

	if req.ID == 0 {
		h.SendError(w, r, 400, "缺少文件ID")
		return
	}

	if err := h.fileService.DeleteFile(r.Context(), req.ID); err != nil {
		h.SendError(w, r, 500, "删除文件失败: "+err.Error())
		return
	}

	h.SendSuccess(w, r, "删除文件成功", nil)
}

// handleListLogs returns paginated download logs for a file or all files
func (h *FileHandler) handleListLogs(w http.ResponseWriter, r *http.Request) {
	currentStr := r.URL.Query().Get("current")
	sizeStr := r.URL.Query().Get("size")
	fileIDStr := r.URL.Query().Get("file_id")

	current := 1
	if c, err := strconv.Atoi(currentStr); err == nil && c > 0 {
		current = c
	}
	size := 20
	if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
		size = s
	}

	fileID := 0
	if id, err := strconv.Atoi(fileIDStr); err == nil && id > 0 {
		fileID = id
	}

	limit := size
	offset := (current - 1) * size

	list, total, err := h.fileService.ListDownloadLogs(r.Context(), fileID, limit, offset)
	if err != nil {
		h.SendError(w, r, 500, "加载下载日志失败: "+err.Error())
		return
	}

	res := map[string]interface{}{
		"list":  list,
		"total": total,
	}
	h.SendSuccess(w, r, "获取成功", res)
}

// handleUploadFile processes file upload to runtime/upload directory
func (h *FileHandler) handleUploadFile(w http.ResponseWriter, r *http.Request) {
	// Parse max 200MB file upload
	err := r.ParseMultipartForm(200 << 20)
	if err != nil {
		h.SendError(w, r, 400, "文件上传解析失败: "+err.Error())
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		h.SendError(w, r, 400, "获取上传文件失败: "+err.Error())
		return
	}
	defer file.Close()

	uploadDir := filepath.Join("runtimes", "uploads")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		h.SendError(w, r, 500, "创建上传目录失败: "+err.Error())
		return
	}

	savedFileName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(header.Filename))
	dstPath := filepath.Join(uploadDir, savedFileName)

	dst, err := os.Create(dstPath)
	if err != nil {
		h.SendError(w, r, 500, "保存文件失败: "+err.Error())
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		h.SendError(w, r, 500, "写入文件失败: "+err.Error())
		return
	}

	downloadURL := "local://runtimes/uploads/" + savedFileName

	res := map[string]interface{}{
		"file_name":    header.Filename,
		"file_size":    header.Size,
		"download_url": downloadURL,
	}
	h.SendSuccess(w, r, "文件上传成功", res)
}
