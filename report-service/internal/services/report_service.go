package services

import (
	"errors"
	"fmt"

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
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	reports, total, err := s.reportRepo.GetReportsWithPagination(page, limit, userID, status)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения отчетов: %w", err)
	}

	var reportResponses []models.ReportResponse
	for _, report := range reports {
		reportResponses = append(reportResponses, report.ToResponse())
	}

	return &models.ReportsResponse{
		Reports: reportResponses,
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
		if !models.ReportStatus(req.Status).IsValid() {
			return nil, errors.New("недопустимый статус отчета")
		}
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
	// Проверяем, существует ли отчет и принадлежит ли пользователю
	report, err := s.reportRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("отчет не найден")
		}
		return fmt.Errorf("ошибка получения отчета: %w", err)
	}

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
	if !models.ReportStatus(status).IsValid() {
		return errors.New("недопустимый статус отчета")
	}

	if err := s.reportRepo.UpdateStatus(id, status); err != nil {
		return fmt.Errorf("ошибка обновления статуса отчета: %w", err)
	}

	return nil
}

// UpdateReportFilePath обновляет путь к файлу отчета
func (s *ReportService) UpdateReportFilePath(id uint, filePath string, fileSize int64, md5Hash string) error {
	if err := s.reportRepo.UpdateFilePath(id, filePath, fileSize, md5Hash); err != nil {
		return fmt.Errorf("ошибка обновления пути к файлу отчета: %w", err)
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

	// Обновляем статус на "processing"
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
