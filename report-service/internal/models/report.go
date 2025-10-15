package models

import (
	"time"

	"gorm.io/gorm"
)

// Report модель отчета
type Report struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	TemplateID  uint           `json:"template_id" gorm:"not null"`
	UserID      uint           `json:"user_id" gorm:"not null"`
	Status      string         `json:"status" gorm:"default:'pending'"`
	Parameters  string         `json:"parameters" gorm:"type:text"`
	FilePath    string         `json:"file_path"`
	FileSize    int64          `json:"file_size"`
	MD5Hash     string         `json:"md5_hash"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName возвращает имя таблицы
func (Report) TableName() string {
	return "reports"
}

// ReportStatus статусы отчетов
type ReportStatus string

const (
	StatusPending    ReportStatus = "pending"
	StatusProcessing ReportStatus = "processing"
	StatusCompleted  ReportStatus = "completed"
	StatusFailed     ReportStatus = "failed"
	StatusCancelled  ReportStatus = "cancelled"
)

// IsValid проверяет валидность статуса отчета
func (s ReportStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusProcessing, StatusCompleted, StatusFailed, StatusCancelled:
		return true
	default:
		return false
	}
}

// ReportCreateRequest запрос на создание отчета
type ReportCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	TemplateID  uint   `json:"template_id" binding:"required"`
	Parameters  string `json:"parameters"`
}

// ReportUpdateRequest запрос на обновление отчета
type ReportUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Parameters  string `json:"parameters"`
}

// ReportGenerateRequest запрос на генерацию отчета
type ReportGenerateRequest struct {
	Parameters map[string]interface{} `json:"parameters"`
}

// ReportResponse ответ с данными отчета
type ReportResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	TemplateID  uint      `json:"template_id"`
	UserID      uint      `json:"user_id"`
	Status      string    `json:"status"`
	Parameters  string    `json:"parameters"`
	FilePath    string    `json:"file_path"`
	FileSize    int64     `json:"file_size"`
	MD5Hash     string    `json:"md5_hash"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse преобразует Report в ReportResponse
func (r *Report) ToResponse() ReportResponse {
	return ReportResponse{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		TemplateID:  r.TemplateID,
		UserID:      r.UserID,
		Status:      r.Status,
		Parameters:  r.Parameters,
		FilePath:    r.FilePath,
		FileSize:    r.FileSize,
		MD5Hash:     r.MD5Hash,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// ReportsResponse ответ со списком отчетов
type ReportsResponse struct {
	Reports []ReportResponse `json:"reports"`
	Total   int64            `json:"total"`
	Page    int              `json:"page"`
	Limit   int              `json:"limit"`
}

// ReportCreateResponse ответ на создание отчета (асинхронный)
type ReportCreateResponse struct {
	ID      uint   `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ReportStatusResponse ответ со статусом отчета
type ReportStatusResponse struct {
	ID       uint   `json:"id"`
	Status   string `json:"status"`
	FilePath string `json:"file_path,omitempty"`
	Progress int    `json:"progress,omitempty"`
	Error    string `json:"error,omitempty"`
}
