package handlers

import (
	"net/http"
	"strconv"
	"time"

	"report-service/internal/events"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SagaHandler обработчик для Saga операций
type SagaHandler struct {
	sagaCoordinator *events.IdempotentSagaCoordinator
	stateStore      *events.SagaStateStore
}

// NewSagaHandler создает новый обработчик Saga
func NewSagaHandler(sagaCoordinator *events.IdempotentSagaCoordinator, stateStore *events.SagaStateStore) *SagaHandler {
	return &SagaHandler{
		sagaCoordinator: sagaCoordinator,
		stateStore:      stateStore,
	}
}

// CreateReportSagaRequest запрос на создание Saga для отчета
type CreateReportSagaRequest struct {
	TemplateID string                 `json:"template_id" binding:"required"`
	Parameters map[string]interface{} `json:"parameters"`
}

// CreateReportSaga создает новую Saga для создания отчета
func (h *SagaHandler) CreateReportSaga(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	var req CreateReportSagaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Создаем идемпотентную Saga
	idempotentSaga := events.NewIdempotentReportCreationSaga(
		"0", // reportID будет создан позже
		strconv.FormatUint(uint64(userID.(uint)), 10),
		req.TemplateID,
		req.Parameters,
	)

	// Создаем обычную Saga для сохранения в базе данных
	saga := &events.Saga{
		ID:        idempotentSaga.ID,
		Name:      "Idempotent Report Creation Saga",
		Status:    events.SagaStatusPending,
		Steps:     idempotentSaga.Steps,
		Data:      map[string]interface{}{"template_id": req.TemplateID, "parameters": req.Parameters},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Запускаем Saga и сохраняем состояние
	ctx := c.Request.Context()
	if err := h.sagaCoordinator.StartSaga(ctx, saga); err != nil {
		logrus.WithError(err).Errorf("Ошибка запуска Saga %s", saga.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания Saga"})
		return
	}

	// Запускаем выполнение Saga асинхронно
	go func() {
		if err := idempotentSaga.Execute(ctx, h.sagaCoordinator); err != nil {
			logrus.WithError(err).Errorf("Ошибка выполнения Saga %s", saga.ID)
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Saga создания отчета запущена",
		"saga_id": saga.ID,
		"status":  "started",
	})
}

// GetSagaStatus получает статус Saga
func (h *SagaHandler) GetSagaStatus(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	sagaID := c.Param("id")
	if sagaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID Saga не указан"})
		return
	}

	// Получаем состояние Saga
	saga, err := h.sagaCoordinator.GetSaga(c.Request.Context(), sagaID)
	if err != nil {
		logrus.WithError(err).Errorf("Ошибка получения Saga %s", sagaID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Saga не найдена"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"saga_id":      saga.ID,
		"status":       saga.Status,
		"steps":        saga.Steps,
		"created_at":   saga.CreatedAt,
		"updated_at":   saga.UpdatedAt,
		"completed_at": saga.CompletedAt,
		"error":        saga.Error,
	})
}

// GetSagaProgress получает прогресс выполнения Saga
func (h *SagaHandler) GetSagaProgress(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	sagaID := c.Param("id")
	if sagaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID Saga не указан"})
		return
	}

	// Создаем временную Saga для получения прогресса
	tempSaga := &events.IdempotentReportCreationSaga{ID: sagaID}

	progress, err := tempSaga.GetSagaProgress(c.Request.Context(), h.sagaCoordinator)
	if err != nil {
		logrus.WithError(err).Errorf("Ошибка получения прогресса Saga %s", sagaID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Saga не найдена"})
		return
	}

	c.JSON(http.StatusOK, progress)
}

// RetrySaga повторяет выполнение неудачной Saga
func (h *SagaHandler) RetrySaga(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	sagaID := c.Param("id")
	if sagaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID Saga не указан"})
		return
	}

	// Получаем текущее состояние Saga
	saga, err := h.sagaCoordinator.GetSaga(c.Request.Context(), sagaID)
	if err != nil {
		logrus.WithError(err).Errorf("Ошибка получения Saga %s", sagaID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Saga не найдена"})
		return
	}

	// Проверяем, что Saga действительно неудачная
	if saga.Status != events.SagaStatusFailed {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":          "Saga не в статусе Failed",
			"current_status": saga.Status,
		})
		return
	}

	// Создаем временную Saga для повторного выполнения
	tempSaga := &events.IdempotentReportCreationSaga{ID: sagaID}

	// Запускаем повторное выполнение асинхронно
	go func() {
		ctx := c.Request.Context()
		if err := tempSaga.RetryFailedSaga(ctx, h.sagaCoordinator); err != nil {
			logrus.WithError(err).Errorf("Ошибка повторного выполнения Saga %s", sagaID)
		}
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Повторное выполнение Saga запущено",
		"saga_id": sagaID,
		"status":  "retrying",
	})
}

// CancelSaga отменяет выполнение Saga
func (h *SagaHandler) CancelSaga(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	sagaID := c.Param("id")
	if sagaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID Saga не указан"})
		return
	}

	// Получаем текущее состояние Saga
	saga, err := h.sagaCoordinator.GetSaga(c.Request.Context(), sagaID)
	if err != nil {
		logrus.WithError(err).Errorf("Ошибка получения Saga %s", sagaID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Saga не найдена"})
		return
	}

	// Проверяем, что Saga можно отменить
	if saga.Status == events.SagaStatusCompleted {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":          "Нельзя отменить завершенную Saga",
			"current_status": saga.Status,
		})
		return
	}

	// Обновляем статус Saga на Failed для запуска компенсации
	if err := h.sagaCoordinator.UpdateSagaStatus(c.Request.Context(), sagaID, events.SagaStatusFailed); err != nil {
		logrus.WithError(err).Errorf("Ошибка отмены Saga %s", sagaID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка отмены Saga"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Saga отменена",
		"saga_id": sagaID,
		"status":  "cancelled",
	})
}

// ListSagas получает список Saga пользователя
func (h *SagaHandler) ListSagas(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не авторизован"})
		return
	}

	// Получаем параметры запроса
	_ = c.Query("status")
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный параметр page"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный параметр limit"})
		return
	}

	// Здесь должна быть логика получения списка Saga из базы данных
	// Пока возвращаем заглушку
	c.JSON(http.StatusOK, gin.H{
		"sagas": []gin.H{},
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": 0,
		},
	})
}

// ForceCompleteSaga принудительно завершает Saga
func (h *SagaHandler) ForceCompleteSaga(c *gin.Context) {
	sagaID := c.Param("id")
	if sagaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Saga ID required"})
		return
	}

	ctx := c.Request.Context()
	if err := h.sagaCoordinator.ForceCompleteSaga(ctx, sagaID); err != nil {
		logrus.WithError(err).Errorf("Ошибка принудительного завершения Saga %s", sagaID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка завершения Saga"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Saga принудительно завершена",
		"saga_id": sagaID,
		"status":  "completed",
	})
}
