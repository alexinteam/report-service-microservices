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

# Функция для создания дашбордов
create_dashboards() {
    print_header "СОЗДАНИЕ ДАШБОРДОВ"
    
    # Проверяем доступность Grafana
    if ! curl -s "http://localhost/api/health" | grep -q "ok"; then
        print_error "Grafana недоступна. Запустите: kubectl port-forward service/grafana-service 3000:80 -n system"
        return 1
    fi
    
    local dashboards=(
        "../dashboards/microservices-overview-dashboard.json:Microservices Overview"
        "../dashboards/business-metrics-dashboard.json:Business Metrics"
        "../dashboards/saga-metrics-dashboard.json:Saga Metrics"
        "../dashboards/service-details-dashboard.json:Service Details"
        "../dashboards/alerts-health-dashboard.json:Alerts & Health"
    )
    
    local success_count=0
    local total_count=${#dashboards[@]}
    
    for dashboard_info in "${dashboards[@]}"; do
        local file=$(echo "$dashboard_info" | cut -d: -f1)
        local name=$(echo "$dashboard_info" | cut -d: -f2)
        
        if [ -f "$file" ]; then
            print_info "Создание дашборда '$name'..."
            
            # Проверяем, существует ли дашборд
            local existing_id=$(curl -s "http://localhost/api/search?type=dash-db&query=$name" \
                -H "Authorization: Basic $(echo -n 'admin:admin123' | base64)" \
                | jq -r '.[0].id // empty' 2>/dev/null)
            
            if [ -n "$existing_id" ]; then
                print_warning "Дашборд '$name' уже существует (ID: $existing_id)"
                ((success_count++))
                continue
            fi
            
            # Создаем дашборд
            local response=$(curl -s -X POST "http://localhost/api/dashboards/db" \
                -H "Content-Type: application/json" \
                -H "Authorization: Basic $(echo -n 'admin:admin123' | base64)" \
                -d @"$file")
            
            if echo "$response" | grep -q '"status":"success"'; then
                print_success "Дашборд '$name' создан"
                ((success_count++))
            else
                print_error "Ошибка создания дашборда '$name'"
                echo "$response" | jq -r '.message // .' 2>/dev/null || echo "$response"
            fi
        else
            print_error "Файл дашборда '$file' не найден"
        fi
        echo ""
    done
    
    print_info "Результат: $success_count/$total_count дашбордов создано"
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
    
    # Генерируем метрики для CSV экспорта
    print_info "Генерация метрик CSV экспорта..."
    
    # Создаем несколько отчетов через саги для экспорта
    report_ids=()
    for i in {1..4}; do
        print_info "Создание отчета для CSV экспорта $i..."
        saga_response=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
            -d "{\"template_id\": \"1\",\"name\": \"CSV Export Report $i\",\"format\": \"csv\",\"parameters\": {\"title\": \"CSV Export Report $i\"}}" \
            http://arch.homework/api/v1/sagas/reports)
        
        saga_id=$(echo "$saga_response" | jq -r '.saga_id')
        if [ "$saga_id" != "null" ] && [ -n "$saga_id" ]; then
            print_success "Saga для CSV отчета $i создана: $saga_id"
            sleep 2  # Ждем завершения саги
            
            # Получаем ID созданного отчета
            reports_response=$(curl -s -H "Authorization: Bearer $TOKEN" http://arch.homework/api/v1/reports)
            latest_report_id=$(echo "$reports_response" | jq -r '.reports[0].id')
            if [ "$latest_report_id" != "null" ] && [ -n "$latest_report_id" ]; then
                report_ids+=("$latest_report_id")
                print_success "Отчет $i готов для экспорта: ID $latest_report_id"
            fi
        fi
        sleep 1
    done
    
    # Экспортируем отчеты в CSV
    for report_id in "${report_ids[@]}"; do
        if [ -n "$report_id" ]; then
            print_info "Экспорт отчета $report_id в CSV..."
            curl -s -H "Authorization: Bearer $TOKEN" \
                http://arch.homework/api/v1/reports/$report_id/export/csv >/dev/null
            sleep 0.5
        fi
    done
    
    print_success "Метрики CSV экспорта сгенерированы"
    
    # Генерируем метрики для уведомлений
    print_info "Генерация метрик уведомлений..."
    
    # Создаем несколько саг для генерации уведомлений
    for i in {1..5}; do
        print_info "Создание саги для уведомления $i..."
        curl -s -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
            -d "{\"template_id\": \"1\",\"name\": \"Notification Test $i\",\"format\": \"html\",\"parameters\": {\"title\": \"Notification Test $i\"}}" \
            http://arch.homework/api/v1/sagas/reports >/dev/null
        sleep 1
    done
    
    # Проверяем уведомления
    sleep 3
    notifications_response=$(curl -s -H "Authorization: Bearer $TOKEN" http://arch.homework/api/v1/notifications?page=1&limit=10)
    notification_count=$(echo "$notifications_response" | jq -r '.total // 0' 2>/dev/null || echo "0")
    print_success "Уведомления созданы: $notification_count"
    
    print_success "Метрики уведомлений сгенерированы"
    
    # Генерируем Saga метрики
    print_info "Генерация Saga метрик..."
    
    saga_ids=()
    for i in {1..8}; do
        print_info "Создание Saga $i..."
        response=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
            -d "{\"template_id\": \"1\",\"name\": \"Quick Test Saga $i\",\"format\": \"html\",\"parameters\": {\"title\": \"Quick Test Saga $i\"}}" \
            http://arch.homework/api/v1/sagas/reports)
        
        saga_id=$(echo "$response" | jq -r '.saga_id' 2>/dev/null || echo "")
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
    
    # Проверяем HTTP метрики
    print_info "Проверка HTTP метрик..."
    http_count=$(curl -s "http://localhost:9090/api/v1/query?query=http_requests_total" | jq '.data.result | length' 2>/dev/null || echo "0")
    print_info "HTTP метрики: $http_count"
    
    # Проверяем бизнес-метрики
    print_info "Проверка бизнес-метрик..."
    business_count=$(curl -s "http://localhost:9090/api/v1/query?query=business_operations_total" | jq '.data.result | length' 2>/dev/null || echo "0")
    print_info "Бизнес-метрики: $business_count"
    
    # Проверяем Saga метрики
    print_info "Проверка Saga метрик..."
    saga_started=$(curl -s "http://localhost:9090/api/v1/query?query=business_operations_total%7Boperation%3D%22saga_started%22%7D" | jq '.data.result[0].value[1]' 2>/dev/null || echo "0")
    saga_completed=$(curl -s "http://localhost:9090/api/v1/query?query=business_operations_total%7Boperation%3D%22saga_completed%22%7D" | jq '.data.result[0].value[1]' 2>/dev/null || echo "0")
    print_info "Saga started: $saga_started"
    print_info "Saga completed: $saga_completed"
    
    # Проверяем Database метрики
    print_info "Проверка Database метрики..."
    db_count=$(curl -s "http://localhost:9090/api/v1/query?query=database_query_duration_seconds_count" | jq '.data.result | length' 2>/dev/null || echo "0")
    print_info "Database метрики: $db_count"
    
    # Проверяем Memory метрики
    print_info "Проверка Memory метрики..."
    memory_count=$(curl -s "http://localhost:9090/api/v1/query?query=go_memory_classes_heap_objects_bytes" | jq '.data.result | length' 2>/dev/null || echo "0")
    print_info "Memory метрики: $memory_count"
    
    print_success "Проверка метрик завершена"
}

# Функция для проверки сервисов
check_services() {
    print_header "ПРОВЕРКА СЕРВИСОВ"
    
    # Проверяем API Gateway
    if curl -s http://arch.homework/api/v1/public/health | grep -q "healthy"; then
        print_success "API Gateway работает"
    else
        print_error "API Gateway недоступен"
    fi
    
    # Проверяем User Service
    if curl -s -H "Authorization: Bearer $(curl -s -X POST -H "Content-Type: application/json" -d '{"email":"test@example.com","password":"password123"}' http://arch.homework/api/v1/users/login | jq -r '.token')" http://arch.homework/api/v1/users/profile >/dev/null; then
        print_success "User Service работает"
    else
        print_error "User Service недоступен"
    fi
    
    # Проверяем Template Service
    if curl -s -H "Authorization: Bearer $(curl -s -X POST -H "Content-Type: application/json" -d '{"email":"test@example.com","password":"password123"}' http://arch.homework/api/v1/users/login | jq -r '.token')" http://arch.homework/api/v1/templates >/dev/null; then
        print_success "Template Service работает"
    else
        print_error "Template Service недоступен"
    fi
    
    # Проверяем Report Service
    if curl -s -H "Authorization: Bearer $(curl -s -X POST -H "Content-Type: application/json" -d '{"email":"test@example.com","password":"password123"}' http://arch.homework/api/v1/users/login | jq -r '.token')" http://arch.homework/api/v1/reports >/dev/null; then
        print_success "Report Service работает"
    else
        print_error "Report Service недоступен"
    fi
    
    # Проверяем Data Service
    if curl -s -H "Authorization: Bearer $(curl -s -X POST -H "Content-Type: application/json" -d '{"email":"test@example.com","password":"password123"}' http://arch.homework/api/v1/users/login | jq -r '.token')" http://arch.homework/api/v1/data-sources >/dev/null; then
        print_success "Data Service работает"
    else
        print_error "Data Service недоступен"
    fi
}

# Функция для проверки мониторинга
check_monitoring() {
    print_header "ПРОВЕРКА МОНИТОРИНГА"
    
    # Grafana
    if curl -s http://localhost/api/health | grep -q "ok"; then
        print_success "Grafana работает"
    else
        print_error "Grafana недоступна"
    fi
    
    # Prometheus
    if curl -s http://localhost:9090/-/healthy | grep -q "Healthy"; then
        print_success "Prometheus работает"
    else
        print_warning "Prometheus недоступен через Ingress (требует port-forward)"
    fi
}

# Основная функция
main() {
    print_header "QUICK TEST - ПОЛНАЯ ПРОВЕРКА СИСТЕМЫ"
    print_info "Этот скрипт проверяет все сервисы и генерирует метрики для Grafana"
    
    # Проверяем сервисы
    check_services
    
    # Проверяем мониторинг
    check_monitoring
    
    # Создаем дашборды
    create_dashboards
    
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

# Обработка аргументов командной строки
case "${1:-}" in
    "dashboards")
        create_dashboards
        ;;
    "metrics")
        generate_all_metrics
        ;;
    "check")
        check_services && check_monitoring && check_metrics
        ;;
    "help"|"-h"|"--help")
        echo "Использование: $0 [команда]"
        echo ""
        echo "Команды:"
        echo "  dashboards - Создать дашборды в Grafana"
        echo "  metrics    - Генерировать тестовые метрики"
        echo "  check      - Проверить сервисы и метрики"
        echo "  help       - Показать эту справку"
        echo ""
        echo "Без аргументов выполняется полная проверка с созданием дашбордов и генерацией метрик"
        ;;
    *)
        main
        ;;
esac