package service

import (
	"context"
	"errors"

	"api-service/models"
	"api-service/repository"
	logger "api-service/utils"
)

// FileService interface defines business logic for app files & download tracking
type FileService interface {
	CreateFile(ctx context.Context, file *models.AppFile) error
	UpdateFile(ctx context.Context, file *models.AppFile) error
	DeleteFile(ctx context.Context, id int) error
	GetFileByID(ctx context.Context, id int) (*models.AppFile, error)
	ListFiles(ctx context.Context, appRecordID int, limit, offset int) ([]*models.AppFile, int, error)
	ProcessDownload(ctx context.Context, fileID int, ip, userAgent, referer, channel string) (*models.AppFile, error)
	ProcessLatestDownload(ctx context.Context, appID string, ip, userAgent, referer, channel string) (*models.AppFile, error)
	ListDownloadLogs(ctx context.Context, fileID int, limit, offset int) ([]*models.FileDownloadLog, int, error)
}

type fileService struct {
	fileRepo repository.FileRepository
	appRepo  repository.AppRepository
}

// NewFileService creates a FileService instance
func NewFileService(fileRepo repository.FileRepository, appRepo repository.AppRepository) FileService {
	return &fileService{
		fileRepo: fileRepo,
		appRepo:  appRepo,
	}
}

func (s *fileService) CreateFile(ctx context.Context, file *models.AppFile) error {
	if file.AppRecordID == 0 {
		return errors.New("应用记录ID不能为空")
	}
	if file.FileName == "" {
		return errors.New("文件名不能为空")
	}
	if file.DownloadURL == "" {
		return errors.New("下载链接不能为空")
	}
	return s.fileRepo.Create(ctx, file)
}

func (s *fileService) UpdateFile(ctx context.Context, file *models.AppFile) error {
	if file.ID == 0 {
		return errors.New("文件ID不能为空")
	}
	if file.FileName == "" {
		return errors.New("文件名不能为空")
	}
	if file.DownloadURL == "" {
		return errors.New("下载链接不能为空")
	}
	return s.fileRepo.Update(ctx, file)
}

func (s *fileService) DeleteFile(ctx context.Context, id int) error {
	return s.fileRepo.Delete(ctx, id)
}

func (s *fileService) GetFileByID(ctx context.Context, id int) (*models.AppFile, error) {
	return s.fileRepo.GetByID(ctx, id)
}

func (s *fileService) ListFiles(ctx context.Context, appRecordID int, limit, offset int) ([]*models.AppFile, int, error) {
	return s.fileRepo.List(ctx, appRecordID, limit, offset)
}

func (s *fileService) ProcessDownload(ctx context.Context, fileID int, ip, userAgent, referer, channel string) (*models.AppFile, error) {
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, err
	}
	if file == nil || !file.IsActive {
		return nil, errors.New("文件不存在或已被禁用")
	}

	// Record download count & log asynchronously
	go func(fID int, clientIP, ua, ref, ch string) {
		bgCtx := context.Background()
		if err := s.fileRepo.IncrementDownloadCount(bgCtx, fID); err != nil {
			logger.Errorf("Failed to increment download count for file %d: %v", fID, err)
		}

		log := &models.FileDownloadLog{
			FileID:    fID,
			IP:        clientIP,
			UserAgent: ua,
			Referer:   ref,
			Channel:   ch,
		}
		if err := s.fileRepo.LogDownload(bgCtx, log); err != nil {
			logger.Errorf("Failed to log download for file %d: %v", fID, err)
		}
	}(fileID, ip, userAgent, referer, channel)

	return file, nil
}

func (s *fileService) ProcessLatestDownload(ctx context.Context, appID string, ip, userAgent, referer, channel string) (*models.AppFile, error) {
	file, err := s.fileRepo.GetLatestByAppID(ctx, appID)
	if err != nil {
		return nil, err
	}
	if file == nil {
		return nil, errors.New("该应用暂无可用下载文件")
	}

	return s.ProcessDownload(ctx, file.ID, ip, userAgent, referer, channel)
}

func (s *fileService) ListDownloadLogs(ctx context.Context, fileID int, limit, offset int) ([]*models.FileDownloadLog, int, error) {
	return s.fileRepo.ListDownloadLogs(ctx, fileID, limit, offset)
}
