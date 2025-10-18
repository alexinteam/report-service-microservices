package metrics

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics содержит все метрики для микросервиса
type Metrics struct {
	// HTTP метрики
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight *prometheus.GaugeVec

	// Бизнес метрики
	BusinessOperationsTotal   *prometheus.CounterVec
	BusinessOperationDuration *prometheus.HistogramVec

	// База данных метрики
	DatabaseConnections   *prometheus.GaugeVec
	DatabaseQueryDuration *prometheus.HistogramVec
	DatabaseErrorsTotal   *prometheus.CounterVec

	// Системные метрики
	MemoryUsage       *prometheus.GaugeVec
	CPUUsage          *prometheus.GaugeVec
	ActiveConnections *prometheus.GaugeVec
}

// NewMetrics создает новый экземпляр метрик для сервиса
func NewMetrics(serviceName string) *Metrics {
	return &Metrics{
		// HTTP метрики
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"service", "method", "endpoint", "status_code"},
		),

		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"service", "method", "endpoint"},
		),

		HTTPRequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "Number of HTTP requests currently being processed",
			},
			[]string{"service"},
		),

		// Бизнес метрики
		BusinessOperationsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "business_operations_total",
				Help: "Total number of business operations",
			},
			[]string{"service", "operation", "status"},
		),

		BusinessOperationDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "business_operation_duration_seconds",
				Help:    "Business operation duration in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"service", "operation"},
		),

		// База данных метрики
		DatabaseConnections: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "database_connections_active",
				Help: "Number of active database connections",
			},
			[]string{"service", "database"},
		),

		DatabaseQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "database_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
			},
			[]string{"service", "operation"},
		),

		DatabaseErrorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "database_errors_total",
				Help: "Total number of database errors",
			},
			[]string{"service", "operation", "error_type"},
		),

		// Системные метрики
		MemoryUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "memory_usage_bytes",
				Help: "Memory usage in bytes",
			},
			[]string{"service"},
		),

		CPUUsage: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "cpu_usage_percent",
				Help: "CPU usage percentage",
			},
			[]string{"service"},
		),

		ActiveConnections: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "active_connections",
				Help: "Number of active connections",
			},
			[]string{"service", "connection_type"},
		),
	}
}

// HTTPMiddleware создает middleware для сбора HTTP метрик
func (m *Metrics) HTTPMiddleware(serviceName string) gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		// Увеличиваем счетчик активных запросов
		m.HTTPRequestsInFlight.WithLabelValues(serviceName).Inc()
		defer m.HTTPRequestsInFlight.WithLabelValues(serviceName).Dec()

		// Обрабатываем запрос
		c.Next()

		// Записываем метрики
		duration := time.Since(start).Seconds()
		status := c.Writer.Status()
		statusStr := http.StatusText(status)

		m.HTTPRequestsTotal.WithLabelValues(
			serviceName,
			c.Request.Method,
			c.FullPath(),
			statusStr,
		).Inc()

		m.HTTPRequestDuration.WithLabelValues(
			serviceName,
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)
	})
}

// RecordBusinessOperation записывает метрики бизнес-операции
func (m *Metrics) RecordBusinessOperation(serviceName, operation string, duration time.Duration, success bool) {
	status := "success"
	if !success {
		status = "error"
	}

	m.BusinessOperationsTotal.WithLabelValues(serviceName, operation, status).Inc()
	m.BusinessOperationDuration.WithLabelValues(serviceName, operation).Observe(duration.Seconds())
}

// RecordDatabaseOperation записывает метрики операции с базой данных
func (m *Metrics) RecordDatabaseOperation(serviceName, operation string, duration time.Duration, err error) {
	m.DatabaseQueryDuration.WithLabelValues(serviceName, operation).Observe(duration.Seconds())

	if err != nil {
		errorType := "unknown"
		// Можно добавить более детальную классификацию ошибок
		m.DatabaseErrorsTotal.WithLabelValues(serviceName, operation, errorType).Inc()
	}
}

// SetupMetricsEndpoint настраивает endpoint для метрик
func (m *Metrics) SetupMetricsEndpoint(router *gin.Engine, serviceName string) {
	// Добавляем middleware для HTTP метрик
	router.Use(m.HTTPMiddleware(serviceName))

	// Endpoint для метрик Prometheus
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
