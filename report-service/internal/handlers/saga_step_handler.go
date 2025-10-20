package handlers

import (
	"context"
	"fmt"
	"strconv"

	"report-service/internal/events"
	"report-service/internal/models"
	"report-service/internal/services"

	"github.com/sirupsen/logrus"
)

// SagaStepHandler обработчик для выполнения шагов Saga
type SagaStepHandler struct {
	reportService  *services.ReportService
	eventPublisher events.EventPublisher
}

// NewSagaStepHandler создает новый обработчик шагов Saga
func NewSagaStepHandler(reportService *services.ReportService, eventPublisher events.EventPublisher) *SagaStepHandler {
	return &SagaStepHandler{
		reportService:  reportService,
		eventPublisher: eventPublisher,
	}
}

// ExecuteStep выполняет шаг Saga
func (h *SagaStepHandler) ExecuteStep(ctx context.Context, step *events.SagaStep) error {
	logrus.Infof("Выполняем шаг Saga: %s", step.Name)

	switch step.Service {
	case "report-service":
		return h.executeReportServiceStep(ctx, step)
	case "user-service":
		return h.executeUserServiceStep(ctx, step)
	case "template-service":
		return h.executeTemplateServiceStep(ctx, step)
	case "data-service":
		return h.executeDataServiceStep(ctx, step)
	case "storage-service":
		return h.executeStorageServiceStep(ctx, step)
	case "notification-service":
		return h.executeNotificationServiceStep(ctx, step)
	default:
		return fmt.Errorf("неизвестный сервис: %s", step.Service)
	}
}

// executeReportServiceStep выполняет шаги report-service
func (h *SagaStepHandler) executeReportServiceStep(ctx context.Context, step *events.SagaStep) error {
	switch step.Action {
	case "generate_report":
		return h.generateReport(ctx, step)
	case "update_status":
		return h.updateReportStatus(ctx, step)
	default:
		return fmt.Errorf("неизвестное действие для report-service: %s", step.Action)
	}
}

// generateReport генерирует отчет
func (h *SagaStepHandler) generateReport(ctx context.Context, step *events.SagaStep) error {
	// Получаем данные из шага
	templateIDStr, ok := step.Data["template_id"].(string)
	if !ok {
		return fmt.Errorf("отсутствует template_id в данных шага")
	}

	userIDStr, ok := step.Data["user_id"].(string)
	if !ok {
		return fmt.Errorf("отсутствует user_id в данных шага")
	}

	parameters, ok := step.Data["parameters"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("отсутствуют parameters в данных шага")
	}

	// Парсим ID
	templateID, err := strconv.ParseUint(templateIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("некорректный template_id: %w", err)
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return fmt.Errorf("некорректный user_id: %w", err)
	}

	// Создаем запрос на создание отчета
	createReq := &models.ReportCreateRequest{
		Name:        parameters["title"].(string),
		Description: "Отчет создан через сагу",
		TemplateID:  uint(templateID),
		Parameters:  fmt.Sprintf(`{"title":"%s"}`, parameters["title"].(string)),
	}

	// Сохраняем отчет в БД
	createdReport, err := h.reportService.CreateReport(uint(userID), createReq)
	if err != nil {
		return fmt.Errorf("ошибка создания отчета: %w", err)
	}

	// Обновляем report_id в данных шага для последующих шагов
	step.Data["report_id"] = strconv.FormatUint(uint64(createdReport.ID), 10)

	logrus.Infof("Отчет %d создан и статус установлен на processing", createdReport.ID)
	return nil
}

// updateReportStatus обновляет статус отчета
func (h *SagaStepHandler) updateReportStatus(ctx context.Context, step *events.SagaStep) error {
	status, ok := step.Data["status"].(string)
	if !ok {
		return fmt.Errorf("отсутствует status в данных шага")
	}

	// Получаем user_id из данных шага
	userID, _ := step.Data["user_id"].(string)
	if userID == "" {
		return fmt.Errorf("отсутствует user_id в данных шага")
	}

	// Получаем report_id из БД - ищем последний отчет для данного пользователя
	userIDUint, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		return fmt.Errorf("некорректный user_id: %w", err)
	}

	// Получаем последний отчет пользователя
	reportsResponse, err := h.reportService.GetReports(uint(userIDUint), "", 1, 1)
	if err != nil {
		return fmt.Errorf("ошибка получения отчетов пользователя: %w", err)
	}

	if len(reportsResponse.Reports) == 0 {
		return fmt.Errorf("отчеты пользователя %s не найдены", userID)
	}

	// Используем ID последнего отчета
	reportID := reportsResponse.Reports[0].ID

	if err := h.reportService.UpdateReportStatus(reportID, status); err != nil {
		return fmt.Errorf("ошибка обновления статуса отчета: %w", err)
	}

	logrus.Infof("Статус отчета %d обновлен на %s", reportID, status)
	return nil
}

// executeUserServiceStep выполняет шаги user-service
func (h *SagaStepHandler) executeUserServiceStep(ctx context.Context, step *events.SagaStep) error {
	switch step.Action {
	case "validate_user":
		// Здесь должна быть логика валидации пользователя
		// Пока просто логируем
		logrus.Info("Валидация пользователя выполнена")
		return nil
	default:
		return fmt.Errorf("неизвестное действие для user-service: %s", step.Action)
	}
}

// executeTemplateServiceStep выполняет шаги template-service
func (h *SagaStepHandler) executeTemplateServiceStep(ctx context.Context, step *events.SagaStep) error {
	switch step.Action {
	case "validate_template":
		// Здесь должна быть логика валидации шаблона
		// Пока просто логируем
		logrus.Info("Валидация шаблона выполнена")
		return nil
	default:
		return fmt.Errorf("неизвестное действие для template-service: %s", step.Action)
	}
}

// executeDataServiceStep выполняет шаги data-service
func (h *SagaStepHandler) executeDataServiceStep(ctx context.Context, step *events.SagaStep) error {
	switch step.Action {
	case "collect_data":
		// Здесь должна быть логика сбора данных
		// Пока просто логируем
		logrus.Info("Сбор данных выполнен")
		return nil
	default:
		return fmt.Errorf("неизвестное действие для data-service: %s", step.Action)
	}
}

// executeStorageServiceStep выполняет шаги storage-service
func (h *SagaStepHandler) executeStorageServiceStep(ctx context.Context, step *events.SagaStep) error {
	switch step.Action {
	case "store_file":
		reportIDStr, ok := step.Data["report_id"].(string)
		if !ok {
			return fmt.Errorf("отсутствует report_id в данных шага")
		}

		reportID, err := strconv.ParseUint(reportIDStr, 10, 32)
		if err != nil {
			return fmt.Errorf("некорректный report_id: %w", err)
		}

		// Симулируем сохранение файла
		filePath := fmt.Sprintf("/reports/report_%d.pdf", reportID)
		fileSize := int64(1024 * 1024) // 1MB
		md5Hash := fmt.Sprintf("hash_%d", reportID)

		// Обновляем отчет с путем к файлу
		if err := h.reportService.UpdateReportFilePath(uint(reportID), filePath, fileSize, md5Hash); err != nil {
			return fmt.Errorf("ошибка обновления пути к файлу: %w", err)
		}

		logrus.Infof("Файл отчета %d сохранен по пути %s", reportID, filePath)
		return nil
	default:
		return fmt.Errorf("неизвестное действие для storage-service: %s", step.Action)
	}
}

// executeNotificationServiceStep выполняет шаги notification-service
func (h *SagaStepHandler) executeNotificationServiceStep(ctx context.Context, step *events.SagaStep) error {
	switch step.Action {
	case "send_notification":
		// Получаем user_id из данных шага
		userID, _ := step.Data["user_id"].(string)

		// Получаем report_id из БД - ищем последний отчет для данного пользователя
		userIDUint, err := strconv.ParseUint(userID, 10, 32)
		if err != nil {
			return fmt.Errorf("некорректный user_id: %w", err)
		}

		// Получаем последний отчет пользователя
		reportsResponse, err := h.reportService.GetReports(uint(userIDUint), "", 1, 1)
		if err != nil {
			return fmt.Errorf("ошибка получения отчетов пользователя: %w", err)
		}

		if len(reportsResponse.Reports) == 0 {
			return fmt.Errorf("отчеты пользователя %s не найдены", userID)
		}

		// Используем ID последнего отчета
		reportID := strconv.FormatUint(uint64(reportsResponse.Reports[0].ID), 10)

		// Публикуем событие, которое прочитает notification-service
		event := events.NewEvent(events.ReportCompleted, "report-service", map[string]interface{}{
			"report_id": reportID,
			"user_id":   userID,
			"type":      "report_ready",
		})
		if err := h.eventPublisher.Publish(ctx, event); err != nil {
			return fmt.Errorf("ошибка публикации события уведомления: %w", err)
		}
		logrus.Infof("Событие ReportCompleted опубликовано для notification-service с report_id: %s", reportID)
		return nil
	default:
		return fmt.Errorf("неизвестное действие для notification-service: %s", step.Action)
	}
}

// CompensateStep выполняет компенсацию шага Saga
func (h *SagaStepHandler) CompensateStep(ctx context.Context, step *events.SagaStep) error {
	logrus.Infof("Компенсируем шаг Saga: %s", step.Name)

	switch step.Service {
	case "report-service":
		return h.compensateReportServiceStep(ctx, step)
	case "storage-service":
		return h.compensateStorageServiceStep(ctx, step)
	default:
		logrus.Infof("Компенсация для сервиса %s не требуется", step.Service)
		return nil
	}
}

// compensateReportServiceStep компенсирует шаги report-service
func (h *SagaStepHandler) compensateReportServiceStep(ctx context.Context, step *events.SagaStep) error {
	switch step.Action {
	case "generate_report":
		reportIDStr, ok := step.Data["report_id"].(string)
		if !ok {
			return fmt.Errorf("отсутствует report_id в данных шага")
		}

		reportID, err := strconv.ParseUint(reportIDStr, 10, 32)
		if err != nil {
			return fmt.Errorf("некорректный report_id: %w", err)
		}

		// Обновляем статус на failed
		if err := h.reportService.UpdateReportStatus(uint(reportID), string(models.StatusFailed)); err != nil {
			return fmt.Errorf("ошибка обновления статуса на failed: %w", err)
		}

		logrus.Infof("Статус отчета %d обновлен на failed (компенсация)", reportID)
		return nil
	default:
		return fmt.Errorf("неизвестное действие для компенсации report-service: %s", step.Action)
	}
}

// compensateStorageServiceStep компенсирует шаги storage-service
func (h *SagaStepHandler) compensateStorageServiceStep(ctx context.Context, step *events.SagaStep) error {
	switch step.Action {
	case "store_file":
		// Здесь должна быть логика удаления файла
		// Пока просто логируем
		logrus.Info("Файл удален (компенсация)")
		return nil
	default:
		return fmt.Errorf("неизвестное действие для компенсации storage-service: %s", step.Action)
	}
}
