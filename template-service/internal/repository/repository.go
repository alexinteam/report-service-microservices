package repository

import (
	"template-service/internal/models"

	"gorm.io/gorm"
)

type TemplateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// Create создает новый шаблон
func (r *TemplateRepository) Create(template *models.Template) error {
	return r.db.Create(template).Error
}

// GetByID получает шаблон по ID
func (r *TemplateRepository) GetByID(id uint) (*models.Template, error) {
	var template models.Template
	err := r.db.First(&template, id).Error
	return &template, err
}

// GetAll получает все шаблоны с пагинацией
func (r *TemplateRepository) GetAll(page, limit int, category string, isActive *bool) ([]models.Template, int64, error) {
	var templates []models.Template
	var total int64

	query := r.db.Model(&models.Template{})
	if category != "" {
		query = query.Where("category = ?", category)
	}
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&templates).Error
	return templates, total, err
}

// Update обновляет шаблон
func (r *TemplateRepository) Update(template *models.Template) error {
	return r.db.Save(template).Error
}

// Delete удаляет шаблон
func (r *TemplateRepository) Delete(id uint) error {
	return r.db.Delete(&models.Template{}, id).Error
}

// Search ищет шаблоны по имени и описанию
func (r *TemplateRepository) Search(query string, page, limit int) ([]models.Template, int64, error) {
	var templates []models.Template
	var total int64

	searchQuery := "%" + query + "%"
	queryBuilder := r.db.Model(&models.Template{}).Where(
		"name ILIKE ? OR description ILIKE ?",
		searchQuery, searchQuery,
	)

	if err := queryBuilder.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := queryBuilder.Offset(offset).Limit(limit).Order("created_at DESC").Find(&templates).Error
	return templates, total, err
}

type TemplateCategoryRepository struct {
	db *gorm.DB
}

func NewTemplateCategoryRepository(db *gorm.DB) *TemplateCategoryRepository {
	return &TemplateCategoryRepository{db: db}
}

// Create создает новую категорию
func (r *TemplateCategoryRepository) Create(category *models.TemplateCategory) error {
	return r.db.Create(category).Error
}

// GetByID получает категорию по ID
func (r *TemplateCategoryRepository) GetByID(id uint) (*models.TemplateCategory, error) {
	var category models.TemplateCategory
	err := r.db.First(&category, id).Error
	return &category, err
}

// GetAll получает все категории с пагинацией
func (r *TemplateCategoryRepository) GetAll(page, limit int, isActive *bool) ([]models.TemplateCategory, int64, error) {
	var categories []models.TemplateCategory
	var total int64

	query := r.db.Model(&models.TemplateCategory{})
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&categories).Error
	return categories, total, err
}

// Update обновляет категорию
func (r *TemplateCategoryRepository) Update(category *models.TemplateCategory) error {
	return r.db.Save(category).Error
}

// Delete удаляет категорию
func (r *TemplateCategoryRepository) Delete(id uint) error {
	return r.db.Delete(&models.TemplateCategory{}, id).Error
}

type TemplateVariableRepository struct {
	db *gorm.DB
}

func NewTemplateVariableRepository(db *gorm.DB) *TemplateVariableRepository {
	return &TemplateVariableRepository{db: db}
}

// Create создает новую переменную
func (r *TemplateVariableRepository) Create(variable *models.TemplateVariable) error {
	return r.db.Create(variable).Error
}

// GetByID получает переменную по ID
func (r *TemplateVariableRepository) GetByID(id uint) (*models.TemplateVariable, error) {
	var variable models.TemplateVariable
	err := r.db.First(&variable, id).Error
	return &variable, err
}

// GetByTemplateID получает переменные по ID шаблона
func (r *TemplateVariableRepository) GetByTemplateID(templateID uint) ([]models.TemplateVariable, error) {
	var variables []models.TemplateVariable
	err := r.db.Where("template_id = ?", templateID).Find(&variables).Error
	return variables, err
}

// GetAll получает все переменные с пагинацией
func (r *TemplateVariableRepository) GetAll(page, limit int, templateID uint) ([]models.TemplateVariable, int64, error) {
	var variables []models.TemplateVariable
	var total int64

	query := r.db.Model(&models.TemplateVariable{})
	if templateID != 0 {
		query = query.Where("template_id = ?", templateID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&variables).Error
	return variables, total, err
}

// Update обновляет переменную
func (r *TemplateVariableRepository) Update(variable *models.TemplateVariable) error {
	return r.db.Save(variable).Error
}

// Delete удаляет переменную
func (r *TemplateVariableRepository) Delete(id uint) error {
	return r.db.Delete(&models.TemplateVariable{}, id).Error
}
