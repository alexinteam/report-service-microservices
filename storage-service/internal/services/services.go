package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"storage-service/internal/models"
	"storage-service/internal/repository"

	"gorm.io/gorm"
)

type FileService struct {
	fileRepo    *repository.FileRepository
	storagePath string
}

func NewFileService(fileRepo *repository.FileRepository, storagePath string) *FileService {
	return &FileService{
		fileRepo:    fileRepo,
		storagePath: storagePath,
	}
}

// UploadFile загружает файл
func (s *FileService) UploadFile(req *models.FileUploadRequest, filename string, content []byte, hash string) (*models.FileUploadResponse, error) {
	existingFile, err := s.fileRepo.GetByHash(hash)
	if err == nil && existingFile != nil {
		return &models.FileUploadResponse{
			File:    existingFile.ToResponse(),
			Message: "Файл уже существует",
		}, nil
	}

	mimeType := s.getMimeType(filename)

	filePath := filepath.Join(s.storagePath, hash)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return nil, fmt.Errorf("ошибка создания директории: %w", err)
	}

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return nil, fmt.Errorf("ошибка сохранения файла: %w", err)
	}

	file := &models.File{
		Name:        req.Name,
		Path:        filePath,
		Size:        int64(len(content)),
		MimeType:    mimeType,
		Hash:        hash,
		Description: req.Description,
		IsPublic:    req.IsPublic,
	}

	if err := s.fileRepo.Create(file); err != nil {
		os.Remove(filePath)
		return nil, fmt.Errorf("ошибка создания записи файла: %w", err)
	}

	return &models.FileUploadResponse{
		File:    file.ToResponse(),
		Message: "Файл успешно загружен",
	}, nil
}

// GetFiles получает список файлов
func (s *FileService) GetFiles(page, limit int, public string) ([]models.FileResponse, int64, error) {
	var isPublic *bool
	if public != "" {
		publicBool := public == "true"
		isPublic = &publicBool
	}

	files, total, err := s.fileRepo.GetAll(page, limit, isPublic)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения файлов: %w", err)
	}

	responses := make([]models.FileResponse, len(files))
	for i, f := range files {
		responses[i] = f.ToResponse()
	}

	return responses, total, nil
}

// GetFile получает файл по ID
func (s *FileService) GetFile(id uint) (*models.FileResponse, error) {
	file, err := s.fileRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("файл не найден")
		}
		return nil, fmt.Errorf("ошибка получения файла: %w", err)
	}

	response := file.ToResponse()
	return &response, nil
}

// DownloadFile скачивает файл
func (s *FileService) DownloadFile(id uint) (*models.FileDownloadResponse, error) {
	file, err := s.fileRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("файл не найден")
		}
		return nil, fmt.Errorf("ошибка получения файла: %w", err)
	}

	content, err := os.ReadFile(file.Path)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения файла: %w", err)
	}

	return &models.FileDownloadResponse{
		File:    file.ToResponse(),
		Content: content,
	}, nil
}

// UpdateFile обновляет файл
func (s *FileService) UpdateFile(id uint, req *models.FileUpdateRequest) (*models.FileResponse, error) {
	file, err := s.fileRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("файл не найден")
		}
		return nil, fmt.Errorf("ошибка получения файла: %w", err)
	}

	if req.Name != "" {
		file.Name = req.Name
	}
	if req.Description != "" {
		file.Description = req.Description
	}
	file.IsPublic = req.IsPublic

	if err := s.fileRepo.Update(file); err != nil {
		return nil, fmt.Errorf("ошибка обновления файла: %w", err)
	}

	response := file.ToResponse()
	return &response, nil
}

// DeleteFile удаляет файл
func (s *FileService) DeleteFile(id uint) error {
	file, err := s.fileRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("файл не найден")
		}
		return fmt.Errorf("ошибка получения файла: %w", err)
	}

	if err := os.Remove(file.Path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("ошибка удаления файла с диска: %w", err)
	}

	if err := s.fileRepo.Delete(id); err != nil {
		return fmt.Errorf("ошибка удаления записи файла: %w", err)
	}

	return nil
}

// GetFileByHash получает файл по хешу
func (s *FileService) GetFileByHash(hash string) (*models.FileResponse, error) {
	file, err := s.fileRepo.GetByHash(hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("файл не найден")
		}
		return nil, fmt.Errorf("ошибка получения файла: %w", err)
	}

	response := file.ToResponse()
	return &response, nil
}

// GetStorageStats получает статистику хранилища
func (s *FileService) GetStorageStats() (*models.StorageStatsResponse, error) {
	stats, err := s.fileRepo.GetStorageStats()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения статистики: %w", err)
	}

	return stats, nil
}

// SearchFiles ищет файлы
func (s *FileService) SearchFiles(query string, page, limit int) ([]models.FileResponse, int64, error) {
	files, total, err := s.fileRepo.Search(query, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка поиска файлов: %w", err)
	}

	responses := make([]models.FileResponse, len(files))
	for i, f := range files {
		responses[i] = f.ToResponse()
	}

	return responses, total, nil
}

// getMimeType определяет MIME тип файла по расширению
func (s *FileService) getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".txt":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".mp3":
		return "audio/mpeg"
	case ".mp4":
		return "video/mp4"
	case ".avi":
		return "video/x-msvideo"
	case ".zip":
		return "application/zip"
	case ".rar":
		return "application/x-rar-compressed"
	case ".tar":
		return "application/x-tar"
	case ".gz":
		return "application/gzip"
	default:
		return "application/octet-stream"
	}
}
