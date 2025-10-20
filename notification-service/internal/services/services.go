package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"notification-service/internal/models"
	"notification-service/internal/repository"

	"gorm.io/gorm"
)

// replaceVariables заменяет переменные в шаблоне на значения из данных
func replaceVariables(template string, data map[string]interface{}) string {
	// Регулярное выражение для поиска переменных в формате {{variable}}
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)

	return re.ReplaceAllStringFunc(template, func(match string) string {
		// Извлекаем имя переменной (убираем {{ и }})
		variableName := strings.TrimSpace(match[2 : len(match)-2])

		// Ищем значение в данных
		if value, exists := data[variableName]; exists {
			// Преобразуем значение в строку
			switch v := value.(type) {
			case string:
				return v
			case int, int32, int64:
				return fmt.Sprintf("%d", v)
			case uint, uint32, uint64:
				return fmt.Sprintf("%d", v)
			case float32, float64:
				return fmt.Sprintf("%.0f", v)
			default:
				return fmt.Sprintf("%v", v)
			}
		}

		// Если переменная не найдена, возвращаем оригинальную строку
		return match
	})
}

type NotificationTemplateService struct {
	templateRepo *repository.NotificationTemplateRepository
}

func NewNotificationTemplateService(templateRepo *repository.NotificationTemplateRepository) *NotificationTemplateService {
	return &NotificationTemplateService{
		templateRepo: templateRepo,
	}
}

// CreateTemplate создает новый шаблон уведомления
func (s *NotificationTemplateService) CreateTemplate(req *models.NotificationTemplateCreateRequest) (*models.NotificationTemplateResponse, error) {
	template := &models.NotificationTemplate{
		Name:      req.Name,
		Subject:   req.Subject,
		Body:      req.Body,
		Type:      req.Type,
		Variables: req.Variables,
		IsActive:  req.IsActive,
	}

	if err := s.templateRepo.Create(template); err != nil {
		return nil, fmt.Errorf("ошибка создания шаблона уведомления: %w", err)
	}

	response := template.ToResponse()
	return &response, nil
}

// GetTemplates получает список шаблонов уведомлений
func (s *NotificationTemplateService) GetTemplates(page, limit int, active string) ([]models.NotificationTemplateResponse, int64, error) {
	var isActive *bool
	if active != "" {
		activeBool := active == "true"
		isActive = &activeBool
	}

	templates, total, err := s.templateRepo.GetAll(page, limit, isActive)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения шаблонов уведомлений: %w", err)
	}

	responses := make([]models.NotificationTemplateResponse, len(templates))
	for i, t := range templates {
		responses[i] = t.ToResponse()
	}

	return responses, total, nil
}

// GetTemplate получает шаблон уведомления по ID
func (s *NotificationTemplateService) GetTemplate(id uint) (*models.NotificationTemplateResponse, error) {
	template, err := s.templateRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("шаблон уведомления не найден")
		}
		return nil, fmt.Errorf("ошибка получения шаблона уведомления: %w", err)
	}

	response := template.ToResponse()
	return &response, nil
}

// UpdateTemplate обновляет шаблон уведомления
func (s *NotificationTemplateService) UpdateTemplate(id uint, req *models.NotificationTemplateUpdateRequest) (*models.NotificationTemplateResponse, error) {
	template, err := s.templateRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("шаблон уведомления не найден")
		}
		return nil, fmt.Errorf("ошибка получения шаблона уведомления: %w", err)
	}

	// Обновляем поля
	if req.Name != "" {
		template.Name = req.Name
	}
	if req.Subject != "" {
		template.Subject = req.Subject
	}
	if req.Body != "" {
		template.Body = req.Body
	}
	if req.Type != "" {
		template.Type = req.Type
	}
	if req.Variables != "" {
		template.Variables = req.Variables
	}
	template.IsActive = req.IsActive

	if err := s.templateRepo.Update(template); err != nil {
		return nil, fmt.Errorf("ошибка обновления шаблона уведомления: %w", err)
	}

	response := template.ToResponse()
	return &response, nil
}

// DeleteTemplate удаляет шаблон уведомления
func (s *NotificationTemplateService) DeleteTemplate(id uint) error {
	if err := s.templateRepo.Delete(id); err != nil {
		return fmt.Errorf("ошибка удаления шаблона уведомления: %w", err)
	}
	return nil
}

type NotificationService struct {
	notificationRepo *repository.NotificationRepository
	templateRepo     *repository.NotificationTemplateRepository
}

func NewNotificationService(notificationRepo *repository.NotificationRepository, templateRepo *repository.NotificationTemplateRepository) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		templateRepo:     templateRepo,
	}
}

// SendNotification отправляет уведомление
func (s *NotificationService) SendNotification(req *models.NotificationCreateRequest) (*models.SendNotificationResponse, error) {
	template, err := s.templateRepo.GetByID(req.TemplateID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Автосоздание дефолтного шаблона, если указанный не найден
			defaultTemplate := &models.NotificationTemplate{
				Name:      "Report Ready",
				Subject:   "Report Ready",
				Body:      "Report {{report_id}} is ready",
				Type:      "email",
				Variables: "{}",
				IsActive:  true,
			}
			if createErr := s.templateRepo.Create(defaultTemplate); createErr != nil {
				return nil, fmt.Errorf("не удалось создать дефолтный шаблон уведомления: %w", createErr)
			}
			template = defaultTemplate
			// Переключаемся на только что созданный шаблон
			req.TemplateID = template.ID
		} else {
			return nil, fmt.Errorf("ошибка получения шаблона уведомления: %w", err)
		}
	}

	dataJSON := ""
	if req.Data != nil {
		dataBytes, err := json.Marshal(req.Data)
		if err != nil {
			return nil, fmt.Errorf("ошибка сериализации данных: %w", err)
		}
		dataJSON = string(dataBytes)
	}

	// Заменяем переменные в шаблоне
	subject := replaceVariables(template.Subject, req.Data)
	body := replaceVariables(template.Body, req.Data)

	notification := &models.Notification{
		TemplateID: req.TemplateID,
		Recipient:  req.Recipient,
		Subject:    subject,
		Body:       body,
		Type:       req.Type,
		Status:     "pending",
		Data:       dataJSON,
	}

	if notification.Type == "" {
		notification.Type = template.Type
	}

	if err := s.notificationRepo.Create(notification); err != nil {
		return nil, fmt.Errorf("ошибка создания уведомления: %w", err)
	}

	now := time.Now()
	notification.Status = "sent"
	notification.SentAt = &now

	if err := s.notificationRepo.Update(notification); err != nil {
		return nil, fmt.Errorf("ошибка обновления статуса уведомления: %w", err)
	}

	return &models.SendNotificationResponse{
		NotificationID: notification.ID,
		Status:         notification.Status,
		Message:        "Уведомление отправлено успешно",
	}, nil
}

// GetNotifications получает список уведомлений
func (s *NotificationService) GetNotifications(page, limit int, status, recipient string) ([]models.NotificationResponse, int64, error) {
	notifications, total, err := s.notificationRepo.GetAll(page, limit, status, recipient)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения уведомлений: %w", err)
	}

	responses := make([]models.NotificationResponse, len(notifications))
	for i, n := range notifications {
		responses[i] = n.ToResponse()
	}

	return responses, total, nil
}

// GetNotification получает уведомление по ID
func (s *NotificationService) GetNotification(id uint) (*models.NotificationResponse, error) {
	notification, err := s.notificationRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("уведомление не найдено")
		}
		return nil, fmt.Errorf("ошибка получения уведомления: %w", err)
	}

	response := notification.ToResponse()
	return &response, nil
}

// UpdateNotificationStatus обновляет статус уведомления
func (s *NotificationService) UpdateNotificationStatus(id uint, status, errorMessage string) (*models.NotificationResponse, error) {
	notification, err := s.notificationRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("уведомление не найдено")
		}
		return nil, fmt.Errorf("ошибка получения уведомления: %w", err)
	}

	notification.Status = status
	notification.ErrorMessage = errorMessage

	if status == "sent" && notification.SentAt == nil {
		now := time.Now()
		notification.SentAt = &now
	}

	if status == "delivered" && notification.DeliveredAt == nil {
		now := time.Now()
		notification.DeliveredAt = &now
	}

	if err := s.notificationRepo.Update(notification); err != nil {
		return nil, fmt.Errorf("ошибка обновления статуса уведомления: %w", err)
	}

	response := notification.ToResponse()
	return &response, nil
}

type NotificationChannelService struct {
	channelRepo *repository.NotificationChannelRepository
}

func NewNotificationChannelService(channelRepo *repository.NotificationChannelRepository) *NotificationChannelService {
	return &NotificationChannelService{
		channelRepo: channelRepo,
	}
}

// CreateChannel создает новый канал уведомлений
func (s *NotificationChannelService) CreateChannel(req *models.NotificationChannelCreateRequest) (*models.NotificationChannelResponse, error) {
	channel := &models.NotificationChannel{
		Name:     req.Name,
		Type:     req.Type,
		Config:   req.Config,
		IsActive: req.IsActive,
	}

	if err := s.channelRepo.Create(channel); err != nil {
		return nil, fmt.Errorf("ошибка создания канала уведомлений: %w", err)
	}

	response := channel.ToResponse()
	return &response, nil
}

// GetChannels получает список каналов уведомлений
func (s *NotificationChannelService) GetChannels(page, limit int, active string) ([]models.NotificationChannelResponse, int64, error) {
	var isActive *bool
	if active != "" {
		activeBool := active == "true"
		isActive = &activeBool
	}

	channels, total, err := s.channelRepo.GetAll(page, limit, isActive)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения каналов уведомлений: %w", err)
	}

	responses := make([]models.NotificationChannelResponse, len(channels))
	for i, c := range channels {
		responses[i] = c.ToResponse()
	}

	return responses, total, nil
}

// GetChannel получает канал уведомлений по ID
func (s *NotificationChannelService) GetChannel(id uint) (*models.NotificationChannelResponse, error) {
	channel, err := s.channelRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("канал уведомлений не найден")
		}
		return nil, fmt.Errorf("ошибка получения канала уведомлений: %w", err)
	}

	response := channel.ToResponse()
	return &response, nil
}

// UpdateChannel обновляет канал уведомлений
func (s *NotificationChannelService) UpdateChannel(id uint, req *models.NotificationChannelUpdateRequest) (*models.NotificationChannelResponse, error) {
	channel, err := s.channelRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("канал уведомлений не найден")
		}
		return nil, fmt.Errorf("ошибка получения канала уведомлений: %w", err)
	}

	if req.Name != "" {
		channel.Name = req.Name
	}
	if req.Type != "" {
		channel.Type = req.Type
	}
	if req.Config != "" {
		channel.Config = req.Config
	}
	channel.IsActive = req.IsActive

	if err := s.channelRepo.Update(channel); err != nil {
		return nil, fmt.Errorf("ошибка обновления канала уведомлений: %w", err)
	}

	response := channel.ToResponse()
	return &response, nil
}

// DeleteChannel удаляет канал уведомлений
func (s *NotificationChannelService) DeleteChannel(id uint) error {
	if err := s.channelRepo.Delete(id); err != nil {
		return fmt.Errorf("ошибка удаления канала уведомлений: %w", err)
	}
	return nil
}
