package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"report-service/internal/events"
	"report-service/internal/metrics"
	"report-service/internal/models"
	"report-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ReportHandler обработчик для отчетов
type ReportHandler struct {
	reportService   *services.ReportService
	sagaCoordinator *events.IdempotentSagaCoordinator
	metrics         *metrics.Metrics
}

// NewReportHandler создает новый обработчик отчетов
func NewReportHandler(reportService *services.ReportService, sagaCoordinator *events.IdempotentSagaCoordinator, metrics *metrics.Metrics) *ReportHandler {
	return &ReportHandler{
		reportService:   reportService,
		sagaCoordinator: sagaCoordinator,
		metrics:         metrics,
	}
}

// CreateReport создание нового отчета через Saga (асинхронно)
func (h *ReportHandler) CreateReport(c *gin.Context) {
	start := time.Now()
	userID, exists := c.Get("user_id")
	if !exists {
		h.metrics.RecordBusinessOperation("report-service", "create_report", time.Since(start), false)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	var req models.ReportCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.metrics.RecordBusinessOperation("report-service", "create_report", time.Since(start), false)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Создаем отчет в статусе pending
	report, err := h.reportService.CreateReport(userID.(uint), &req)
	if err != nil {
		logrus.WithError(err).Error("Ошибка создания отчета")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Создаем идемпотентную Saga для генерации отчета
	saga := events.NewIdempotentReportCreationSaga(
		strconv.FormatUint(uint64(report.ID), 10),
		strconv.FormatUint(uint64(userID.(uint)), 10),
		strconv.FormatUint(uint64(req.TemplateID), 10),
		map[string]interface{}{
			"parameters":  req.Parameters,
			"name":        req.Name,
			"description": req.Description,
		},
	)

	// Запускаем Saga асинхронно
	go func() {
		ctx := context.Background()
		if err := saga.Execute(ctx, h.sagaCoordinator); err != nil {
			logrus.WithError(err).Errorf("Ошибка выполнения Saga создания отчета %s", saga.ID)
			// Обновляем статус отчета на failed
			h.reportService.UpdateReportStatus(report.ID, string(models.StatusFailed))
		}
	}()

	h.metrics.RecordBusinessOperation("report-service", "create_report", time.Since(start), true)
	c.JSON(http.StatusAccepted, models.ReportCreateResponse{
		ID:      report.ID,
		Status:  string(models.StatusPending),
		Message: "Отчет создан и поставлен в очередь на генерацию",
	})
}

// GetReports получение списка отчетов
func (h *ReportHandler) GetReports(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	// Получаем параметры запроса
	status := c.Query("status")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный параметр page"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный параметр limit"})
		return
	}

	reports, err := h.reportService.GetReports(userID.(uint), status, page, limit)
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения списка отчетов")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reports)
}

// GetReport получение отчета по ID
func (h *ReportHandler) GetReport(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID отчета"})
		return
	}

	report, err := h.reportService.GetReport(uint(id), userID.(uint))
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения отчета")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// GetReportStatus получение статуса отчета
func (h *ReportHandler) GetReportStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID отчета"})
		return
	}

	report, err := h.reportService.GetReport(uint(id), userID.(uint))
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения отчета")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	response := models.ReportStatusResponse{
		ID:     report.ID,
		Status: report.Status,
	}

	// Если отчет готов, добавляем путь к файлу
	if report.Status == string(models.StatusCompleted) && report.FilePath != "" {
		response.FilePath = report.FilePath
	}

	// Если отчет в процессе, добавляем прогресс (можно расширить логику)
	if report.Status == string(models.StatusProcessing) {
		response.Progress = 50 // Примерное значение, можно сделать более точным
	}

	c.JSON(http.StatusOK, response)
}

// UpdateReport обновление отчета
func (h *ReportHandler) UpdateReport(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID отчета"})
		return
	}

	var req models.ReportUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.reportService.UpdateReport(uint(id), userID.(uint), &req)
	if err != nil {
		logrus.WithError(err).Error("Ошибка обновления отчета")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// DeleteReport удаление отчета
func (h *ReportHandler) DeleteReport(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID отчета"})
		return
	}

	err = h.reportService.DeleteReport(uint(id), userID.(uint))
	if err != nil {
		logrus.WithError(err).Error("Ошибка удаления отчета")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Отчет успешно удален"})
}

// GenerateReport генерация отчета
func (h *ReportHandler) GenerateReport(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID отчета"})
		return
	}

	var req models.ReportGenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	report, err := h.reportService.GenerateReport(uint(id), userID.(uint), &req)
	if err != nil {
		logrus.WithError(err).Error("Ошибка генерации отчета")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Генерация отчета запущена", "report": report})
}

// DownloadReport скачивание отчета
func (h *ReportHandler) DownloadReport(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID отчета"})
		return
	}

	report, err := h.reportService.DownloadReport(uint(id), userID.(uint))
	if err != nil {
		logrus.WithError(err).Error("Ошибка скачивания отчета")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Отчет готов к скачиванию", "report": report})
}

// ExportReportCSV экспортирует отчет в формат CSV
func (h *ReportHandler) ExportReportCSV(c *gin.Context) {
	start := time.Now()
	defer func() {
		h.metrics.RecordBusinessOperation("report-service", "export_report_csv", time.Since(start), true)
	}()

	// Получаем ID пользователя из контекста
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID отчета"})
		return
	}

	csvData, err := h.reportService.ExportReportToCSV(uint(id), userID.(uint))
	if err != nil {
		logrus.WithError(err).Error("Ошибка экспорта отчета в CSV")
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Устанавливаем заголовки для скачивания CSV файла
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=report_"+idStr+".csv")
	c.String(http.StatusOK, csvData)
}
