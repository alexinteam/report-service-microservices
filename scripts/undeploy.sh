#!/bin/bash

# Скрипт удаления микросервисов из Kubernetes

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Удаление из Kubernetes
undeploy_k8s() {
    print_status "Удаляем микросервисы из Kubernetes..."
    
    # Удаляем микросервисы
    print_status "Удаляем микросервисы..."
    
    services=(
        "13-storage-service.yaml"
        "11-notification-service.yaml"
        "10-data-service.yaml"
        "09-report-service.yaml"
        "08-template-service.yaml"
        "07-user-service.yaml"
        "06-api-gateway.yaml"
    )
    
    for service in "${services[@]}"; do
        print_status "Удаляем $service..."
        kubectl delete -f "../k8s/$service" --ignore-not-found=true
        if [ $? -eq 0 ]; then
            print_success "$service успешно удален"
        else
            print_warning "Предупреждение при удалении $service"
        fi
    done
    
    # Удаляем мониторинг
    print_status "Удаляем систему мониторинга..."
    kubectl delete -f ../k8s/15-grafana.yaml --ignore-not-found=true
    kubectl delete -f ../k8s/14-prometheus.yaml --ignore-not-found=true
    
    # Удаляем системные сервисы
    print_status "Удаляем системные сервисы..."
    kubectl delete -f ../k8s/05-rabbitmq.yaml --ignore-not-found=true
    kubectl delete -f ../k8s/04-redis.yaml --ignore-not-found=true
    
    # Удаляем namespace'ы
    print_status "Удаляем namespace'ы..."
    kubectl delete -f ../k8s/02-namespaces.yaml --ignore-not-found=true
    kubectl delete -f ../k8s/01-system-namespace.yaml --ignore-not-found=true
    
    print_success "Все сервисы удалены из Kubernetes"
}

# Очистка дашбордов Grafana
cleanup_grafana() {
    print_status "Очищаем дашборды Grafana..."
    
    # Проверяем доступность Grafana
    local grafana_ready=false
    for i in {1..5}; do
        if curl -s "http://localhost/api/health" | grep -q "ok"; then
            grafana_ready=true
            break
        fi
        print_status "Ожидание доступности Grafana... ($i/5)"
        sleep 2
    done
    
    if [ "$grafana_ready" = true ]; then
        print_status "Удаляем дашборды..."
        
        local dashboards=$(curl -s "http://localhost/api/search?type=dash-db" \
            -H "Authorization: Basic $(echo -n 'admin:admin123' | base64)" 2>/dev/null || echo "[]")
        
        local our_dashboards=("Microservices Overview" "Business Metrics" "Saga Metrics" "Service Details" "Alerts & Health")
        
        for dashboard_name in "${our_dashboards[@]}"; do
            local dashboard_id=$(echo "$dashboards" | jq -r ".[] | select(.title == \"$dashboard_name\") | .id" 2>/dev/null | head -1)
            if [ -n "$dashboard_id" ] && [ "$dashboard_id" != "null" ] && [ "$dashboard_id" != "" ]; then
                print_status "Удаляем дашборд: $dashboard_name (ID: $dashboard_id)"
                curl -s -X DELETE "http://localhost/api/dashboards/db/$dashboard_id" \
                    -H "Authorization: Basic $(echo -n 'admin:admin123' | base64)" >/dev/null
                print_success "Дашборд $dashboard_name удален"
            fi
        done
        
        print_success "Дашборды Grafana очищены"
    else
        print_warning "Grafana недоступна, дашборды не удалены"
    fi
}

show_final_status() {
    print_status "Финальный статус:"
    echo "=================="
    
    namespaces=("system" "api-gateway" "user-service" "template-service" "report-service" "data-service" "notification-service" "storage-service")
    
    for namespace in "${namespaces[@]}"; do
        if kubectl get namespace "$namespace" >/dev/null 2>&1; then
            print_warning "Namespace $namespace все еще существует"
        else
            print_success "Namespace $namespace удален"
        fi
    done
}

# Основная функция
main() {
    echo "🗑️ Скрипт удаления микросервисов из Kubernetes"
    echo "=============================================="
    echo ""
    
    SHOW_STATUS=false
    CLEANUP_DASHBOARDS=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --status)
                SHOW_STATUS=true
                shift
                ;;
            --dashboards)
                CLEANUP_DASHBOARDS=true
                shift
                ;;
            --help)
                echo "Использование: $0 [опции]"
                echo ""
                echo "Опции:"
                echo "  --status      Показать статус удаления"
                echo "  --dashboards  Удалить только дашборды Grafana"
                echo "  --help        Показать эту справку"
                echo ""
                echo "Примеры:"
                echo "  $0                      # Полное удаление из Kubernetes"
                echo "  $0 --status            # Показать статус"
                echo "  $0 --dashboards        # Удалить только дашборды"
                exit 0
                ;;
            *)
                print_error "Неизвестная опция: $1"
                echo "Используйте --help для справки"
                exit 1
                ;;
        esac
    done
    
    if [ "$SHOW_STATUS" = true ]; then
        show_final_status
    elif [ "$CLEANUP_DASHBOARDS" = true ]; then
        cleanup_grafana
    else
        cleanup_grafana
        undeploy_k8s
        show_final_status
    fi
    
    echo ""
    print_success "🎉 Удаление завершено!"
    echo "======================"
}

main "$@"