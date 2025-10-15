package models

import (
	"time"

	"gorm.io/gorm"
)

// NotificationTemplate модель шаблона уведомления
type NotificationTemplate struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"not null"`
	Subject   string         `json:"subject" gorm:"not null"`
	Body      string         `json:"body" gorm:"type:text"`
	Type      string         `json:"type" gorm:"not null"`       // email, sms, push, webhook
	Variables string         `json:"variables" gorm:"type:text"` // JSON переменные
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (NotificationTemplate) TableName() string {
	return "notification_templates"
}

// Notification модель уведомления
type Notification struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	TemplateID   uint           `json:"template_id" gorm:"not null"`
	Recipient    string         `json:"recipient" gorm:"not null"` // email, phone, user_id
	Subject      string         `json:"subject"`
	Body         string         `json:"body" gorm:"type:text"`
	Type         string         `json:"type" gorm:"not null"`            // email, sms, push, webhook
	Status       string         `json:"status" gorm:"default:'pending'"` // pending, sent, failed, delivered
	Data         string         `json:"data" gorm:"type:text"`           // JSON данные для подстановки
	SentAt       *time.Time     `json:"sent_at"`
	DeliveredAt  *time.Time     `json:"delivered_at"`
	ErrorMessage string         `json:"error_message"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Notification) TableName() string {
	return "notifications"
}

// NotificationChannel модель канала уведомлений
type NotificationChannel struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"not null"`
	Type      string         `json:"type" gorm:"not null"`    // email, sms, push, webhook
	Config    string         `json:"config" gorm:"type:text"` // JSON конфигурация
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (NotificationChannel) TableName() string {
	return "notification_channels"
}

type NotificationTemplateCreateRequest struct {
	Name      string `json:"name" binding:"required"`
	Subject   string `json:"subject" binding:"required"`
	Body      string `json:"body"`
	Type      string `json:"type" binding:"required"`
	Variables string `json:"variables"`
	IsActive  bool   `json:"is_active"`
}

type NotificationTemplateUpdateRequest struct {
	Name      string `json:"name"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	Type      string `json:"type"`
	Variables string `json:"variables"`
	IsActive  bool   `json:"is_active"`
}

type NotificationCreateRequest struct {
	TemplateID uint                   `json:"template_id" binding:"required"`
	Recipient  string                 `json:"recipient" binding:"required"`
	Data       map[string]interface{} `json:"data"`
	Type       string                 `json:"type"`
}

type NotificationChannelCreateRequest struct {
	Name     string `json:"name" binding:"required"`
	Type     string `json:"type" binding:"required"`
	Config   string `json:"config"`
	IsActive bool   `json:"is_active"`
}

type NotificationChannelUpdateRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Config   string `json:"config"`
	IsActive bool   `json:"is_active"`
}

type NotificationTemplateResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	Type      string    `json:"type"`
	Variables string    `json:"variables"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (nt *NotificationTemplate) ToResponse() NotificationTemplateResponse {
	return NotificationTemplateResponse{
		ID:        nt.ID,
		Name:      nt.Name,
		Subject:   nt.Subject,
		Body:      nt.Body,
		Type:      nt.Type,
		Variables: nt.Variables,
		IsActive:  nt.IsActive,
		CreatedAt: nt.CreatedAt,
		UpdatedAt: nt.UpdatedAt,
	}
}

type NotificationResponse struct {
	ID           uint       `json:"id"`
	TemplateID   uint       `json:"template_id"`
	Recipient    string     `json:"recipient"`
	Subject      string     `json:"subject"`
	Body         string     `json:"body"`
	Type         string     `json:"type"`
	Status       string     `json:"status"`
	Data         string     `json:"data"`
	SentAt       *time.Time `json:"sent_at"`
	DeliveredAt  *time.Time `json:"delivered_at"`
	ErrorMessage string     `json:"error_message"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

func (n *Notification) ToResponse() NotificationResponse {
	return NotificationResponse{
		ID:           n.ID,
		TemplateID:   n.TemplateID,
		Recipient:    n.Recipient,
		Subject:      n.Subject,
		Body:         n.Body,
		Type:         n.Type,
		Status:       n.Status,
		Data:         n.Data,
		SentAt:       n.SentAt,
		DeliveredAt:  n.DeliveredAt,
		ErrorMessage: n.ErrorMessage,
		CreatedAt:    n.CreatedAt,
		UpdatedAt:    n.UpdatedAt,
	}
}

type NotificationChannelResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Config    string    `json:"config"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (nc *NotificationChannel) ToResponse() NotificationChannelResponse {
	return NotificationChannelResponse{
		ID:        nc.ID,
		Name:      nc.Name,
		Type:      nc.Type,
		Config:    nc.Config,
		IsActive:  nc.IsActive,
		CreatedAt: nc.CreatedAt,
		UpdatedAt: nc.UpdatedAt,
	}
}

type NotificationTemplatesResponse struct {
	Templates []NotificationTemplateResponse `json:"templates"`
	Total     int64                          `json:"total"`
	Page      int                            `json:"page"`
	Limit     int                            `json:"limit"`
}

type NotificationsResponse struct {
	Notifications []NotificationResponse `json:"notifications"`
	Total         int64                  `json:"total"`
	Page          int                    `json:"page"`
	Limit         int                    `json:"limit"`
}

type NotificationChannelsResponse struct {
	Channels []NotificationChannelResponse `json:"channels"`
	Total    int64                         `json:"total"`
	Page     int                           `json:"page"`
	Limit    int                           `json:"limit"`
}

type SendNotificationResponse struct {
	NotificationID uint   `json:"notification_id"`
	Status         string `json:"status"`
	Message        string `json:"message"`
}
