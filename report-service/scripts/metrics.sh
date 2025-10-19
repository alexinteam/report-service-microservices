#!/bin/bash

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${YELLOW}ℹ️  $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Функция для генерации всех метрик
generate_all_metrics() {
    print_header "ГЕНЕРАЦИЯ ВСЕХ МЕТРИК ДЛЯ GRAFANA"
    
    # Получаем токен
    print_info "Получение JWT токена..."
    TOKEN=$(curl -s -X POST -H "Content-Type: application/json" \
        -d '{"email":"test@example.com","password":"password123"}' \
        http://arch.homework/api/v1/users/login | jq -r '.token')
    
    if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
        print_error "Не удалось получить токен"
        return 1
    fi
    
    print_success "Токен получен"
    
    # Генерируем HTTP метрики
    print_info "Генерация HTTP метрик..."
    
    # User Service метрики
    for i in {1..10}; do
        curl -s -H "Authorization: Bearer $TOKEN" http://arch.homework/api/v1/users/profile >/dev/null
        sleep 0.1
    done
    
    # Template Service метрики
    for i in {1..8}; do
        curl -s -H "Authorization: Bearer $TOKEN" http://arch.homework/api/v1/templates >/dev/null
        sleep 0.1
    done
    
    # Data Service метрики
    for i in {1..6}; do
        curl -s -H "Authorization: Bearer $TOKEN" http://arch.homework/api/v1/data-sources >/dev/null
        sleep 0.1
    done
    
    # Report Service метрики
    for i in {1..8}; do
        curl -s -H "Authorization: Bearer $TOKEN" http://arch.homework/api/v1/reports >/dev/null
        sleep 0.1
    done
    
    print_success "HTTP метрики сгенерированы"
    
    # Генерируем бизнес-метрики
    print_info "Генерация бизнес-метрик..."
    
    # Регистрация пользователей
    for i in {1..5}; do
        curl -s -X POST -H "Content-Type: application/json" \
            -d "{\"name\":\"Test User $i\",\"email\":\"test$i@example.com\",\"password\":\"password123\"}" \
            http://arch.homework/api/v1/users/register >/dev/null
        sleep 0.2
    done
    
    # Логин пользователей
    for i in {1..8}; do
        curl -s -X POST -H "Content-Type: application/json" \
            -d '{"email":"test@example.com","password":"password123"}' \
            http://arch.homework/api/v1/users/login >/dev/null
        sleep 0.2
    done
    
    # Создание шаблонов
    for i in {1..5}; do
        curl -s -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
            -d "{\"name\":\"Test Template $i\",\"description\":\"Test template $i\",\"content\":\"<h1>Test Template $i</h1>\",\"type\":\"html\",\"category\":\"test\"}" \
            http://arch.homework/api/v1/templates >/dev/null
        sleep 0.2
    done
    
    # Создание источников данных
    for i in {1..4}; do
        curl -s -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
            -d "{\"name\":\"Test Data Source $i\",\"type\":\"database\",\"connection_string\":\"test\",\"description\":\"Test data source $i\"}" \
            http://arch.homework/api/v1/data-sources >/dev/null
        sleep 0.2
    done
    
    # Создание отчетов
    for i in {1..6}; do
        curl -s -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
            -d "{\"template_id\": 1,\"name\": \"Test Report $i\",\"format\": \"html\",\"description\": \"Test report $i\",\"parameters\": \"{\\\"title\\\": \\\"Test Report $i\\\"}\"}" \
            http://arch.homework/api/v1/reports >/dev/null
        sleep 0.2
    done
    
    print_success "Бизнес-метрики сгенерированы"
    
    # Генерируем Saga метрики
    print_info "Генерация Saga метрики..."
    
    saga_ids=()
    for i in {1..8}; do
        print_info "Создание Saga $i..."
        response=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
            -d "{\"template_id\": \"1\",\"name\": \"Quick Test Saga $i\",\"format\": \"html\",\"parameters\": {\"title\": \"Quick Test Saga $i\"}}" \
            http://arch.homework/api/v1/sagas/reports)
        
        saga_id=$(echo "$response" | jq -r '.saga_id')
        if [ "$saga_id" != "null" ] && [ -n "$saga_id" ]; then
            saga_ids+=("$saga_id")
            print_success "Saga $i создана: $saga_id"
        else
            print_error "Не удалось создать Saga $i"
        fi
        sleep 0.5
    done
    
    print_success "Saga операции созданы"
    
    # Ждем немного
    print_info "Ожидание обработки Saga..."
    sleep 3
    
    # Принудительно завершаем половину Saga
    print_info "Принудительное завершение Saga..."
    
    for i in "${!saga_ids[@]}"; do
        if [ $((i % 2)) -eq 0 ] && [ -n "${saga_ids[$i]}" ]; then
            saga_id="${saga_ids[$i]}"
            print_info "Завершение Saga: $saga_id"
            curl -s -X POST -H "Authorization: Bearer $TOKEN" \
                http://arch.homework/api/v1/sagas/$saga_id/force-complete >/dev/null
            sleep 0.3
        fi
    done
    
    print_success "Saga метрики сгенерированы"
    
    print_success "Все метрики сгенерированы!"
}

# Функция для проверки метрик
check_metrics() {
    print_header "ПРОВЕРКА МЕТРИК"
    
    # Проверяем Saga метрики
    print_info "Проверка Saga метрики..."
    saga_started=$(curl -s "http://localhost:9090/api/v1/query?query=business_operations_total%7Boperation%3D%22saga_started%22%7D" | jq '.data.result[0].value[1]' 2>/dev/null || echo "0")
    saga_completed=$(curl -s "http://localhost:9090/api/v1/query?query=business_operations_total%7Boperation%3D%22saga_completed%22%7D" | jq '.data.result[0].value[1]' 2>/dev/null || echo "0")
    print_info "Saga started: $saga_started"
    print_info "Saga completed: $saga_completed"
    
    # Проверяем бизнес-метрики
    print_info "Проверка бизнес-метрики..."
    business_count=$(curl -s "http://localhost:9090/api/v1/query?query=business_operations_total" | jq '.data.result | length' 2>/dev/null || echo "0")
    print_info "Бизнес-метрики: $business_count"
    
    print_success "Проверка метрик завершена"
}

# Основная функция
main() {
    print_header "QUICK TEST - ГЕНЕРАЦИЯ ВСЕХ МЕТРИК"
    print_info "Этот скрипт генерирует все метрики для заполнения дашбордов Grafana"
    
    # Генерируем все метрики
    generate_all_metrics
    
    # Ждем обработки метрик
    print_info "Ожидание обработки метрик..."
    sleep 10
    
    # Проверяем метрики
    check_metrics
    
    print_header "ЗАВЕРШЕНО"
    print_success "Quick test завершен! Все метрики сгенерированы."
    print_info "Grafana: http://localhost/"
    print_info "Prometheus: http://localhost:9090"
    print_info "Все дашборды должны быть заполнены данными!"
}

# Запуск
main "$@"
