#!/bin/bash

# Скрипт развертывания микросервисов в Kubernetes

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

check_cluster() {
    print_status "Проверяем подключение к кластеру Kubernetes..."
    
    if ! kubectl cluster-info > /dev/null 2>&1; then
        print_error "Не удается подключиться к кластеру Kubernetes"
        print_error "Убедитесь, что kubectl настроен и кластер доступен"
        exit 1
    fi
    
    print_success "Подключение к кластеру установлено"
}

# Развертывание в Kubernetes
deploy_k8s() {
    print_status "Развертываем в Kubernetes..."
    
    # Создаем namespace'ы
    print_status "Создаем namespace'ы..."
kubectl apply -f ../k8s/01-system-namespace.yaml
kubectl apply -f ../k8s/02-namespaces.yaml

    # Развертываем системные сервисы
    print_status "Развертываем системные сервисы..."
kubectl apply -f ../k8s/04-redis.yaml
kubectl apply -f ../k8s/05-rabbitmq.yaml

    # Развертываем мониторинг
    print_status "Развертываем систему мониторинга..."
    kubectl apply -f ../k8s/14-prometheus.yaml
    kubectl apply -f ../k8s/15-grafana.yaml
    
    # Ждем готовности системных сервисов
    print_status "Ждем готовности системных сервисов..."
    kubectl wait --for=condition=ready pod -l app=redis --namespace=system --timeout=60s
    kubectl wait --for=condition=ready pod -l app=rabbitmq --namespace=system --timeout=60s
    
    # Ждем готовности мониторинга
    print_status "Ждем готовности системы мониторинга..."
    kubectl wait --for=condition=ready pod -l app=prometheus --namespace=system --timeout=120s
    kubectl wait --for=condition=ready pod -l app=grafana --namespace=system --timeout=120s
    
    # Развертываем микросервисы
    print_status "Развертываем микросервисы..."
    
    services=(
        "06-api-gateway.yaml"
        "07-user-service.yaml"
        "08-template-service.yaml"
        "09-report-service.yaml"
        "10-data-service.yaml"
        "11-notification-service.yaml"
        "13-storage-service.yaml"
    )
    
    for service in "${services[@]}"; do
        print_status "Развертываем $service..."
        kubectl apply -f "../k8s/$service"
        if [ $? -eq 0 ]; then
            print_success "$service успешно развернут"
        else
            print_error "Ошибка при развертывании $service"
            exit 1
        fi
    done
    
    # Ждем готовности всех подов
    print_status "Ждем готовности всех подов..."
    
    namespaces=("api-gateway" "user-service" "template-service" "report-service" "data-service" "notification-service" "storage-service")
    
    for namespace in "${namespaces[@]}"; do
        print_status "Проверяем namespace: $namespace"
        kubectl wait --for=condition=ready pod --all --namespace="$namespace" --timeout=120s
    done
    
    print_success "Все сервисы развернуты в Kubernetes"
}

# Настройка мониторинга и дашбордов
setup_monitoring() {
    print_status "Настраиваем мониторинг и дашборды..."
    
    # Ждем, пока Grafana полностью запустится
    print_status "Ждем готовности Grafana..."
    sleep 30
    
    # Проверяем доступность Grafana
    local grafana_ready=false
    for i in {1..10}; do
        if curl -s "http://localhost/api/health" | grep -q "ok"; then
            grafana_ready=true
            break
        fi
        print_status "Ожидание готовности Grafana... ($i/10)"
        sleep 10
    done
    
    if [ "$grafana_ready" = true ]; then
        print_success "Grafana готова к работе"
        
        # Создаем дашборды и генерируем тестовые данные
        print_status "Настраиваем мониторинг..."
        if [ -f "monitoring.sh" ]; then
            ./monitoring.sh dashboards
            ./monitoring.sh test
            print_success "Мониторинг настроен"
        else
            print_warning "Скрипт мониторинга не найден"
        fi
    else
        print_warning "Grafana не готова, дашборды будут созданы позже"
        print_status "Для создания дашбордов выполните: cd scripts && ./create-dashboards.sh"
    fi
}

show_status() {
    print_status "Статус развертывания:"
    echo "========================"
    
    namespaces=("system" "api-gateway" "user-service" "template-service" "report-service" "data-service" "notification-service" "storage-service")
    
    for namespace in "${namespaces[@]}"; do
        echo ""
        echo "Namespace: $namespace"
        kubectl get pods --namespace="$namespace" 2>/dev/null || echo "Namespace не найден"
    done
    
    echo ""
    print_status "Сервисы:"
    echo "=========="
    kubectl get services --all-namespaces | grep -E "(api-gateway|user-service|template-service|report-service|data-service|notification-service|storage-service|prometheus|grafana)" || true
}

show_access_info() {
    echo ""
    print_success "🎉 Развертывание завершено!"
    echo "=========================="
    echo ""
    print_status "🌐 Доступ к сервисам:"
    echo "===================="
    echo "API Gateway: kubectl port-forward service/api-gateway-service 8080:80 --namespace=api-gateway"
    echo ""
    print_status "📊 Мониторинг:"
    echo "=============="
    echo "Prometheus: kubectl port-forward service/prometheus-service 9090:80 --namespace=system"
    echo "Grafana: http://localhost/"
    echo "  Логин: admin"
    echo "  Пароль: admin123"
    echo ""
    print_status "📈 Дашборды Grafana:"
    echo "==================="
    echo "- Microservices Overview: Общий обзор всех микросервисов"
    echo "- Business Metrics: Бизнес-метрики и операции с БД"
    echo "- Saga Metrics: Метрики Saga паттерна"
    echo "- Service Details: Детальный мониторинг каждого сервиса"
    echo "- Alerts & Health: Алерты и состояние здоровья сервисов"
    echo ""
    print_status "🔍 Просмотр метрик:"
    echo "=================="
    echo "Каждый сервис предоставляет метрики на /metrics endpoint"
    echo "Пример: curl http://localhost:8080/metrics (через port-forward)"
    echo ""
    print_status "📝 Postman коллекция:"
    echo "===================="
    echo "Файл: postman/Report_Microservices_Complete.postman_collection.json"
    echo "Домен: arch.homework"
    echo "Порт API Gateway: 8080"
    echo ""
    print_status "🔧 Полезные скрипты:"
    echo "====================="
    echo "cd scripts"
    echo "./monitoring.sh             # Универсальный скрипт мониторинга"
    echo "./deploy.sh --status        # Показать статус развертывания"
    echo "./undeploy.sh               # Удаление системы"
    echo ""
    print_status "Или через LoadBalancer (если поддерживается):"
    echo "kubectl get service --all-namespaces | grep LoadBalancer"
}

main() {
    echo "🚀 Скрипт развертывания микросервисов в Kubernetes"
    echo "=================================================="
    echo ""
    
    # Парсинг аргументов
    SHOW_STATUS=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --status)
                SHOW_STATUS=true
                shift
                ;;
            --help)
                echo "Использование: $0 [опции]"
                echo ""
                echo "Опции:"
                echo "  --status    Показать статус развертывания"
                echo "  --help      Показать эту справку"
                echo ""
                echo "Примеры:"
                echo "  $0                      # Развертывание в Kubernetes"
                echo "  $0 --status            # Показать статус"
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
        show_status
    else
        check_cluster
        deploy_k8s
        setup_monitoring
        show_access_info
    fi
}

main "$@"