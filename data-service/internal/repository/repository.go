package repository

import (
	"data-service/internal/models"

	"gorm.io/gorm"
)

type DataSourceRepository struct {
	db *gorm.DB
}

func NewDataSourceRepository(db *gorm.DB) *DataSourceRepository {
	return &DataSourceRepository{db: db}
}

func (r *DataSourceRepository) Create(dataSource *models.DataSource) error {
	return r.db.Create(dataSource).Error
}

func (r *DataSourceRepository) GetByID(id uint) (*models.DataSource, error) {
	var dataSource models.DataSource
	err := r.db.First(&dataSource, id).Error
	return &dataSource, err
}

func (r *DataSourceRepository) GetAll(page, limit int, isActive *bool) ([]models.DataSource, int64, error) {
	var dataSources []models.DataSource
	var total int64

	query := r.db.Model(&models.DataSource{})
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Find(&dataSources).Error
	return dataSources, total, err
}

func (r *DataSourceRepository) Update(dataSource *models.DataSource) error {
	return r.db.Save(dataSource).Error
}

func (r *DataSourceRepository) Delete(id uint) error {
	return r.db.Delete(&models.DataSource{}, id).Error
}

type DataCollectionRepository struct {
	db *gorm.DB
}

func NewDataCollectionRepository(db *gorm.DB) *DataCollectionRepository {
	return &DataCollectionRepository{db: db}
}

func (r *DataCollectionRepository) Create(dataCollection *models.DataCollection) error {
	return r.db.Create(dataCollection).Error
}

func (r *DataCollectionRepository) GetByID(id uint) (*models.DataCollection, error) {
	var dataCollection models.DataCollection
	err := r.db.First(&dataCollection, id).Error
	return &dataCollection, err
}

func (r *DataCollectionRepository) GetAll(page, limit int, isActive *bool) ([]models.DataCollection, int64, error) {
	var dataCollections []models.DataCollection
	var total int64

	query := r.db.Model(&models.DataCollection{})
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Find(&dataCollections).Error
	return dataCollections, total, err
}

func (r *DataCollectionRepository) Update(dataCollection *models.DataCollection) error {
	return r.db.Save(dataCollection).Error
}

func (r *DataCollectionRepository) Delete(id uint) error {
	return r.db.Delete(&models.DataCollection{}, id).Error
}

type DataRecordRepository struct {
	db *gorm.DB
}

func NewDataRecordRepository(db *gorm.DB) *DataRecordRepository {
	return &DataRecordRepository{db: db}
}

func (r *DataRecordRepository) Create(dataRecord *models.DataRecord) error {
	return r.db.Create(dataRecord).Error
}

func (r *DataRecordRepository) GetByID(id uint) (*models.DataRecord, error) {
	var dataRecord models.DataRecord
	err := r.db.First(&dataRecord, id).Error
	return &dataRecord, err
}

func (r *DataRecordRepository) GetAll(page, limit int, collectionID uint) ([]models.DataRecord, int64, error) {
	var dataRecords []models.DataRecord
	var total int64

	query := r.db.Model(&models.DataRecord{})
	if collectionID != 0 {
		query = query.Where("collection_id = ?", collectionID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&dataRecords).Error
	return dataRecords, total, err
}

func (r *DataRecordRepository) Update(dataRecord *models.DataRecord) error {
	return r.db.Save(dataRecord).Error
}

func (r *DataRecordRepository) Delete(id uint) error {
	return r.db.Delete(&models.DataRecord{}, id).Error
}
