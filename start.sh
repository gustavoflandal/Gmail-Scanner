#!/bin/bash

################################################################################
# Gmail Scanner - Script de Inicialização
################################################################################
#
# Este script facilita a inicialização da aplicação Gmail Scanner no Docker
#
# Uso:
#   ./start.sh                    # Iniciar com configuração padrão
#   ./start.sh --help             # Mostrar ajuda
#   ./start.sh --build            # Fazer rebuild das imagens
#   ./start.sh --logs             # Mostrar logs em tempo real
#   ./start.sh --stop             # Parar todos os containers
#   ./start.sh --clean            # Parar e remover containers/volumes
#
################################################################################

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Variáveis
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/.env"
DOCKER_COMPOSE="docker-compose"

# Funções
print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  Gmail Scanner - Docker${NC}"
    echo -e "${BLUE}================================${NC}"
    echo ""
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker não está instalado ou não está no PATH"
        exit 1
    fi
    print_success "Docker encontrado"

    if ! command -v docker-compose &> /dev/null; then
        print_warning "docker-compose não encontrado, usando docker compose"
        DOCKER_COMPOSE="docker compose"
    else
        print_success "Docker Compose encontrado"
    fi
}

check_env() {
    if [ ! -f "${ENV_FILE}" ]; then
        print_error ".env não encontrado"
        echo ""
        echo "Criando .env a partir de .env.docker..."
        cp "${SCRIPT_DIR}/.env.docker" "${ENV_FILE}"
        print_warning "Arquivo .env criado com valores de exemplo"
        print_warning "EDITE .env e preencha: GMAIL_CLIENT_ID e GMAIL_CLIENT_SECRET"
        echo ""
        echo "Exemplo:"
        echo "  GMAIL_CLIENT_ID=xxx.apps.googleusercontent.com"
        echo "  GMAIL_CLIENT_SECRET=yyy"
        echo ""
        return 1
    fi

    # Verificar variáveis obrigatórias
    if grep -q "seu_client_id_aqui" "${ENV_FILE}" || grep -q "seu_client_secret_aqui" "${ENV_FILE}"; then
        print_error "Variáveis obrigatórias não configuradas em .env"
        print_warning "Edite .env e preencha GMAIL_CLIENT_ID e GMAIL_CLIENT_SECRET"
        return 1
    fi

    print_success ".env encontrado e validado"
}

show_help() {
    print_header
    echo "Opções disponíveis:"
    echo ""
    echo "  ./start.sh [COMANDO]"
    echo ""
    echo "Comandos:"
    echo "  (nenhum)      Iniciar todos os serviços em background"
    echo "  --build       Fazer rebuild das imagens Docker"
    echo "  --up          Iniciar serviços (mesmo que sem opção)"
    echo "  --down        Parar todos os serviços"
    echo "  --stop        Alias para --down"
    echo "  --logs        Mostrar logs em tempo real"
    echo "  --clean       Parar containers e remover volumes"
    echo "  --ps          Ver status dos containers"
    echo "  --health      Verificar saúde dos serviços"
    echo "  --shell       Entrar em shell interativo do backend"
    echo "  --help        Mostrar esta ajuda"
    echo ""
    echo "Exemplos:"
    echo "  ./start.sh                # Iniciar"
    echo "  ./start.sh --build        # Rebuild e iniciar"
    echo "  ./start.sh --logs         # Ver logs"
    echo "  ./start.sh --clean        # Parar e limpar"
    echo ""
}

start_services() {
    print_header
    print_info "Iniciando serviços..."
    echo ""

    check_docker
    if ! check_env; then
        exit 1
    fi
    echo ""

    print_info "Iniciando containers..."
    ${DOCKER_COMPOSE} up -d

    echo ""
    print_success "Serviços iniciados!"
    echo ""

    # Aguardar um pouco para os serviços ficarem prontos
    print_info "Aguardando inicialização dos serviços (40 segundos)..."
    sleep 40

    print_info "Verificando saúde dos serviços..."
    check_health

    echo ""
    print_success "Aplicação pronta!"
    echo ""
    echo "Acesse: ${BLUE}http://localhost:8080${NC}"
    echo ""
    echo "Serviços:"
    echo "  Frontend:       http://localhost:8080"
    echo "  Backend API:    http://localhost:8080/api"
    echo "  OpenSearch:     http://localhost:9200"
    echo "  LibreTranslate: http://localhost:5000"
    echo ""
    echo "Para ver logs: ${BLUE}./start.sh --logs${NC}"
    echo "Para parar:    ${BLUE}./start.sh --down${NC}"
    echo ""
}

build_images() {
    print_header
    print_info "Fazendo rebuild das imagens..."
    ${DOCKER_COMPOSE} build --no-cache
    print_success "Build concluído"
    echo ""
}

down_services() {
    print_info "Parando serviços..."
    ${DOCKER_COMPOSE} down
    print_success "Serviços parados"
}

clean_services() {
    print_warning "Removendo containers, volumes e dados..."
    ${DOCKER_COMPOSE} down -v
    print_success "Limpeza concluída"
}

show_logs() {
    print_info "Mostrando logs (Ctrl+C para sair)..."
    ${DOCKER_COMPOSE} logs -f
}

show_ps() {
    print_header
    ${DOCKER_COMPOSE} ps
    echo ""
}

check_health() {
    echo "  OpenSearch:     $(check_service "opensearch" "9200" "/_cluster/health")"
    echo "  LibreTranslate: $(check_service "libretranslate" "5000" "/languages")"
    echo "  Backend:        $(check_service "backend" "8080" "/api/health")"
}

check_service() {
    local service=$1
    local port=$2
    local path=$3

    if curl -s "http://localhost:${port}${path}" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Online${NC}"
    else
        echo -e "${YELLOW}⏳ Inicializando...${NC}"
    fi
}

open_shell() {
    print_info "Abrindo shell no backend..."
    ${DOCKER_COMPOSE} exec gmail-scanner-backend sh
}

# Main
case "${1:-up}" in
    --help|-h)
        show_help
        ;;
    --build|-b)
        build_images
        start_services
        ;;
    --up)
        start_services
        ;;
    --down|--stop|stop|down)
        down_services
        ;;
    --logs|-l|logs)
        show_logs
        ;;
    --clean)
        clean_services
        ;;
    --ps)
        show_ps
        ;;
    --health)
        print_header
        print_info "Verificando saúde dos serviços..."
        echo ""
        check_health
        echo ""
        ;;
    --shell)
        open_shell
        ;;
    *)
        start_services
        ;;
esac
