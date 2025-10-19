package handlers

import (
	"net/http"
	"strconv"
	"time"

	"data-service/internal/metrics"
	"data-service/internal/models"
	"data-service/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type DataSourceHandler struct {
	dataSourceService *services.DataSourceService
	metrics           *metrics.Metrics
}

func NewDataSourceHandler(dataSourceService *services.DataSourceService, metrics *metrics.Metrics) *DataSourceHandler {
	return &DataSourceHandler{
		dataSourceService: dataSourceService,
		metrics:           metrics,
	}
}

func (h *DataSourceHandler) CreateDataSource(c *gin.Context) {
	start := time.Now()
	var req models.DataSourceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.metrics.RecordBusinessOperation("data-service", "create_data_source", time.Since(start), false)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dataSource, err := h.dataSourceService.CreateDataSource(&req)
	if err != nil {
		logrus.WithError(err).Error("Ошибка создания источника данных")
		h.metrics.RecordBusinessOperation("data-service", "create_data_source", time.Since(start), false)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.metrics.RecordBusinessOperation("data-service", "create_data_source", time.Since(start), true)
	c.JSON(http.StatusCreated, dataSource)
}

func (h *DataSourceHandler) GetDataSources(c *gin.Context) {
	start := time.Now()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	active := c.Query("active")

	dataSources, total, err := h.dataSourceService.GetDataSources(page, limit, active)
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения источников данных")
		h.metrics.RecordBusinessOperation("data-service", "get_data_sources", time.Since(start), false)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.metrics.RecordBusinessOperation("data-service", "get_data_sources", time.Since(start), true)
	c.JSON(http.StatusOK, models.DataSourcesResponse{
		DataSources: dataSources,
		Total:       total,
		Page:        page,
		Limit:       limit,
	})
}

func (h *DataSourceHandler) GetDataSource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	dataSource, err := h.dataSourceService.GetDataSource(uint(id))
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения источника данных")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dataSource)
}

func (h *DataSourceHandler) UpdateDataSource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	var req models.DataSourceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dataSource, err := h.dataSourceService.UpdateDataSource(uint(id), &req)
	if err != nil {
		logrus.WithError(err).Error("Ошибка обновления источника данных")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dataSource)
}

func (h *DataSourceHandler) DeleteDataSource(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	if err := h.dataSourceService.DeleteDataSource(uint(id)); err != nil {
		logrus.WithError(err).Error("Ошибка удаления источника данных")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

type DataCollectionHandler struct {
	dataCollectionService *services.DataCollectionService
}

func NewDataCollectionHandler(dataCollectionService *services.DataCollectionService) *DataCollectionHandler {
	return &DataCollectionHandler{
		dataCollectionService: dataCollectionService,
	}
}

func (h *DataCollectionHandler) CreateDataCollection(c *gin.Context) {
	var req models.DataCollectionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dataCollection, err := h.dataCollectionService.CreateDataCollection(&req)
	if err != nil {
		logrus.WithError(err).Error("Ошибка создания сбора данных")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, dataCollection)
}

func (h *DataCollectionHandler) GetDataCollections(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	active := c.Query("active")

	dataCollections, total, err := h.dataCollectionService.GetDataCollections(page, limit, active)
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения сборов данных")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.DataCollectionsResponse{
		DataCollections: dataCollections,
		Total:           total,
		Page:            page,
		Limit:           limit,
	})
}

func (h *DataCollectionHandler) GetDataCollection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	dataCollection, err := h.dataCollectionService.GetDataCollection(uint(id))
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения сбора данных")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dataCollection)
}

func (h *DataCollectionHandler) UpdateDataCollection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	var req models.DataCollectionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dataCollection, err := h.dataCollectionService.UpdateDataCollection(uint(id), &req)
	if err != nil {
		logrus.WithError(err).Error("Ошибка обновления сбора данных")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dataCollection)
}

func (h *DataCollectionHandler) DeleteDataCollection(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	if err := h.dataCollectionService.DeleteDataCollection(uint(id)); err != nil {
		logrus.WithError(err).Error("Ошибка удаления сбора данных")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

type CollectDataHandler struct {
	collectDataService *services.CollectDataService
}

func NewCollectDataHandler(collectDataService *services.CollectDataService) *CollectDataHandler {
	return &CollectDataHandler{
		collectDataService: collectDataService,
	}
}

func (h *CollectDataHandler) CollectData(c *gin.Context) {
	var req models.DataCollectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.collectDataService.CollectData(&req)
	if err != nil {
		logrus.WithError(err).Error("Ошибка сбора данных")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *CollectDataHandler) GetDataRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	collectionIDStr := c.Query("collection_id")

	var collectionID uint
	if collectionIDStr != "" {
		id, err := strconv.ParseUint(collectionIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный collection_id"})
			return
		}
		collectionID = uint(id)
	}

	dataRecords, total, err := h.collectDataService.GetDataRecords(page, limit, collectionID)
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения записей данных")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.DataRecordsResponse{
		DataRecords: dataRecords,
		Total:       total,
		Page:        page,
		Limit:       limit,
	})
}

func (h *CollectDataHandler) GetDataRecord(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID"})
		return
	}

	dataRecord, err := h.collectDataService.GetDataRecord(uint(id))
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения записи данных")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dataRecord)
}
