# Report Service - Асинхронное создание отчетов

## Обзор

Report Service теперь поддерживает асинхронное создание отчетов с использованием Saga Pattern. Отчеты создаются в фоновом режиме, и клиенты могут отслеживать их статус.

## API Endpoints

### 1. Создание отчета (асинхронно)

**POST** `/api/v1/reports/`

```json
{
  "name": "Ежемесячный отчет по продажам",
  "description": "Отчет за январь 2024",
  "template_id": 1,
  "parameters": "{\"period\": \"2024-01\", \"department\": \"sales\"}"
}
```

**Ответ:**
```json
{
  "id": 123,
  "status": "pending",
  "message": "Отчет создан и поставлен в очередь на генерацию"
}
```

### 2. Получение статуса отчета

**GET** `/api/v1/reports/{id}/status`

**Ответ для отчета в процессе:**
```json
{
  "id": 123,
  "status": "processing",
  "progress": 50
}
```

**Ответ для готового отчета:**
```json
{
  "id": 123,
  "status": "completed",
  "file_path": "/reports/report_123.pdf"
}
```

**Ответ для неудачного отчета:**
```json
{
  "id": 123,
  "status": "failed",
  "error": "Ошибка генерации отчета"
}
```

### 3. Получение полной информации об отчете

**GET** `/api/v1/reports/{id}`

**Ответ:**
```json
{
  "id": 123,
  "name": "Ежемесячный отчет по продажам",
  "description": "Отчет за январь 2024",
  "template_id": 1,
  "user_id": 1,
  "status": "completed",
  "parameters": "{\"period\": \"2024-01\", \"department\": \"sales\"}",
  "file_path": "/reports/report_123.pdf",
  "file_size": 1048576,
  "md5_hash": "hash_123",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:35:00Z"
}
```

## Статусы отчетов

- **`pending`** - Отчет создан и ожидает обработки
- **`processing`** - Отчет генерируется
- **`completed`** - Отчет готов, файл доступен
- **`failed`** - Ошибка при генерации отчета
- **`cancelled`** - Отчет отменен

## Saga Pattern

Создание отчета выполняется через Saga с следующими шагами:

1. **Validate User** - Валидация пользователя
2. **Validate Template** - Валидация шаблона
3. **Collect Data** - Сбор данных для отчета
4. **Generate Report** - Генерация отчета
5. **Store File** - Сохранение файла
6. **Send Notification** - Отправка уведомления

## Пример использования

### 1. Создание отчета

```bash
curl -X POST http://arch.homework:8083/api/v1/reports/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "name": "Отчет по продажам",
    "description": "Ежемесячный отчет",
    "template_id": 1,
    "parameters": "{\"month\": \"2024-01\"}"
  }'
```

**Ответ:**
```json
{
  "id": 123,
  "status": "pending",
  "message": "Отчет создан и поставлен в очередь на генерацию"
}
```

### 2. Проверка статуса

```bash
curl -X GET http://arch.homework:8083/api/v1/reports/123/status \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Через несколько секунд:**
```json
{
  "id": 123,
  "status": "processing",
  "progress": 50
}
```

**Когда готов:**
```json
{
  "id": 123,
  "status": "completed",
  "file_path": "/reports/report_123.pdf"
}
```

### 3. Скачивание отчета

```bash
curl -X GET http://arch.homework:8083/api/v1/reports/123/download \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Saga Management API

### Получение статуса Saga

**GET** `/api/v1/sagas/{saga_id}`

### Получение прогресса Saga

**GET** `/api/v1/sagas/{saga_id}/progress`

### Повторное выполнение Saga

**POST** `/api/v1/sagas/{saga_id}/retry`

### Отмена Saga

**DELETE** `/api/v1/sagas/{saga_id}`

## Преимущества нового подхода

1. **Асинхронность** - Клиенты не блокируются при создании больших отчетов
2. **Надежность** - Saga Pattern обеспечивает консистентность данных
3. **Идемпотентность** - Повторные запросы не создают дубликаты
4. **Отслеживание** - Клиенты могут отслеживать прогресс выполнения
5. **Компенсация** - При ошибках выполняется откат изменений

## Мониторинг

Все операции логируются и могут быть отслежены через:
- Логи сервиса
- Saga State Store
- Event Log
- Outbox Pattern для надежной публикации событий

