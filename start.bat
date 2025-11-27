@echo off
REM ============================================================================
REM Gmail Scanner - Script de Inicialização (Windows)
REM ============================================================================
REM
REM Este script facilita a inicialização da aplicação Gmail Scanner no Docker
REM
REM Uso:
REM   start.bat                 # Iniciar com configuração padrão
REM   start.bat --help          # Mostrar ajuda
REM   start.bat --build         # Fazer rebuild das imagens
REM   start.bat --logs          # Mostrar logs em tempo real
REM   start.bat --stop          # Parar todos os containers
REM   start.bat --clean         # Parar e remover containers/volumes
REM
REM ============================================================================

setlocal enabledelayedexpansion

REM Configurações
set SCRIPT_DIR=%~dp0
set ENV_FILE=%SCRIPT_DIR%.env

REM Cores (Windows 10+)
for /F %%A in ('echo prompt $H ^| cmd') do set "BS=%%A"

if not exist "%ENV_FILE%" (
    echo [!] Arquivo .env não encontrado
    echo.
    echo Criando .env a partir de .env.docker...
    copy "%SCRIPT_DIR%.env.docker" "%SCRIPT_DIR%.env" >nul
    echo [+] Arquivo .env criado com valores de exemplo
    echo [!] EDITE .env e preencha: GMAIL_CLIENT_ID e GMAIL_CLIENT_SECRET
    echo.
    pause
    exit /b 1
)

REM Verificar se as variáveis estão preenchidas
findstr /r "seu_client_id_aqui\|seu_client_secret_aqui" "%ENV_FILE%" >nul
if !errorlevel! equ 0 (
    echo [X] Variáveis obrigatórias não configuradas em .env
    echo [!] Edite .env e preencha GMAIL_CLIENT_ID e GMAIL_CLIENT_SECRET
    pause
    exit /b 1
)

REM Verificar Docker
docker --version >nul 2>&1
if !errorlevel! neq 0 (
    echo [X] Docker não está instalado ou não está no PATH
    pause
    exit /b 1
)

REM Processar argumentos
if "%1"=="" goto :start_services
if /i "%1"=="--help" goto :show_help
if /i "%1"=="-h" goto :show_help
if /i "%1"=="--build" goto :build_images
if /i "%1"=="--up" goto :start_services
if /i "%1"=="--down" goto :down_services
if /i "%1"=="--stop" goto :down_services
if /i "%1"=="--logs" goto :show_logs
if /i "%1"=="--clean" goto :clean_services
if /i "%1"=="--ps" goto :show_ps
if /i "%1"=="--health" goto :check_health
if /i "%1"=="--shell" goto :open_shell

:show_help
cls
echo ================================
echo   Gmail Scanner - Docker
echo ================================
echo.
echo Opcoes disponiveis:
echo.
echo   start.bat [COMANDO]
echo.
echo Comandos:
echo   (nenhum)      Iniciar todos os servicos em background
echo   --build       Fazer rebuild das imagens Docker
echo   --up          Iniciar servicos (mesmo que sem opcao)
echo   --down        Parar todos os servicos
echo   --stop        Alias para --down
echo   --logs        Mostrar logs em tempo real
echo   --clean       Parar containers e remover volumes
echo   --ps          Ver status dos containers
echo   --health      Verificar saude dos servicos
echo   --shell       Entrar em shell interativo do backend
echo   --help        Mostrar esta ajuda
echo.
echo Exemplos:
echo   start.bat                # Iniciar
echo   start.bat --build        # Rebuild e iniciar
echo   start.bat --logs         # Ver logs
echo   start.bat --clean        # Parar e limpar
echo.
pause
exit /b 0

:start_services
cls
echo ================================
echo   Gmail Scanner - Docker
echo ================================
echo.
echo [+] Iniciando servicos...
echo.

echo [+] Iniciando containers...
docker-compose up -d

if !errorlevel! neq 0 (
    echo [X] Erro ao iniciar containers
    pause
    exit /b 1
)

echo.
echo [+] Servicos iniciados!
echo.
echo [!] Aguardando inicializacao dos servicos (40 segundos)...

REM Aguardar 40 segundos
timeout /t 40 /nobreak

cls
echo ================================
echo   Gmail Scanner - Pronto!
echo ================================
echo.
echo [+] Aplicacao pronta!
echo.
echo Acesse: http://localhost:8080
echo.
echo Servicos:
echo   Frontend:       http://localhost:8080
echo   Backend API:    http://localhost:8080/api
echo   OpenSearch:     http://localhost:9200
echo   LibreTranslate: http://localhost:5000
echo.
echo Para ver logs:  start.bat --logs
echo Para parar:     start.bat --down
echo.
pause
exit /b 0

:build_images
echo [+] Fazendo rebuild das imagens...
docker-compose build --no-cache
if !errorlevel! equ 0 (
    echo [+] Build concluido
    call :start_services
) else (
    echo [X] Erro no build
    pause
    exit /b 1
)
exit /b 0

:down_services
echo [+] Parando servicos...
docker-compose down
echo [+] Servicos parados
pause
exit /b 0

:clean_services
echo [!] Removendo containers, volumes e dados...
docker-compose down -v
echo [+] Limpeza concluida
pause
exit /b 0

:show_logs
docker-compose logs -f
exit /b 0

:show_ps
cls
docker-compose ps
pause
exit /b 0

:check_health
cls
echo ================================
echo   Gmail Scanner - Health Check
echo ================================
echo.
echo Verificando saude dos servicos...
echo.

REM Simples verificação sem parsing de cores
echo Acesse para verificar:
echo   OpenSearch:     http://localhost:9200/_cluster/health
echo   LibreTranslate: http://localhost:5000/languages
echo   Backend:        http://localhost:8080/api/health
echo.

pause
exit /b 0

:open_shell
echo [+] Abrindo shell no backend...
docker-compose exec gmail-scanner-backend cmd
exit /b 0
