package services

import (
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"report-service/internal/models"
	"report-service/internal/repository"

	"gorm.io/gorm"
)

// ReportService сервис для работы с отчетами
type ReportService struct {
	reportRepo *repository.ReportRepository
}

// NewReportService создает новый сервис отчетов
func NewReportService(reportRepo *repository.ReportRepository) *ReportService {
	return &ReportService{
		reportRepo: reportRepo,
	}
}

// CreateReport создает новый отчет
func (s *ReportService) CreateReport(userID uint, req *models.ReportCreateRequest) (*models.ReportResponse, error) {
	// Проверяем валидность статуса
	if req.TemplateID == 0 {
		return nil, errors.New("ID шаблона обязателен")
	}

	// Создаем новый отчет
	report := &models.Report{
		Name:        req.Name,
		Description: req.Description,
		TemplateID:  req.TemplateID,
		UserID:      userID,
		Status:      string(models.StatusPending),
		Parameters:  req.Parameters,
	}

	if err := s.reportRepo.Create(report); err != nil {
		return nil, fmt.Errorf("ошибка создания отчета: %w", err)
	}

	response := report.ToResponse()
	return &response, nil
}

// GetReports получает список отчетов пользователя
func (s *ReportService) GetReports(userID uint, status string, page, limit int) (*models.ReportsResponse, error) {
	reports, total, err := s.reportRepo.GetAll(page, limit, status)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения отчетов: %w", err)
	}

	responses := make([]models.ReportResponse, len(reports))
	for i, report := range reports {
		responses[i] = report.ToResponse()
	}

	return &models.ReportsResponse{
		Reports: responses,
		Total:   total,
		Page:    page,
		Limit:   limit,
	}, nil
}

// GetReport получает отчет по ID
func (s *ReportService) GetReport(id uint, userID uint) (*models.ReportResponse, error) {
	report, err := s.reportRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("отчет не найден")
		}
		return nil, fmt.Errorf("ошибка получения отчета: %w", err)
	}

	// Проверяем, что отчет принадлежит пользователю
	if report.UserID != userID {
		return nil, errors.New("доступ запрещен")
	}

	response := report.ToResponse()
	return &response, nil
}

// UpdateReport обновляет отчет
func (s *ReportService) UpdateReport(id uint, userID uint, req *models.ReportUpdateRequest) (*models.ReportResponse, error) {
	report, err := s.reportRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("отчет не найден")
		}
		return nil, fmt.Errorf("ошибка получения отчета: %w", err)
	}

	// Проверяем, что отчет принадлежит пользователю
	if report.UserID != userID {
		return nil, errors.New("доступ запрещен")
	}

	// Обновляем поля
	if req.Name != "" {
		report.Name = req.Name
	}
	if req.Description != "" {
		report.Description = req.Description
	}
	if req.Status != "" {
		report.Status = req.Status
	}
	if req.Parameters != "" {
		report.Parameters = req.Parameters
	}

	if err := s.reportRepo.Update(report); err != nil {
		return nil, fmt.Errorf("ошибка обновления отчета: %w", err)
	}

	response := report.ToResponse()
	return &response, nil
}

// DeleteReport удаляет отчет
func (s *ReportService) DeleteReport(id uint, userID uint) error {
	report, err := s.reportRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("отчет не найден")
		}
		return fmt.Errorf("ошибка получения отчета: %w", err)
	}

	// Проверяем, что отчет принадлежит пользователю
	if report.UserID != userID {
		return errors.New("доступ запрещен")
	}

	if err := s.reportRepo.Delete(id); err != nil {
		return fmt.Errorf("ошибка удаления отчета: %w", err)
	}

	return nil
}

// UpdateReportStatus обновляет статус отчета
func (s *ReportService) UpdateReportStatus(id uint, status string) error {
	if err := s.reportRepo.UpdateStatus(id, status); err != nil {
		return fmt.Errorf("ошибка обновления статуса: %w", err)
	}
	return nil
}

// UpdateReportFilePath обновляет путь к файлу отчета
func (s *ReportService) UpdateReportFilePath(id uint, filePath string, fileSize int64, md5Hash string) error {
	if err := s.reportRepo.UpdateFilePath(id, filePath, fileSize, md5Hash); err != nil {
		return fmt.Errorf("ошибка обновления пути к файлу: %w", err)
	}
	return nil
}

// GenerateReport генерирует отчет
func (s *ReportService) GenerateReport(id uint, userID uint, req *models.ReportGenerateRequest) (*models.ReportResponse, error) {
	report, err := s.reportRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("отчет не найден")
		}
		return nil, fmt.Errorf("ошибка получения отчета: %w", err)
	}

	// Проверяем, что отчет принадлежит пользователю
	if report.UserID != userID {
		return nil, errors.New("доступ запрещен")
	}

	// Обновляем статус на processing
	if err := s.reportRepo.UpdateStatus(id, string(models.StatusProcessing)); err != nil {
		return nil, fmt.Errorf("ошибка обновления статуса: %w", err)
	}

	// Здесь должна быть логика генерации отчета
	// Пока просто возвращаем обновленный отчет
	report.Status = string(models.StatusProcessing)
	response := report.ToResponse()
	return &response, nil
}

// DownloadReport возвращает информацию для скачивания отчета
func (s *ReportService) DownloadReport(id uint, userID uint) (*models.ReportResponse, error) {
	report, err := s.reportRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("отчет не найден")
		}
		return nil, fmt.Errorf("ошибка получения отчета: %w", err)
	}

	// Проверяем, что отчет принадлежит пользователю
	if report.UserID != userID {
		return nil, errors.New("доступ запрещен")
	}

	// Проверяем, что отчет готов
	if report.Status != string(models.StatusCompleted) {
		return nil, errors.New("отчет еще не готов")
	}

	response := report.ToResponse()
	return &response, nil
}

// ExportReportToCSV экспортирует отчет в формат CSV
func (s *ReportService) ExportReportToCSV(id uint, userID uint) (string, error) {
	report, err := s.reportRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("отчет не найден")
		}
		return "", fmt.Errorf("ошибка получения отчета: %w", err)
	}

	// Проверяем, что отчет принадлежит пользователю
	if report.UserID != userID {
		return "", errors.New("доступ запрещен")
	}

	// Проверяем, что отчет готов
	if report.Status != string(models.StatusCompleted) {
		return "", errors.New("отчет еще не готов")
	}

	// Создаем CSV данные
	var csvData strings.Builder
	writer := csv.NewWriter(&csvData)

	// Заголовки CSV
	headers := []string{
		"ID",
		"Name",
		"Description",
		"Template ID",
		"User ID",
		"Status",
		"Parameters",
		"File Path",
		"File Size",
		"MD5 Hash",
		"Created At",
		"Updated At",
	}

	if err := writer.Write(headers); err != nil {
		return "", fmt.Errorf("ошибка записи заголовков CSV: %w", err)
	}

	// Данные отчета
	record := []string{
		strconv.FormatUint(uint64(report.ID), 10),
		report.Name,
		report.Description,
		strconv.FormatUint(uint64(report.TemplateID), 10),
		strconv.FormatUint(uint64(report.UserID), 10),
		report.Status,
		report.Parameters,
		report.FilePath,
		strconv.FormatInt(report.FileSize, 10),
		report.MD5Hash,
		report.CreatedAt.Format(time.RFC3339),
		report.UpdatedAt.Format(time.RFC3339),
	}

	if err := writer.Write(record); err != nil {
		return "", fmt.Errorf("ошибка записи данных CSV: %w", err)
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("ошибка записи CSV: %w", err)
	}

	return csvData.String(), nil
}
