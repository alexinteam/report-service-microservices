# 📊 Дашборды Grafana

Эта папка содержит JSON файлы дашбордов для Grafana.

## 📈 Доступные дашборды

### 1. Microservices Overview (`microservices-overview-dashboard.json`)
**Общий обзор всех микросервисов**
- HTTP Requests Rate - скорость HTTP запросов
- HTTP Response Time - время ответа
- HTTP Requests by Status Code - запросы по кодам статуса
- Service Health - состояние сервисов
- Memory Usage - использование памяти

### 2. Business Metrics (`business-metrics-dashboard.json`)
**Бизнес-метрики и операции с базой данных**
- Business Operations Rate - скорость бизнес-операций
- Business Operation Duration - длительность операций
- Business Operations Success Rate - процент успешных операций
- Database Operations - операции с БД
- Database Query Duration - время выполнения запросов

### 3. Saga Metrics (`saga-metrics-dashboard.json`)
**Метрики Saga паттерна для распределенных транзакций**
- Saga Transactions Started/Completed/Failed - статистика транзакций
- Saga Step Executions - выполнение шагов Saga
- Saga Compensation Rate - частота компенсаций
- Saga Duration - длительность Saga транзакций

### 4. Service Details (`service-details-dashboard.json`)
**Детальный мониторинг каждого сервиса**
- Переменная для выбора сервиса
- HTTP Requests Rate по выбранному сервису
- Response Time по выбранному сервису
- Business Operations по выбранному сервису
- Database Operations по выбранному сервису
- Memory Usage по выбранному сервису

### 5. Alerts & Health (`alerts-health-dashboard.json`)
**Алерты и мониторинг состояния здоровья**
- Service Health Status - статус сервисов (зеленый/красный)
- Error Rate - уровень ошибок с порогами
- Response Time Alert - алерты по времени ответа
- Memory Usage Alert - алерты по использованию памяти
- Database Errors - ошибки базы данных

## 📊 Метрики

Дашборды используют следующие метрики:

### HTTP метрики
- `http_requests_total` - общее количество HTTP запросов
- `http_request_duration_seconds` - длительность HTTP запросов

### Бизнес метрики
- `business_operations_total` - бизнес-операции
- `business_operation_duration_seconds` - длительность операций

### База данных
- `database_operations_total` - операции с БД
- `database_query_duration_seconds` - время запросов
- `database_errors_total` - ошибки БД

### Saga метрики
- `saga_transactions_started_total` - начатые транзакции
- `saga_transactions_completed_total` - завершенные транзакции
- `saga_transactions_failed_total` - неудачные транзакции
- `saga_step_executions_total` - выполнение шагов
- `saga_compensations_total` - компенсации
- `saga_duration_seconds` - длительность Saga

### Системные метрики
- `memory_usage_bytes` - использование памяти
- `cpu_usage` - использование CPU
