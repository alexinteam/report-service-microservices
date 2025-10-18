#!/bin/bash

# 📊 Универсальный скрипт для мониторинга микросервисов
# Описание: Объединяет все функции мониторинга в одном скрипте

set -e

# Цвета для вывода
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Конфигурация
GRAFANA_URL="http://localhost"
GRAFANA_USER="admin"
GRAFANA_PASS="admin123"

# Функция для логирования
log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

info() {
    echo -e "${PURPLE}ℹ️  $1${NC}"
}

# Проверка зависимостей
check_dependencies() {
    log "Проверка зависимостей..."
    
    local missing_deps=()
    
    if ! command -v kubectl &> /dev/null; then
        missing_deps+=("kubectl")
    fi
    
    if ! command -v curl &> /dev/null; then
        missing_deps+=("curl")
    fi
    
    if ! command -v jq &> /dev/null; then
        warning "jq не установлен (опционально для JSON обработки)"
    fi
    
    if [ ${#missing_deps[@]} -eq 0 ]; then
        success "Все зависимости установлены"
        return 0
    else
        error "Отсутствуют зависимости: ${missing_deps[*]}"
        return 1
    fi
}

# Проверка кластера Kubernetes
check_cluster() {
    log "Проверка кластера Kubernetes..."
    
    if ! kubectl cluster-info > /dev/null 2>&1; then
        error "Не удается подключиться к кластеру Kubernetes"
        error "Убедитесь, что kubectl настроен и кластер доступен"
        return 1
    fi
    
    success "Кластер Kubernetes доступен"
    return 0
}

# Проверка статуса микросервисов
check_services() {
    log "Проверка статуса микросервисов..."
    
    local namespaces=("api-gateway" "user-service" "template-service" "report-service" "data-service" "notification-service" "storage-service")
    local all_ready=true
    
    for namespace in "${namespaces[@]}"; do
        local ready_pods=$(kubectl get pods -n "$namespace" --no-headers 2>/dev/null | grep -c "Running" || echo "0")
        local total_pods=$(kubectl get pods -n "$namespace" --no-headers 2>/dev/null | wc -l || echo "0")
        
        if [ "$total_pods" -gt 0 ]; then
            if [ "$ready_pods" -eq "$total_pods" ]; then
                success "$namespace готов ($ready_pods/$total_pods)"
            else
                warning "$namespace не готов ($ready_pods/$total_pods)"
                all_ready=false
            fi
        else
            warning "$namespace: поды не найдены"
            all_ready=false
        fi
    done
    
    if [ "$all_ready" = true ]; then
        success "Все сервисы готовы"
        return 0
    else
        warning "Некоторые сервисы не готовы"
        return 1
    fi
}

# Проверка системы мониторинга
check_monitoring() {
    log "Проверка системы мониторинга..."
    
    # Проверяем Prometheus
    local prometheus_ready=$(kubectl get pods -n system -l app=prometheus --no-headers 2>/dev/null | grep -c "Running" || echo "0")
    if [ "$prometheus_ready" -gt 0 ]; then
        success "Prometheus готов"
    else
        error "Prometheus не готов"
        return 1
    fi
    
    # Проверяем Grafana
    local grafana_ready=$(kubectl get pods -n system -l app=grafana --no-headers 2>/dev/null | grep -c "Running" || echo "0")
    if [ "$grafana_ready" -gt 0 ]; then
        success "Grafana готова"
    else
        error "Grafana не готова"
        return 1
    fi
    
    return 0
}

# Проверка Grafana
check_grafana() {
    log "Проверка Grafana..."
    
    if curl -s "http://localhost/api/health" | grep -q "ok"; then
        success "Grafana доступна по http://localhost/"
        info "Логин: admin, Пароль: admin123"
        return 0
    else
        error "Grafana недоступна"
        return 1
    fi
}

# Проверка Prometheus
check_prometheus() {
    log "Проверка Prometheus..."
    
    # Запускаем port-forward для Prometheus
    local pid=$(kubectl port-forward service/prometheus-service 9090:80 -n system &> /dev/null & echo $!)
    sleep 3
    
    if curl -s "http://localhost:9090/api/v1/query?query=up" | grep -q "success"; then
        success "Prometheus доступен"
        
        # Проверяем метрики микросервисов
        local metrics_count=$(curl -s "http://localhost:9090/api/v1/query?query=http_requests_total" | jq -r '.data.result | length' 2>/dev/null || echo "0")
        if [ "$metrics_count" -gt 0 ]; then
            success "Метрики микросервисов найдены ($metrics_count метрик)"
            
            # Показываем статистику по сервисам
            echo "   Статистика по сервисам:"
            curl -s "http://localhost:9090/api/v1/query?query=http_requests_total" | \
            jq -r '.data.result[] | "   \(.metric.service): \(.value[1]) запросов"' 2>/dev/null | \
            sort | uniq -c | sort -nr | head -10 || echo "   Не удалось получить детальную статистику"
        else
            warning "Метрики микросервисов не найдены"
        fi
    else
        error "Prometheus недоступен"
    fi
    
    # Закрываем port-forward
    kill $pid 2>/dev/null || true
}

# Проверка дашбордов
check_dashboards() {
    log "Проверка дашбордов..."
    
    # Проверяем доступность дашбордов через API
    local dashboards=$(curl -s "http://localhost/api/search?type=dash-db" -H "Authorization: Basic $(echo -n 'admin:admin123' | base64)" 2>/dev/null || echo "[]")
    
    local our_dashboards=("Microservices Overview" "Business Metrics" "Saga Metrics" "Service Details" "Alerts & Health")
    local found_count=0
    
    for dashboard_name in "${our_dashboards[@]}"; do
        if echo "$dashboards" | grep -q "$dashboard_name"; then
            ((found_count++))
        fi
    done
    
    if [ "$found_count" -gt 0 ]; then
        success "Готовые дашборды найдены ($found_count/5)"
        echo "   Доступные дашборды:"
        for dashboard_name in "${our_dashboards[@]}"; do
            if echo "$dashboards" | grep -q "$dashboard_name"; then
                echo "   • $dashboard_name"
            fi
        done
    else
        warning "Готовые дашборды не найдены"
    fi
}

# Создание дашборда
create_dashboard() {
    local dashboard_file="$1"
    local dashboard_name="$2"
    
    log "Создание дашборда '$dashboard_name'..."
    
    # Проверяем, существует ли дашборд
    local existing_id=$(curl -s "http://localhost/api/search?type=dash-db&query=$dashboard_name" \
        -H "Authorization: Basic $(echo -n 'admin:admin123' | base64)" \
        | jq -r '.[0].id // empty' 2>/dev/null)
    
    if [ -n "$existing_id" ]; then
        warning "Дашборд '$dashboard_name' уже существует (ID: $existing_id)"
        return 0
    fi
    
    # Создаем дашборд
    local response=$(curl -s -X POST "http://localhost/api/dashboards/db" \
        -H "Content-Type: application/json" \
        -H "Authorization: Basic $(echo -n 'admin:admin123' | base64)" \
        -d @"$dashboard_file")
    
    if echo "$response" | grep -q '"status":"success"'; then
        success "Дашборд '$dashboard_name' создан"
        return 0
    else
        error "Ошибка создания дашборда '$dashboard_name'"
        echo "$response" | jq -r '.message // .' 2>/dev/null || echo "$response"
        return 1
    fi
}

# Создание всех дашбордов
create_all_dashboards() {
    log "Создание всех дашбордов..."
    
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
            if create_dashboard "$file" "$name"; then
                ((success_count++))
            fi
        else
            error "Файл дашборда '$file' не найден"
        fi
        echo ""
    done
    
    log "Результат: $success_count/$total_count дашбордов создано"
}

# Генерация тестовых данных
generate_test_data() {
    log "Генерация тестовых данных..."
    
    local services=(
        "user-service-service:8081:user-service"
        "template-service-service:8082:template-service"
        "notification-service-service:8085:notification-service"
    )
    
    for service_info in "${services[@]}"; do
        local service_name=$(echo "$service_info" | cut -d: -f1)
        local port=$(echo "$service_info" | cut -d: -f2)
        local namespace=$(echo "$service_info" | cut -d: -f3)
        
        log "Генерация запросов для $service_name..."
        
        # Запускаем port-forward
        local pid=$(kubectl port-forward "service/$service_name" "$port:$port" -n "$namespace" &> /dev/null & echo $!)
        sleep 2
        
        # Генерируем тестовые запросы
        for i in {1..20}; do
            curl -s "http://localhost:$port/health" &> /dev/null || true
            curl -s "http://localhost:$port/metrics" &> /dev/null || true
            sleep 0.1
        done
        
        success "Тестовые запросы для $service_name отправлены"
        
        # Закрываем port-forward
        kill $pid 2>/dev/null || true
    done
}

# Показать статус системы мониторинга
show_status() {
    echo ""
    echo -e "${PURPLE}📊 Статус системы мониторинга${NC}"
    echo "================================"
    
    # Проверяем доступность Grafana
    if curl -s "http://localhost/api/health" | grep -q "ok"; then
        success "Grafana доступна по http://localhost/"
        info "Логин: admin, Пароль: admin123"
    else
        warning "Grafana недоступна по http://localhost/"
        info "Попробуйте: kubectl port-forward service/grafana-service 3000:80 -n system"
    fi
    
    # Проверяем доступность Prometheus
    local prometheus_pid=$(kubectl port-forward service/prometheus-service 9090:80 -n system &> /dev/null & echo $!)
    sleep 2
    
    if curl -s "http://localhost:9090/api/health" | grep -q "ok"; then
        success "Prometheus доступен"
        info "Для доступа: kubectl port-forward service/prometheus-service 9090:80 -n system"
    else
        warning "Prometheus недоступен"
    fi
    
    kill $prometheus_pid 2>/dev/null || true
}

# Показать полезные команды
show_commands() {
    echo -e "${PURPLE}🔧 Полезные команды${NC}"
    echo "=================="
    echo ""
    echo -e "${GREEN}Доступ к системам:${NC}"
    echo "• Grafana: http://localhost/"
    echo "• Prometheus: kubectl port-forward service/prometheus-service 9090:80 -n system"
    echo ""
    echo -e "${GREEN}Проверка сервисов:${NC}"
    echo "• Статус всех подов: kubectl get pods -A"
    echo "• Статус микросервисов: kubectl get pods -n microservices"
    echo "• Логи сервиса: kubectl logs -f deployment/user-service -n user-service"
    echo ""
    echo -e "${GREEN}Проверка метрик:${NC}"
    echo "• Метрики сервиса: kubectl port-forward service/user-service 8081:8081 -n user-service && curl http://localhost:8081/metrics"
    echo "• Prometheus targets: kubectl port-forward service/prometheus-service 9090:80 -n system && curl http://localhost:9090/targets"
    echo ""
    echo -e "${GREEN}Тестирование:${NC}"
    echo "• Быстрое тестирование: ./monitoring.sh test"
    echo "• Расширенное тестирование: ./monitoring.sh check"
    echo "• Проверка статуса: ./monitoring.sh status"
    echo ""
    echo -e "${GREEN}Дашборды:${NC}"
    echo "• Создать дашборды: ./monitoring.sh dashboards"
    echo "• Проверить дашборды: ./monitoring.sh check"
}

# Основная функция
main() {
    echo -e "${BLUE}"
    echo "📊 Универсальный скрипт мониторинга микросервисов"
    echo "================================================="
    echo -e "${NC}"
    
    check_dependencies || exit 1
    check_cluster || exit 1
    check_services || exit 1
    check_monitoring || exit 1
    
    echo ""
    check_grafana
    echo ""
    check_prometheus
    echo ""
    check_dashboards
    
    echo ""
    log "Генерация тестовых данных для демонстрации..."
    generate_test_data
    
    echo ""
    success "Проверка завершена!"
    echo ""
    show_status
    echo ""
    show_commands
}

# Обработка аргументов командной строки
case "${1:-}" in
    "check")
        check_dependencies && check_cluster && check_services && check_monitoring && check_grafana && check_prometheus && check_dashboards
        ;;
    "grafana")
        check_grafana
        ;;
    "prometheus")
        check_prometheus
        ;;
    "dashboards")
        check_grafana && create_all_dashboards
        ;;
    "test")
        generate_test_data
        ;;
    "status")
        show_status
        ;;
    "commands")
        show_commands
        ;;
    "help"|"-h"|"--help")
        echo "Использование: $0 [команда]"
        echo ""
        echo "Команды:"
        echo "  check      - Проверить все компоненты системы"
        echo "  grafana    - Проверить Grafana"
        echo "  prometheus - Проверить Prometheus и метрики"
        echo "  dashboards - Создать дашборды"
        echo "  test       - Генерировать тестовые данные"
        echo "  status     - Показать статус системы"
        echo "  commands   - Показать полезные команды"
        echo "  help       - Показать эту справку"
        echo ""
        echo "Без аргументов выполняется полная проверка"
        ;;
    *)
        main
        ;;
esac
