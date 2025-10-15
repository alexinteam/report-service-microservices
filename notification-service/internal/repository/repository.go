package repository

import (
	"notification-service/internal/models"

	"gorm.io/gorm"
)

type NotificationTemplateRepository struct {
	db *gorm.DB
}

func NewNotificationTemplateRepository(db *gorm.DB) *NotificationTemplateRepository {
	return &NotificationTemplateRepository{db: db}
}

// Create создает новый шаблон уведомления
func (r *NotificationTemplateRepository) Create(template *models.NotificationTemplate) error {
	return r.db.Create(template).Error
}

// GetByID получает шаблон уведомления по ID
func (r *NotificationTemplateRepository) GetByID(id uint) (*models.NotificationTemplate, error) {
	var template models.NotificationTemplate
	err := r.db.First(&template, id).Error
	return &template, err
}

// GetAll получает все шаблоны уведомлений с пагинацией
func (r *NotificationTemplateRepository) GetAll(page, limit int, isActive *bool) ([]models.NotificationTemplate, int64, error) {
	var templates []models.NotificationTemplate
	var total int64

	query := r.db.Model(&models.NotificationTemplate{})
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Find(&templates).Error
	return templates, total, err
}

// Update обновляет шаблон уведомления
func (r *NotificationTemplateRepository) Update(template *models.NotificationTemplate) error {
	return r.db.Save(template).Error
}

// Delete удаляет шаблон уведомления
func (r *NotificationTemplateRepository) Delete(id uint) error {
	return r.db.Delete(&models.NotificationTemplate{}, id).Error
}

type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository создает новый репозиторий уведомлений
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create создает новое уведомление
func (r *NotificationRepository) Create(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

// GetByID получает уведомление по ID
func (r *NotificationRepository) GetByID(id uint) (*models.Notification, error) {
	var notification models.Notification
	err := r.db.First(&notification, id).Error
	return &notification, err
}

// GetAll получает все уведомления с пагинацией
func (r *NotificationRepository) GetAll(page, limit int, status, recipient string) ([]models.Notification, int64, error) {
	var notifications []models.Notification
	var total int64

	query := r.db.Model(&models.Notification{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if recipient != "" {
		query = query.Where("recipient = ?", recipient)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&notifications).Error
	return notifications, total, err
}

// Update обновляет уведомление
func (r *NotificationRepository) Update(notification *models.Notification) error {
	return r.db.Save(notification).Error
}

// Delete удаляет уведомление
func (r *NotificationRepository) Delete(id uint) error {
	return r.db.Delete(&models.Notification{}, id).Error
}

type NotificationChannelRepository struct {
	db *gorm.DB
}

func NewNotificationChannelRepository(db *gorm.DB) *NotificationChannelRepository {
	return &NotificationChannelRepository{db: db}
}

// Create создает новый канал уведомлений
func (r *NotificationChannelRepository) Create(channel *models.NotificationChannel) error {
	return r.db.Create(channel).Error
}

// GetByID получает канал уведомлений по ID
func (r *NotificationChannelRepository) GetByID(id uint) (*models.NotificationChannel, error) {
	var channel models.NotificationChannel
	err := r.db.First(&channel, id).Error
	return &channel, err
}

// GetAll получает все каналы уведомлений с пагинацией
func (r *NotificationChannelRepository) GetAll(page, limit int, isActive *bool) ([]models.NotificationChannel, int64, error) {
	var channels []models.NotificationChannel
	var total int64

	query := r.db.Model(&models.NotificationChannel{})
	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Find(&channels).Error
	return channels, total, err
}

// Update обновляет канал уведомлений
func (r *NotificationChannelRepository) Update(channel *models.NotificationChannel) error {
	return r.db.Save(channel).Error
}

// Delete удаляет канал уведомлений
func (r *NotificationChannelRepository) Delete(id uint) error {
	return r.db.Delete(&models.NotificationChannel{}, id).Error
}
