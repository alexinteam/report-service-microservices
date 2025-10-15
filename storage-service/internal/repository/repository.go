package repository

import (
	"storage-service/internal/models"

	"gorm.io/gorm"
)

type FileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

// Create создает новый файл
func (r *FileRepository) Create(file *models.File) error {
	return r.db.Create(file).Error
}

// GetByID получает файл по ID
func (r *FileRepository) GetByID(id uint) (*models.File, error) {
	var file models.File
	err := r.db.First(&file, id).Error
	return &file, err
}

// GetByHash получает файл по хешу
func (r *FileRepository) GetByHash(hash string) (*models.File, error) {
	var file models.File
	err := r.db.Where("hash = ?", hash).First(&file).Error
	return &file, err
}

// GetAll получает все файлы с пагинацией
func (r *FileRepository) GetAll(page, limit int, isPublic *bool) ([]models.File, int64, error) {
	var files []models.File
	var total int64

	query := r.db.Model(&models.File{})
	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&files).Error
	return files, total, err
}

// Update обновляет файл
func (r *FileRepository) Update(file *models.File) error {
	return r.db.Save(file).Error
}

// Delete удаляет файл
func (r *FileRepository) Delete(id uint) error {
	return r.db.Delete(&models.File{}, id).Error
}

// Search ищет файлы по имени и описанию
func (r *FileRepository) Search(query string, page, limit int) ([]models.File, int64, error) {
	var files []models.File
	var total int64

	searchQuery := "%" + query + "%"
	queryBuilder := r.db.Model(&models.File{}).Where(
		"name ILIKE ? OR description ILIKE ?",
		searchQuery, searchQuery,
	)

	if err := queryBuilder.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := queryBuilder.Offset(offset).Limit(limit).Order("created_at DESC").Find(&files).Error
	return files, total, err
}

// GetStorageStats получает статистику хранилища
func (r *FileRepository) GetStorageStats() (*models.StorageStatsResponse, error) {
	var stats models.StorageStatsResponse

	if err := r.db.Model(&models.File{}).Count(&stats.TotalFiles).Error; err != nil {
		return nil, err
	}

	if err := r.db.Model(&models.File{}).Select("COALESCE(SUM(size), 0)").Scan(&stats.TotalSize).Error; err != nil {
		return nil, err
	}

	if err := r.db.Model(&models.File{}).Where("is_public = ?", true).Count(&stats.PublicFiles).Error; err != nil {
		return nil, err
	}

	if err := r.db.Model(&models.File{}).Where("is_public = ?", false).Count(&stats.PrivateFiles).Error; err != nil {
		return nil, err
	}

	if stats.TotalFiles > 0 {
		stats.AverageSize = stats.TotalSize / stats.TotalFiles
	}

	return &stats, nil
}
