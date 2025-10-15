package services

import (
	"errors"
	"fmt"

	"data-service/internal/models"
	"data-service/internal/repository"

	"gorm.io/gorm"
)

type DataSourceService struct {
	dataSourceRepo *repository.DataSourceRepository
}

func NewDataSourceService(dataSourceRepo *repository.DataSourceRepository) *DataSourceService {
	return &DataSourceService{
		dataSourceRepo: dataSourceRepo,
	}
}

func (s *DataSourceService) CreateDataSource(req *models.DataSourceCreateRequest) (*models.DataSourceResponse, error) {
	dataSource := &models.DataSource{
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Config:      req.Config,
		IsActive:    req.IsActive,
	}

	if err := s.dataSourceRepo.Create(dataSource); err != nil {
		return nil, fmt.Errorf("ошибка создания источника данных: %w", err)
	}

	response := dataSource.ToResponse()
	return &response, nil
}

func (s *DataSourceService) GetDataSources(page, limit int, active string) ([]models.DataSourceResponse, int64, error) {
	var isActive *bool
	if active != "" {
		activeBool := active == "true"
		isActive = &activeBool
	}

	dataSources, total, err := s.dataSourceRepo.GetAll(page, limit, isActive)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения источников данных: %w", err)
	}

	responses := make([]models.DataSourceResponse, len(dataSources))
	for i, ds := range dataSources {
		responses[i] = ds.ToResponse()
	}

	return responses, total, nil
}

func (s *DataSourceService) GetDataSource(id uint) (*models.DataSourceResponse, error) {
	dataSource, err := s.dataSourceRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("источник данных не найден")
		}
		return nil, fmt.Errorf("ошибка получения источника данных: %w", err)
	}

	response := dataSource.ToResponse()
	return &response, nil
}

func (s *DataSourceService) UpdateDataSource(id uint, req *models.DataSourceUpdateRequest) (*models.DataSourceResponse, error) {
	dataSource, err := s.dataSourceRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("источник данных не найден")
		}
		return nil, fmt.Errorf("ошибка получения источника данных: %w", err)
	}

	if req.Name != "" {
		dataSource.Name = req.Name
	}
	if req.Description != "" {
		dataSource.Description = req.Description
	}
	if req.Type != "" {
		dataSource.Type = req.Type
	}
	if req.Config != "" {
		dataSource.Config = req.Config
	}
	dataSource.IsActive = req.IsActive

	if err := s.dataSourceRepo.Update(dataSource); err != nil {
		return nil, fmt.Errorf("ошибка обновления источника данных: %w", err)
	}

	response := dataSource.ToResponse()
	return &response, nil
}

func (s *DataSourceService) DeleteDataSource(id uint) error {
	if err := s.dataSourceRepo.Delete(id); err != nil {
		return fmt.Errorf("ошибка удаления источника данных: %w", err)
	}
	return nil
}

type DataCollectionService struct {
	dataCollectionRepo *repository.DataCollectionRepository
}

func NewDataCollectionService(dataCollectionRepo *repository.DataCollectionRepository) *DataCollectionService {
	return &DataCollectionService{
		dataCollectionRepo: dataCollectionRepo,
	}
}

func (s *DataCollectionService) CreateDataCollection(req *models.DataCollectionCreateRequest) (*models.DataCollectionResponse, error) {
	dataCollection := &models.DataCollection{
		Name:         req.Name,
		Description:  req.Description,
		DataSourceID: req.DataSourceID,
		Query:        req.Query,
		Parameters:   req.Parameters,
		IsActive:     req.IsActive,
	}

	if err := s.dataCollectionRepo.Create(dataCollection); err != nil {
		return nil, fmt.Errorf("ошибка создания сбора данных: %w", err)
	}

	response := dataCollection.ToResponse()
	return &response, nil
}

func (s *DataCollectionService) GetDataCollections(page, limit int, active string) ([]models.DataCollectionResponse, int64, error) {
	var isActive *bool
	if active != "" {
		activeBool := active == "true"
		isActive = &activeBool
	}

	dataCollections, total, err := s.dataCollectionRepo.GetAll(page, limit, isActive)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения сборов данных: %w", err)
	}

	responses := make([]models.DataCollectionResponse, len(dataCollections))
	for i, dc := range dataCollections {
		responses[i] = dc.ToResponse()
	}

	return responses, total, nil
}

func (s *DataCollectionService) GetDataCollection(id uint) (*models.DataCollectionResponse, error) {
	dataCollection, err := s.dataCollectionRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("сбор данных не найден")
		}
		return nil, fmt.Errorf("ошибка получения сбора данных: %w", err)
	}

	response := dataCollection.ToResponse()
	return &response, nil
}

func (s *DataCollectionService) UpdateDataCollection(id uint, req *models.DataCollectionUpdateRequest) (*models.DataCollectionResponse, error) {
	dataCollection, err := s.dataCollectionRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("сбор данных не найден")
		}
		return nil, fmt.Errorf("ошибка получения сбора данных: %w", err)
	}

	if req.Name != "" {
		dataCollection.Name = req.Name
	}
	if req.Description != "" {
		dataCollection.Description = req.Description
	}
	if req.DataSourceID != 0 {
		dataCollection.DataSourceID = req.DataSourceID
	}
	if req.Query != "" {
		dataCollection.Query = req.Query
	}
	if req.Parameters != "" {
		dataCollection.Parameters = req.Parameters
	}
	dataCollection.IsActive = req.IsActive

	if err := s.dataCollectionRepo.Update(dataCollection); err != nil {
		return nil, fmt.Errorf("ошибка обновления сбора данных: %w", err)
	}

	response := dataCollection.ToResponse()
	return &response, nil
}

func (s *DataCollectionService) DeleteDataCollection(id uint) error {
	if err := s.dataCollectionRepo.Delete(id); err != nil {
		return fmt.Errorf("ошибка удаления сбора данных: %w", err)
	}
	return nil
}

type CollectDataService struct {
	dataRecordRepo *repository.DataRecordRepository
}

func NewCollectDataService(dataRecordRepo *repository.DataRecordRepository) *CollectDataService {
	return &CollectDataService{
		dataRecordRepo: dataRecordRepo,
	}
}

func (s *CollectDataService) CollectData(req *models.DataCollectRequest) (*models.CollectDataResponse, error) {
	dataRecord := &models.DataRecord{
		CollectionID: req.CollectionID,
		Data:         `{"collected": true, "timestamp": "2024-01-01T00:00:00Z"}`,
		Metadata:     `{"source": "simulation", "parameters": "test"}`,
	}

	if err := s.dataRecordRepo.Create(dataRecord); err != nil {
		return nil, fmt.Errorf("ошибка создания записи данных: %w", err)
	}

	return &models.CollectDataResponse{
		RecordsCollected: 1,
		Message:          "Данные успешно собраны",
	}, nil
}

func (s *CollectDataService) GetDataRecords(page, limit int, collectionID uint) ([]models.DataRecordResponse, int64, error) {
	dataRecords, total, err := s.dataRecordRepo.GetAll(page, limit, collectionID)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения записей данных: %w", err)
	}

	responses := make([]models.DataRecordResponse, len(dataRecords))
	for i, dr := range dataRecords {
		responses[i] = dr.ToResponse()
	}

	return responses, total, nil
}

func (s *CollectDataService) GetDataRecord(id uint) (*models.DataRecordResponse, error) {
	dataRecord, err := s.dataRecordRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("запись данных не найдена")
		}
		return nil, fmt.Errorf("ошибка получения записи данных: %w", err)
	}

	response := dataRecord.ToResponse()
	return &response, nil
}
