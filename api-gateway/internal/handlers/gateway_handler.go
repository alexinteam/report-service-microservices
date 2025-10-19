package handlers

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"api-gateway/internal/config"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type GatewayHandler struct {
	config *config.Config
}

func NewGatewayHandler(cfg *config.Config) *GatewayHandler {
	return &GatewayHandler{
		config: cfg,
	}
}

func (h *GatewayHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "api-gateway",
		"version": "1.0.0",
	})
}

// ProxyToUserService проксирование запросов к User Service
func (h *GatewayHandler) ProxyToUserService(c *gin.Context) {
	h.proxyRequest(c, h.config.UserServiceURL)
}

// ProxyToTemplateService проксирование запросов к Template Service
func (h *GatewayHandler) ProxyToTemplateService(c *gin.Context) {
	h.proxyRequest(c, h.config.TemplateServiceURL)
}

// ProxyToReportService проксирование запросов к Report Service
func (h *GatewayHandler) ProxyToReportService(c *gin.Context) {
	h.proxyRequest(c, h.config.ReportServiceURL)
}

// ProxyToDataService проксирование запросов к Data Service
func (h *GatewayHandler) ProxyToDataService(c *gin.Context) {
	h.proxyRequest(c, h.config.DataServiceURL)
}

// ProxyToNotificationService проксирование запросов к Notification Service
func (h *GatewayHandler) ProxyToNotificationService(c *gin.Context) {
	h.proxyRequest(c, h.config.NotificationServiceURL)
}

// ProxyToStorageService проксирование запросов к Storage Service
func (h *GatewayHandler) ProxyToStorageService(c *gin.Context) {
	h.proxyRequest(c, h.config.StorageServiceURL)
}

// proxyRequest общий метод для проксирования запросов
func (h *GatewayHandler) proxyRequest(c *gin.Context, targetURL string) {
	path := c.Param("path")
	if path == "" {
		path = c.Request.URL.Path
	} else {
		path = c.Request.URL.Path
	}

	if path == "/api/v1/data-sources" || path == "/api/v1/storage/files" {
		path += "/"
	}

	if strings.HasPrefix(path, "/api/v1/storage/") {
		path = strings.Replace(path, "/api/v1/storage/", "/api/v1/", 1)
	}

	if path == "/api/v1/files" {
		path += "/"
	}

	fullURL := targetURL + path

	if c.Request.URL.RawQuery != "" {
		fullURL += "?" + c.Request.URL.RawQuery
	}

	logrus.WithFields(logrus.Fields{
		"original_path":  c.Request.URL.Path,
		"processed_path": path,
		"target_url":     targetURL,
		"full_url":       fullURL,
		"method":         c.Request.Method,
	}).Info("Proxying request")

	var body io.Reader
	if c.Request.Body != nil {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка чтения тела запроса"})
			return
		}
		body = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(c.Request.Method, fullURL, body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка создания запроса"})
		return
	}

	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Ошибка проксирования запроса"})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Ошибка чтения ответа"})
		return
	}

	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	c.Status(resp.StatusCode)

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}
