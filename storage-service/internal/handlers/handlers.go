package handlers

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"storage-service/internal/metrics"
	"storage-service/internal/models"
	"storage-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type FileHandler struct {
	fileService *services.FileService
	metrics     *metrics.Metrics
}

func NewFileHandler(fileService *services.FileService, metrics *metrics.Metrics) *FileHandler {
	return &FileHandler{
		fileService: fileService,
		metrics:     metrics,
	}
}

// UploadFile загрузка файла
func (h *FileHandler) UploadFile(c *gin.Context) {
	start := time.Now()
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		h.metrics.RecordBusinessOperation("storage-service", "upload_file", time.Since(start), false)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл не найден"})
		return
	}
	defer file.Close()

	name := c.PostForm("name")
	if name == "" {
		name = header.Filename
	}
	description := c.PostForm("description")
	isPublic := c.PostForm("is_public") == "true"

	content, err := io.ReadAll(file)
	if err != nil {
		logrus.WithError(err).Error("Ошибка чтения файла")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения файла"})
		return
	}

	hash := fmt.Sprintf("%x", md5.Sum(content))

	req := &models.FileUploadRequest{
		Name:        name,
		Description: description,
		IsPublic:    isPublic,
	}

	result, err := h.fileService.UploadFile(req, header.Filename, content, hash)
	if err != nil {
		logrus.WithError(err).Error("Ошибка загрузки файла")
		h.metrics.RecordBusinessOperation("storage-service", "upload_file", time.Since(start), false)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.metrics.RecordBusinessOperation("storage-service", "upload_file", time.Since(start), true)
	c.JSON(http.StatusCreated, result)
}

// GetFiles получение списка файлов
func (h *FileHandler) GetFiles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	public := c.Query("public")

	files, total, err := h.fileService.GetFiles(page, limit, public)
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения файлов")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.FilesResponse{
		Files: files,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetFile получение файла по ID
func (h *FileHandler) GetFile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	file, err := h.fileService.GetFile(uint(id))
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения файла")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, file)
}

// DownloadFile скачивание файла
func (h *FileHandler) DownloadFile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	result, err := h.fileService.DownloadFile(uint(id))
	if err != nil {
		logrus.WithError(err).Error("Ошибка скачивания файла")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", result.File.Name))
	c.Header("Content-Type", result.File.MimeType)
	c.Header("Content-Length", strconv.FormatInt(result.File.Size, 10))

	c.Data(http.StatusOK, result.File.MimeType, result.Content)
}

// UpdateFile обновление файла
func (h *FileHandler) UpdateFile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	var req models.FileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, err := h.fileService.UpdateFile(uint(id), &req)
	if err != nil {
		logrus.WithError(err).Error("Ошибка обновления файла")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, file)
}

// DeleteFile удаление файла
func (h *FileHandler) DeleteFile(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	if err := h.fileService.DeleteFile(uint(id)); err != nil {
		logrus.WithError(err).Error("Ошибка удаления файла")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// GetFileByHash получение файла по хешу
func (h *FileHandler) GetFileByHash(c *gin.Context) {
	hash := c.Param("hash")
	if hash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Хеш не указан"})
		return
	}

	file, err := h.fileService.GetFileByHash(hash)
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения файла по хешу")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, file)
}

// GetStorageStats получение статистики хранилища
func (h *FileHandler) GetStorageStats(c *gin.Context) {
	stats, err := h.fileService.GetStorageStats()
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения статистики")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SearchFiles поиск файлов
func (h *FileHandler) SearchFiles(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Поисковый запрос не указан"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	files, total, err := h.fileService.SearchFiles(query, page, limit)
	if err != nil {
		logrus.WithError(err).Error("Ошибка поиска файлов")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.FilesResponse{
		Files: files,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

// GetFileContent получение содержимого файла
func (h *FileHandler) GetFileContent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	result, err := h.fileService.DownloadFile(uint(id))
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения содержимого файла")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", result.File.MimeType)
	c.Header("Content-Length", strconv.FormatInt(result.File.Size, 10))

	c.Data(http.StatusOK, result.File.MimeType, result.Content)
}
