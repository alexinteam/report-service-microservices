package repository

import (
	"report-service/internal/models"

	"gorm.io/gorm"
)

// ReportRepository репозиторий для работы с отчетами
type ReportRepository struct {
	db *gorm.DB
}

// NewReportRepository создает новый репозиторий отчетов
func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// Create создает новый отчет
func (r *ReportRepository) Create(report *models.Report) error {
	return r.db.Create(report).Error
}

// GetByID получает отчет по ID
func (r *ReportRepository) GetByID(id uint) (*models.Report, error) {
	var report models.Report
	err := r.db.First(&report, id).Error
	return &report, err
}

// GetAll получает все отчеты с пагинацией
func (r *ReportRepository) GetAll(page, limit int, status string) ([]models.Report, int64, error) {
	var reports []models.Report
	var total int64

	query := r.db.Model(&models.Report{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Подсчитываем общее количество
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Получаем данные с пагинацией
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&reports).Error
	return reports, total, err
}

// GetReportsWithPagination получает отчеты с пагинацией
func (r *ReportRepository) GetReportsWithPagination(page, limit int, userID uint, status string) ([]models.Report, int64, error) {
	var reports []models.Report
	var total int64

	query := r.db.Model(&models.Report{})
	if userID != 0 {
		query = query.Where("user_id = ?", userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Подсчитываем общее количество
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Получаем данные с пагинацией
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&reports).Error
	return reports, total, err
}

// UpdateFilePath обновляет путь к файлу отчета
func (r *ReportRepository) UpdateFilePath(id uint, filePath string, fileSize int64, md5Hash string) error {
	return r.db.Model(&models.Report{}).Where("id = ?", id).Updates(map[string]interface{}{
		"file_path": filePath,
		"file_size": fileSize,
		"md5_hash":  md5Hash,
	}).Error
}

// UpdateStatus обновляет статус отчета
func (r *ReportRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&models.Report{}).Where("id = ?", id).Update("status", status).Error
}

// Update обновляет отчет
func (r *ReportRepository) Update(report *models.Report) error {
	return r.db.Save(report).Error
}

// Delete удаляет отчет (мягкое удаление)
func (r *ReportRepository) Delete(id uint) error {
	return r.db.Delete(&models.Report{}, id).Error
}
