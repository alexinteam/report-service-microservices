package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"template-service/internal/metrics"
	"template-service/internal/models"
	"template-service/internal/repository"

	"gorm.io/gorm"
)

type TemplateService struct {
	templateRepo *repository.TemplateRepository
	metrics      *metrics.Metrics
}

func NewTemplateService(templateRepo *repository.TemplateRepository, metrics *metrics.Metrics) *TemplateService {
	return &TemplateService{
		templateRepo: templateRepo,
		metrics:      metrics,
	}
}

// CreateTemplate создает новый шаблон
func (s *TemplateService) CreateTemplate(req *models.TemplateCreateRequest) (*models.TemplateResponse, error) {
	start := time.Now()
	template := &models.Template{
		Name:        req.Name,
		Description: req.Description,
		Content:     req.Content,
		Type:        req.Type,
		Category:    req.Category,
		Variables:   req.Variables,
		IsActive:    req.IsActive,
	}

	if err := s.templateRepo.Create(template); err != nil {
		s.metrics.RecordDatabaseOperation("template-service", "create_template", time.Since(start), err)
		return nil, fmt.Errorf("ошибка создания шаблона: %w", err)
	}
	s.metrics.RecordDatabaseOperation("template-service", "create_template", time.Since(start), nil)

	response := template.ToResponse()
	return &response, nil
}

// GetTemplates получает список шаблонов
func (s *TemplateService) GetTemplates(page, limit int, category, active string) ([]models.TemplateResponse, int64, error) {
	var isActive *bool
	if active != "" {
		activeBool := active == "true"
		isActive = &activeBool
	}

	templates, total, err := s.templateRepo.GetAll(page, limit, category, isActive)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения шаблонов: %w", err)
	}

	responses := make([]models.TemplateResponse, len(templates))
	for i, t := range templates {
		responses[i] = t.ToResponse()
	}

	return responses, total, nil
}

// GetTemplate получает шаблон по ID
func (s *TemplateService) GetTemplate(id uint) (*models.TemplateResponse, error) {
	template, err := s.templateRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("шаблон не найден")
		}
		return nil, fmt.Errorf("ошибка получения шаблона: %w", err)
	}

	response := template.ToResponse()
	return &response, nil
}

// UpdateTemplate обновляет шаблон
func (s *TemplateService) UpdateTemplate(id uint, req *models.TemplateUpdateRequest) (*models.TemplateResponse, error) {
	template, err := s.templateRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("шаблон не найден")
		}
		return nil, fmt.Errorf("ошибка получения шаблона: %w", err)
	}

	if req.Name != "" {
		template.Name = req.Name
	}
	if req.Description != "" {
		template.Description = req.Description
	}
	if req.Content != "" {
		template.Content = req.Content
	}
	if req.Type != "" {
		template.Type = req.Type
	}
	if req.Category != "" {
		template.Category = req.Category
	}
	if req.Variables != "" {
		template.Variables = req.Variables
	}
	template.IsActive = req.IsActive

	if err := s.templateRepo.Update(template); err != nil {
		return nil, fmt.Errorf("ошибка обновления шаблона: %w", err)
	}

	response := template.ToResponse()
	return &response, nil
}

// DeleteTemplate удаляет шаблон
func (s *TemplateService) DeleteTemplate(id uint) error {
	if err := s.templateRepo.Delete(id); err != nil {
		return fmt.Errorf("ошибка удаления шаблона: %w", err)
	}
	return nil
}

// SearchTemplates ищет шаблоны
func (s *TemplateService) SearchTemplates(query string, page, limit int) ([]models.TemplateResponse, int64, error) {
	templates, total, err := s.templateRepo.Search(query, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка поиска шаблонов: %w", err)
	}

	responses := make([]models.TemplateResponse, len(templates))
	for i, t := range templates {
		responses[i] = t.ToResponse()
	}

	return responses, total, nil
}

// RenderTemplate рендерит шаблон с переменными
func (s *TemplateService) RenderTemplate(req *models.RenderTemplateRequest) (*models.RenderTemplateResponse, error) {
	template, err := s.templateRepo.GetByID(req.TemplateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("шаблон не найден")
		}
		return nil, fmt.Errorf("ошибка получения шаблона: %w", err)
	}

	// Простой рендеринг - замена переменных в шаблоне
	content := template.Content
	for key, value := range req.Variables {
		placeholder := "{{" + key + "}}"
		content = strings.ReplaceAll(content, placeholder, fmt.Sprintf("%v", value))
	}

	format := req.Format
	if format == "" {
		format = template.Type
	}

	return &models.RenderTemplateResponse{
		Content: content,
		Format:  format,
		Size:    len(content),
	}, nil
}

type TemplateCategoryService struct {
	categoryRepo *repository.TemplateCategoryRepository
}

func NewTemplateCategoryService(categoryRepo *repository.TemplateCategoryRepository) *TemplateCategoryService {
	return &TemplateCategoryService{
		categoryRepo: categoryRepo,
	}
}

// CreateCategory создает новую категорию
func (s *TemplateCategoryService) CreateCategory(req *models.TemplateCategoryCreateRequest) (*models.TemplateCategoryResponse, error) {
	category := &models.TemplateCategory{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
	}

	if err := s.categoryRepo.Create(category); err != nil {
		return nil, fmt.Errorf("ошибка создания категории: %w", err)
	}

	response := category.ToResponse()
	return &response, nil
}

// GetCategories получает список категорий
func (s *TemplateCategoryService) GetCategories(page, limit int, active string) ([]models.TemplateCategoryResponse, int64, error) {
	var isActive *bool
	if active != "" {
		activeBool := active == "true"
		isActive = &activeBool
	}

	categories, total, err := s.categoryRepo.GetAll(page, limit, isActive)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения категорий: %w", err)
	}

	responses := make([]models.TemplateCategoryResponse, len(categories))
	for i, c := range categories {
		responses[i] = c.ToResponse()
	}

	return responses, total, nil
}

// GetCategory получает категорию по ID
func (s *TemplateCategoryService) GetCategory(id uint) (*models.TemplateCategoryResponse, error) {
	category, err := s.categoryRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("категория не найдена")
		}
		return nil, fmt.Errorf("ошибка получения категории: %w", err)
	}

	response := category.ToResponse()
	return &response, nil
}

// UpdateCategory обновляет категорию
func (s *TemplateCategoryService) UpdateCategory(id uint, req *models.TemplateCategoryUpdateRequest) (*models.TemplateCategoryResponse, error) {
	category, err := s.categoryRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("категория не найдена")
		}
		return nil, fmt.Errorf("ошибка получения категории: %w", err)
	}

	if req.Name != "" {
		category.Name = req.Name
	}
	if req.Description != "" {
		category.Description = req.Description
	}
	category.IsActive = req.IsActive

	if err := s.categoryRepo.Update(category); err != nil {
		return nil, fmt.Errorf("ошибка обновления категории: %w", err)
	}

	response := category.ToResponse()
	return &response, nil
}

// DeleteCategory удаляет категорию
func (s *TemplateCategoryService) DeleteCategory(id uint) error {
	if err := s.categoryRepo.Delete(id); err != nil {
		return fmt.Errorf("ошибка удаления категории: %w", err)
	}
	return nil
}

type TemplateVariableService struct {
	variableRepo *repository.TemplateVariableRepository
}

func NewTemplateVariableService(variableRepo *repository.TemplateVariableRepository) *TemplateVariableService {
	return &TemplateVariableService{
		variableRepo: variableRepo,
	}
}

// CreateVariable создает новую переменную
func (s *TemplateVariableService) CreateVariable(req *models.TemplateVariableCreateRequest) (*models.TemplateVariableResponse, error) {
	variable := &models.TemplateVariable{
		TemplateID:  req.TemplateID,
		Name:        req.Name,
		Type:        req.Type,
		Required:    req.Required,
		Default:     req.Default,
		Description: req.Description,
	}

	if err := s.variableRepo.Create(variable); err != nil {
		return nil, fmt.Errorf("ошибка создания переменной: %w", err)
	}

	response := variable.ToResponse()
	return &response, nil
}

// GetVariables получает список переменных
func (s *TemplateVariableService) GetVariables(page, limit int, templateID uint) ([]models.TemplateVariableResponse, int64, error) {
	variables, total, err := s.variableRepo.GetAll(page, limit, templateID)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения переменных: %w", err)
	}

	responses := make([]models.TemplateVariableResponse, len(variables))
	for i, v := range variables {
		responses[i] = v.ToResponse()
	}

	return responses, total, nil
}

// GetVariable получает переменную по ID
func (s *TemplateVariableService) GetVariable(id uint) (*models.TemplateVariableResponse, error) {
	variable, err := s.variableRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("переменная не найдена")
		}
		return nil, fmt.Errorf("ошибка получения переменной: %w", err)
	}

	response := variable.ToResponse()
	return &response, nil
}

// UpdateVariable обновляет переменную
func (s *TemplateVariableService) UpdateVariable(id uint, req *models.TemplateVariableUpdateRequest) (*models.TemplateVariableResponse, error) {
	variable, err := s.variableRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("переменная не найдена")
		}
		return nil, fmt.Errorf("ошибка получения переменной: %w", err)
	}

	if req.Name != "" {
		variable.Name = req.Name
	}
	if req.Type != "" {
		variable.Type = req.Type
	}
	variable.Required = req.Required
	if req.Default != "" {
		variable.Default = req.Default
	}
	if req.Description != "" {
		variable.Description = req.Description
	}

	if err := s.variableRepo.Update(variable); err != nil {
		return nil, fmt.Errorf("ошибка обновления переменной: %w", err)
	}

	response := variable.ToResponse()
	return &response, nil
}

// DeleteVariable удаляет переменную
func (s *TemplateVariableService) DeleteVariable(id uint) error {
	if err := s.variableRepo.Delete(id); err != nil {
		return fmt.Errorf("ошибка удаления переменной: %w", err)
	}
	return nil
}
