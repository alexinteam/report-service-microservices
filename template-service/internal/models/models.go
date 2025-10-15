package models

import (
	"time"

	"gorm.io/gorm"
)

type Template struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Content     string         `json:"content" gorm:"type:text"`
	Type        string         `json:"type" gorm:"not null"` // html, pdf, excel, csv
	Category    string         `json:"category"`
	Variables   string         `json:"variables" gorm:"type:text"` // JSON переменные
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Template) TableName() string {
	return "templates"
}

type TemplateCategory struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (TemplateCategory) TableName() string {
	return "template_categories"
}

type TemplateVariable struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	TemplateID  uint           `json:"template_id" gorm:"not null"`
	Name        string         `json:"name" gorm:"not null"`
	Type        string         `json:"type" gorm:"not null"` // string, number, date, boolean
	Required    bool           `json:"required" gorm:"default:false"`
	Default     string         `json:"default"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (TemplateVariable) TableName() string {
	return "template_variables"
}

type TemplateCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Content     string `json:"content" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Category    string `json:"category"`
	Variables   string `json:"variables"`
	IsActive    bool   `json:"is_active"`
}

type TemplateUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
	Type        string `json:"type"`
	Category    string `json:"category"`
	Variables   string `json:"variables"`
	IsActive    bool   `json:"is_active"`
}

type TemplateCategoryCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

type TemplateCategoryUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

type TemplateVariableCreateRequest struct {
	TemplateID  uint   `json:"template_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Required    bool   `json:"required"`
	Default     string `json:"default"`
	Description string `json:"description"`
}

type TemplateVariableUpdateRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Default     string `json:"default"`
	Description string `json:"description"`
}

type TemplateResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	Type        string    `json:"type"`
	Category    string    `json:"category"`
	Variables   string    `json:"variables"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (t *Template) ToResponse() TemplateResponse {
	return TemplateResponse{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
		Content:     t.Content,
		Type:        t.Type,
		Category:    t.Category,
		Variables:   t.Variables,
		IsActive:    t.IsActive,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

type TemplateCategoryResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (tc *TemplateCategory) ToResponse() TemplateCategoryResponse {
	return TemplateCategoryResponse{
		ID:          tc.ID,
		Name:        tc.Name,
		Description: tc.Description,
		IsActive:    tc.IsActive,
		CreatedAt:   tc.CreatedAt,
		UpdatedAt:   tc.UpdatedAt,
	}
}

type TemplateVariableResponse struct {
	ID          uint      `json:"id"`
	TemplateID  uint      `json:"template_id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Required    bool      `json:"required"`
	Default     string    `json:"default"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (tv *TemplateVariable) ToResponse() TemplateVariableResponse {
	return TemplateVariableResponse{
		ID:          tv.ID,
		TemplateID:  tv.TemplateID,
		Name:        tv.Name,
		Type:        tv.Type,
		Required:    tv.Required,
		Default:     tv.Default,
		Description: tv.Description,
		CreatedAt:   tv.CreatedAt,
		UpdatedAt:   tv.UpdatedAt,
	}
}

type TemplatesResponse struct {
	Templates []TemplateResponse `json:"templates"`
	Total     int64              `json:"total"`
	Page      int                `json:"page"`
	Limit     int                `json:"limit"`
}

type TemplateCategoriesResponse struct {
	Categories []TemplateCategoryResponse `json:"categories"`
	Total      int64                      `json:"total"`
	Page       int                        `json:"page"`
	Limit      int                        `json:"limit"`
}

type TemplateVariablesResponse struct {
	Variables []TemplateVariableResponse `json:"variables"`
	Total     int64                      `json:"total"`
	Page      int                        `json:"page"`
	Limit     int                        `json:"limit"`
}

type RenderTemplateRequest struct {
	TemplateID uint                   `json:"template_id" binding:"required"`
	Variables  map[string]interface{} `json:"variables"`
	Format     string                 `json:"format"` // html, pdf, excel, csv
}

type RenderTemplateResponse struct {
	Content string `json:"content"`
	Format  string `json:"format"`
	Size    int    `json:"size"`
}
