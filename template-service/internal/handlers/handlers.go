package handlers

import (
	"net/http"
	"strconv"

	"template-service/internal/models"
	"template-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type TemplateHandler struct {
	templateService *services.TemplateService
}

func NewTemplateHandler(templateService *services.TemplateService) *TemplateHandler {
	return &TemplateHandler{
		templateService: templateService,
	}
}

// CreateTemplate создание нового шаблона
func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	var req models.TemplateCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, err := h.templateService.CreateTemplate(&req)
	if err != nil {
		logrus.WithError(err).Error("Ошибка создания шаблона")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// GetTemplates получение списка шаблонов
func (h *TemplateHandler) GetTemplates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	category := c.Query("category")
	active := c.Query("active")

	templates, total, err := h.templateService.GetTemplates(page, limit, category, active)
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения шаблонов")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.TemplatesResponse{
		Templates: templates,
		Total:     total,
		Page:      page,
		Limit:     limit,
	})
}

// GetTemplate получение шаблона по ID
func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	template, err := h.templateService.GetTemplate(uint(id))
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения шаблона")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// UpdateTemplate обновление шаблона
func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	var req models.TemplateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, err := h.templateService.UpdateTemplate(uint(id), &req)
	if err != nil {
		logrus.WithError(err).Error("Ошибка обновления шаблона")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// DeleteTemplate удаление шаблона
func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	if err := h.templateService.DeleteTemplate(uint(id)); err != nil {
		logrus.WithError(err).Error("Ошибка удаления шаблона")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// SearchTemplates поиск шаблонов
func (h *TemplateHandler) SearchTemplates(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Поисковый запрос не указан"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	templates, total, err := h.templateService.SearchTemplates(query, page, limit)
	if err != nil {
		logrus.WithError(err).Error("Ошибка поиска шаблонов")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.TemplatesResponse{
		Templates: templates,
		Total:     total,
		Page:      page,
		Limit:     limit,
	})
}

// RenderTemplate рендеринг шаблона
func (h *TemplateHandler) RenderTemplate(c *gin.Context) {
	var req models.RenderTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.templateService.RenderTemplate(&req)
	if err != nil {
		logrus.WithError(err).Error("Ошибка рендеринга шаблона")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

type TemplateCategoryHandler struct {
	categoryService *services.TemplateCategoryService
}

func NewTemplateCategoryHandler(categoryService *services.TemplateCategoryService) *TemplateCategoryHandler {
	return &TemplateCategoryHandler{
		categoryService: categoryService,
	}
}

func (h *TemplateCategoryHandler) CreateCategory(c *gin.Context) {
	var req models.TemplateCategoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.categoryService.CreateCategory(&req)
	if err != nil {
		logrus.WithError(err).Error("Ошибка создания категории")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// GetCategories получение списка категорий
func (h *TemplateCategoryHandler) GetCategories(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	active := c.Query("active")

	categories, total, err := h.categoryService.GetCategories(page, limit, active)
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения категорий")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.TemplateCategoriesResponse{
		Categories: categories,
		Total:      total,
		Page:       page,
		Limit:      limit,
	})
}

// GetCategory получение категории по ID
func (h *TemplateCategoryHandler) GetCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	category, err := h.categoryService.GetCategory(uint(id))
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения категории")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

// UpdateCategory обновление категории
func (h *TemplateCategoryHandler) UpdateCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	var req models.TemplateCategoryUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category, err := h.categoryService.UpdateCategory(uint(id), &req)
	if err != nil {
		logrus.WithError(err).Error("Ошибка обновления категории")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}

// DeleteCategory удаление категории
func (h *TemplateCategoryHandler) DeleteCategory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	if err := h.categoryService.DeleteCategory(uint(id)); err != nil {
		logrus.WithError(err).Error("Ошибка удаления категории")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

type TemplateVariableHandler struct {
	variableService *services.TemplateVariableService
}

func NewTemplateVariableHandler(variableService *services.TemplateVariableService) *TemplateVariableHandler {
	return &TemplateVariableHandler{
		variableService: variableService,
	}
}

// CreateVariable создание новой переменной
func (h *TemplateVariableHandler) CreateVariable(c *gin.Context) {
	var req models.TemplateVariableCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	variable, err := h.variableService.CreateVariable(&req)
	if err != nil {
		logrus.WithError(err).Error("Ошибка создания переменной")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, variable)
}

// GetVariables получение списка переменных
func (h *TemplateVariableHandler) GetVariables(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	templateIDStr := c.Query("template_id")

	var templateID uint
	if templateIDStr != "" {
		id, err := strconv.ParseUint(templateIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный template_id"})
			return
		}
		templateID = uint(id)
	}

	variables, total, err := h.variableService.GetVariables(page, limit, templateID)
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения переменных")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.TemplateVariablesResponse{
		Variables: variables,
		Total:     total,
		Page:      page,
		Limit:     limit,
	})
}

// GetVariable получение переменной по ID
func (h *TemplateVariableHandler) GetVariable(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	variable, err := h.variableService.GetVariable(uint(id))
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения переменной")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, variable)
}

// UpdateVariable обновление переменной
func (h *TemplateVariableHandler) UpdateVariable(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	var req models.TemplateVariableUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	variable, err := h.variableService.UpdateVariable(uint(id), &req)
	if err != nil {
		logrus.WithError(err).Error("Ошибка обновления переменной")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, variable)
}

// DeleteVariable удаление переменной
func (h *TemplateVariableHandler) DeleteVariable(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	if err := h.variableService.DeleteVariable(uint(id)); err != nil {
		logrus.WithError(err).Error("Ошибка удаления переменной")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
