package models

import (
	"time"

	"gorm.io/gorm"
)

type DataSource struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Type        string         `json:"type" gorm:"not null"`    // database, api, file, etc.
	Config      string         `json:"config" gorm:"type:text"` // JSON конфигурация
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (DataSource) TableName() string {
	return "data_sources"
}

type DataCollection struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Name         string         `json:"name" gorm:"not null"`
	Description  string         `json:"description"`
	DataSourceID uint           `json:"data_source_id" gorm:"not null"`
	Query        string         `json:"query" gorm:"type:text"`
	Parameters   string         `json:"parameters" gorm:"type:text"` // JSON параметры
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

func (DataCollection) TableName() string {
	return "data_collections"
}

type DataRecord struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	CollectionID uint           `json:"collection_id" gorm:"not null"`
	Data         string         `json:"data" gorm:"type:text"`     // JSON данные
	Metadata     string         `json:"metadata" gorm:"type:text"` // JSON метаданные
	ProcessedAt  *time.Time     `json:"processed_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

func (DataRecord) TableName() string {
	return "data_records"
}

type DataSourceCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type" binding:"required"`
	Config      string `json:"config"`
	IsActive    bool   `json:"is_active"`
}

type DataSourceUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Config      string `json:"config"`
	IsActive    bool   `json:"is_active"`
}

type DataCollectionCreateRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	DataSourceID uint   `json:"data_source_id" binding:"required"`
	Query        string `json:"query"`
	Parameters   string `json:"parameters"`
	IsActive     bool   `json:"is_active"`
}

type DataCollectionUpdateRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	DataSourceID uint   `json:"data_source_id"`
	Query        string `json:"query"`
	Parameters   string `json:"parameters"`
	IsActive     bool   `json:"is_active"`
}

type DataCollectRequest struct {
	CollectionID uint                   `json:"collection_id" binding:"required"`
	Parameters   map[string]interface{} `json:"parameters"`
}

type DataSourceResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Config      string    `json:"config"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (ds *DataSource) ToResponse() DataSourceResponse {
	return DataSourceResponse{
		ID:          ds.ID,
		Name:        ds.Name,
		Description: ds.Description,
		Type:        ds.Type,
		Config:      ds.Config,
		IsActive:    ds.IsActive,
		CreatedAt:   ds.CreatedAt,
		UpdatedAt:   ds.UpdatedAt,
	}
}

type DataCollectionResponse struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	DataSourceID uint      `json:"data_source_id"`
	Query        string    `json:"query"`
	Parameters   string    `json:"parameters"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (dc *DataCollection) ToResponse() DataCollectionResponse {
	return DataCollectionResponse{
		ID:           dc.ID,
		Name:         dc.Name,
		Description:  dc.Description,
		DataSourceID: dc.DataSourceID,
		Query:        dc.Query,
		Parameters:   dc.Parameters,
		IsActive:     dc.IsActive,
		CreatedAt:    dc.CreatedAt,
		UpdatedAt:    dc.UpdatedAt,
	}
}

type DataRecordResponse struct {
	ID           uint       `json:"id"`
	CollectionID uint       `json:"collection_id"`
	Data         string     `json:"data"`
	Metadata     string     `json:"metadata"`
	ProcessedAt  *time.Time `json:"processed_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (dr *DataRecord) ToResponse() DataRecordResponse {
	return DataRecordResponse{
		ID:           dr.ID,
		CollectionID: dr.CollectionID,
		Data:         dr.Data,
		Metadata:     dr.Metadata,
		ProcessedAt:  dr.ProcessedAt,
		CreatedAt:    dr.CreatedAt,
		UpdatedAt:    dr.UpdatedAt,
	}
}

type DataSourcesResponse struct {
	DataSources []DataSourceResponse `json:"data_sources"`
	Total       int64                `json:"total"`
	Page        int                  `json:"page"`
	Limit       int                  `json:"limit"`
}

type DataCollectionsResponse struct {
	DataCollections []DataCollectionResponse `json:"data_collections"`
	Total           int64                    `json:"total"`
	Page            int                      `json:"page"`
	Limit           int                      `json:"limit"`
}

type DataRecordsResponse struct {
	DataRecords []DataRecordResponse `json:"data_records"`
	Total       int64                `json:"total"`
	Page        int                  `json:"page"`
	Limit       int                  `json:"limit"`
}

type CollectDataResponse struct {
	RecordsCollected int    `json:"records_collected"`
	Message          string `json:"message"`
}
