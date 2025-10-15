package handlers

import (
	"net/http"
	"strconv"

	"notification-service/internal/models"
	"notification-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type NotificationTemplateHandler struct {
	templateService *services.NotificationTemplateService
}

func NewNotificationTemplateHandler(templateService *services.NotificationTemplateService) *NotificationTemplateHandler {
	return &NotificationTemplateHandler{
		templateService: templateService,
	}
}

// CreateTemplate создание нового шаблона уведомления
func (h *NotificationTemplateHandler) CreateTemplate(c *gin.Context) {
	var req models.NotificationTemplateCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, err := h.templateService.CreateTemplate(&req)
	if err != nil {
		logrus.WithError(err).Error("Ошибка создания шаблона уведомления")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// GetTemplates получение списка шаблонов уведомлений
func (h *NotificationTemplateHandler) GetTemplates(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	active := c.Query("active")

	templates, total, err := h.templateService.GetTemplates(page, limit, active)
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения шаблонов уведомлений")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.NotificationTemplatesResponse{
		Templates: templates,
		Total:     total,
		Page:      page,
		Limit:     limit,
	})
}

// GetTemplate получение шаблона уведомления по ID
func (h *NotificationTemplateHandler) GetTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	template, err := h.templateService.GetTemplate(uint(id))
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения шаблона уведомления")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// UpdateTemplate обновление шаблона уведомления
func (h *NotificationTemplateHandler) UpdateTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	var req models.NotificationTemplateUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template, err := h.templateService.UpdateTemplate(uint(id), &req)
	if err != nil {
		logrus.WithError(err).Error("Ошибка обновления шаблона уведомления")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// DeleteTemplate удаление шаблона уведомления
func (h *NotificationTemplateHandler) DeleteTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	if err := h.templateService.DeleteTemplate(uint(id)); err != nil {
		logrus.WithError(err).Error("Ошибка удаления шаблона уведомления")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

type NotificationHandler struct {
	notificationService *services.NotificationService
}

func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

// SendNotification отправка уведомления
func (h *NotificationHandler) SendNotification(c *gin.Context) {
	var req models.NotificationCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.notificationService.SendNotification(&req)
	if err != nil {
		logrus.WithError(err).Error("Ошибка отправки уведомления")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetNotifications получение списка уведомлений
func (h *NotificationHandler) GetNotifications(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	recipient := c.Query("recipient")

	notifications, total, err := h.notificationService.GetNotifications(page, limit, status, recipient)
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения уведомлений")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.NotificationsResponse{
		Notifications: notifications,
		Total:         total,
		Page:          page,
		Limit:         limit,
	})
}

// GetNotification получение уведомления по ID
func (h *NotificationHandler) GetNotification(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	notification, err := h.notificationService.GetNotification(uint(id))
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения уведомления")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notification)
}

// UpdateNotificationStatus обновление статуса уведомления
func (h *NotificationHandler) UpdateNotificationStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	var req struct {
		Status       string `json:"status" binding:"required"`
		ErrorMessage string `json:"error_message"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notification, err := h.notificationService.UpdateNotificationStatus(uint(id), req.Status, req.ErrorMessage)
	if err != nil {
		logrus.WithError(err).Error("Ошибка обновления статуса уведомления")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notification)
}

type NotificationChannelHandler struct {
	channelService *services.NotificationChannelService
}

func NewNotificationChannelHandler(channelService *services.NotificationChannelService) *NotificationChannelHandler {
	return &NotificationChannelHandler{
		channelService: channelService,
	}
}

// CreateChannel создание нового канала уведомлений
func (h *NotificationChannelHandler) CreateChannel(c *gin.Context) {
	var req models.NotificationChannelCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channel, err := h.channelService.CreateChannel(&req)
	if err != nil {
		logrus.WithError(err).Error("Ошибка создания канала уведомлений")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, channel)
}

// GetChannels получение списка каналов уведомлений
func (h *NotificationChannelHandler) GetChannels(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	active := c.Query("active")

	channels, total, err := h.channelService.GetChannels(page, limit, active)
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения каналов уведомлений")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.NotificationChannelsResponse{
		Channels: channels,
		Total:    total,
		Page:     page,
		Limit:    limit,
	})
}

// GetChannel получение канала уведомлений по ID
func (h *NotificationChannelHandler) GetChannel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	channel, err := h.channelService.GetChannel(uint(id))
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения канала уведомлений")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, channel)
}

// UpdateChannel обновление канала уведомлений
func (h *NotificationChannelHandler) UpdateChannel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	var req models.NotificationChannelUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channel, err := h.channelService.UpdateChannel(uint(id), &req)
	if err != nil {
		logrus.WithError(err).Error("Ошибка обновления канала уведомлений")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, channel)
}

func (h *NotificationChannelHandler) DeleteChannel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	if err := h.channelService.DeleteChannel(uint(id)); err != nil {
		logrus.WithError(err).Error("Ошибка удаления канала уведомлений")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
