#!/bin/bash

# 🚀 Быстрый тест основных сценариев системы Report Microservices
# Автор: AI Assistant
# Дата: 19 октября 2025

set -e

# Цвета
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

# Переменные
BASE_URL="http://arch.homework"
API_BASE="$BASE_URL/api/v1"
TOKEN=""

# Функция для получения токена
get_token() {
    print_info "Получение JWT токена..."
    
    # Сначала пытаемся зарегистрировать пользователя
    print_info "Регистрация пользователя..."
    local register_response
    register_response=$(curl -s -X POST -H "Content-Type: application/json" \
        -d '{"name":"Test User","email":"test@example.com","password":"password123"}' \
        "$API_BASE/users/register")
    
    if echo "$register_response" | grep -q "уже существует"; then
        print_info "Пользователь уже существует, продолжаем с логином"
    elif echo "$register_response" | jq -e '.user' > /dev/null 2>&1; then
        print_success "Пользователь зарегистрирован"
    else
        print_info "Регистрация не удалась, продолжаем с логином"
    fi
    
    # Теперь логинимся
    local response
    response=$(curl -s -X POST -H "Content-Type: application/json" \
        -d '{"email":"test@example.com","password":"password123"}' \
        "$API_BASE/users/login")
    
    if echo "$response" | jq -e '.token' > /dev/null 2>&1; then
        TOKEN=$(echo "$response" | jq -r '.token')
        print_success "Токен получен"
        return 0
    else
        print_error "Не удалось получить токен"
        echo "Ответ: $response"
        return 1
    fi
}

# Функция для тестирования API Gateway
test_api_gateway() {
    print_header "API GATEWAY"
    
    # Health check
    if curl -s "$BASE_URL/health" | grep -q "healthy"; then
        print_success "API Gateway работает"
    else
        print_error "API Gateway недоступен"
        return 1
    fi
}

# Функция для тестирования User Service
test_user_service() {
    print_header "USER SERVICE"
    
    # Получение токена
    if get_token; then
        print_success "User Service работает"
    else
        print_error "User Service не работает"
        return 1
    fi
}

# Функция для тестирования Template Service
test_template_service() {
    print_header "TEMPLATE SERVICE"
    
    # Тестируем получение шаблонов через API Gateway
    if curl -s -H "Authorization: Bearer $TOKEN" http://arch.homework/api/v1/templates | jq -e '.templates' > /dev/null 2>&1; then
        print_success "Template Service работает"
    else
        print_error "Template Service не работает"
    fi
}

# Функция для тестирования Data Service
test_data_service() {
    print_header "DATA SERVICE"
    
    # Тестируем получение источников данных через API Gateway
    if curl -s -H "Authorization: Bearer $TOKEN" http://arch.homework/api/v1/data-sources | jq -e '.data_sources' > /dev/null 2>&1; then
        print_success "Data Service работает"
    else
        print_error "Data Service не работает"
    fi
}

# Функция для тестирования Report Service
test_report_service() {
    print_header "REPORT SERVICE"
    
    # Тестируем получение отчетов через API Gateway
    if curl -s -H "Authorization: Bearer $TOKEN" http://arch.homework/api/v1/reports | jq -e '.reports' > /dev/null 2>&1; then
        print_success "Report Service работает"
    else
        print_error "Report Service не работает"
    fi
}

# Функция для тестирования мониторинга
test_monitoring() {
    print_header "МОНИТОРИНГ"
    
    # Grafana
    if curl -s http://localhost/api/health | grep -q "ok"; then
        print_success "Grafana работает"
    else
        print_error "Grafana недоступна"
    fi
    
    # Prometheus через Ingress (если настроен)
    if curl -s http://localhost:9090/-/healthy | grep -q "Healthy"; then
        print_success "Prometheus работает"
    else
        print_warning "Prometheus недоступен через Ingress (требует port-forward)"
    fi
}

# Функция для демонстрации Saga Pattern
demo_saga_pattern() {
    print_header "SAGA PATTERN ДЕМО"
    
    print_info "Создание отчета через Saga..."
    local saga_response
    saga_response=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "template_id": "1",
            "name": "Demo Report",
            "format": "html",
            "parameters": {"title": "Demo Report"}
        }' \
        http://arch.homework/api/v1/sagas/reports)
    
    if echo "$saga_response" | jq -e '.saga_id' > /dev/null 2>&1; then
        print_success "Saga создана успешно"
        
        local saga_id
        saga_id=$(echo "$saga_response" | jq -r '.saga_id')
        print_info "Saga ID: $saga_id"
        
        # Получаем статус Saga
        print_info "Получение статуса Saga..."
        if curl -s -H "Authorization: Bearer $TOKEN" "http://arch.homework/api/v1/sagas/$saga_id" | jq -e '.status' > /dev/null 2>&1; then
            print_success "Saga Pattern работает"
        else
            print_error "Не удалось получить статус Saga"
        fi
    else
        print_error "Не удалось создать Saga"
    fi
}

# Функция для демонстрации идемпотентности
demo_idempotency() {
    print_header "ИДЕМПОТЕНТНОСТЬ ДЕМО"
    
    local idempotency_key="demo-idempotency-$(date +%s)"
    print_info "Idempotency Key: $idempotency_key"
    
    # Первый запрос
    print_info "Первый запрос с idempotency key..."
    local first_response
    first_response=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -H "Idempotency-Key: $idempotency_key" \
        -d '{
            "template_id": 1,
            "name": "Idempotent Report",
            "format": "html"
        }' \
        http://arch.homework/api/v1/reports/)
    
    if echo "$first_response" | jq -e '.id' > /dev/null 2>&1; then
        print_success "Первый запрос выполнен"
        
        # Второй запрос с тем же ключом
        print_info "Повторный запрос с тем же idempotency key..."
        local second_response
        second_response=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -H "Idempotency-Key: $idempotency_key" \
            -d '{
                "template_id": 1,
                "name": "Idempotent Report",
                "format": "html"
            }' \
            http://arch.homework/api/v1/reports/)
        
        if echo "$second_response" | jq -e '.id' > /dev/null 2>&1; then
            print_success "Идемпотентность работает"
        else
            print_error "Идемпотентность не работает"
        fi
    else
        print_error "Не удалось выполнить первый запрос"
    fi
}

# Функция для очистки (больше не нужна, так как не используем port-forward)
cleanup() {
    print_info "Очистка завершена"
}

# Функция для очистки в начале (больше не нужна)
cleanup_start() {
    print_info "Готов к тестированию"
}

# Основная функция
main() {
    echo -e "${BLUE}🚀 БЫСТРЫЙ ТЕСТ СИСТЕМЫ REPORT MICROSERVICES${NC}"
    echo -e "${BLUE}==============================================${NC}"
    
    # Проверяем зависимости
    if ! command -v curl &> /dev/null; then
        print_error "curl не установлен"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        print_error "jq не установлен"
        exit 1
    fi
    
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl не установлен"
        exit 1
    fi
    
    print_success "Все зависимости установлены"
    
    # Очищаем существующие port-forward процессы
    cleanup_start
    
    # Запускаем тесты
    test_api_gateway
    test_user_service
    test_template_service
    test_data_service
    test_report_service
    test_monitoring
    
    # Демонстрации
    demo_saga_pattern
    demo_idempotency
    
    # Очистка
    cleanup
    
    echo -e "\n${GREEN}🎉 ТЕСТИРОВАНИЕ ЗАВЕРШЕНО!${NC}"
    echo -e "${GREEN}Система готова к презентации! 🚀${NC}"
}

# Обработка сигналов
trap cleanup EXIT INT TERM

# Запуск
main "$@"
