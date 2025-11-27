# Gmail Scanner - IMAP Edition

🚀 **Varredura inteligente de Gmail com autenticação IMAP simplificada**

Uma aplicação completa para escanear e indexar emails do Gmail usando protocolo IMAP, sem necessidade de configuração OAuth complexa. Interface web moderna com React, backend em Go e armazenamento em SQLite.

**⚡ Setup em 5 minutos!** Veja [SETUP_IMAP.md](SETUP_IMAP.md)

---

## 🌟 Recursos Principais

- ✅ **Autenticação IMAP Simplificada** - Sem OAuth, apenas email + senha de app
- ✅ **Seleção de Pastas** - Escolha quais pastas escanear (INBOX, Sent, etc)
- ✅ **Progresso em Tempo Real** - Barra de progresso e status detalhado
- ✅ **Cancelamento de Varredura** - Interrompa a varredura a qualquer momento
- ✅ **Grid de Mensagens** - Visualize e filtre emails com interface moderna
- ✅ **Busca Avançada** - Filtre por pasta, assunto, remetente
- ✅ **Leitura Completa** - Importa TODOS os emails da pasta (não apenas os últimos 100)
- ✅ **Interface Responsiva** - Dashboard React com Tailwind CSS
- ✅ **Docker Ready** - Deploy em containers
- ✅ **100% Open Source** - MIT License

---

## 📋 Requisitos

- **Go 1.23+** (para desenvolvimento local)
- **Node.js 20+** (para frontend)
- **Docker & Docker Compose** (para produção)
- **Conta Gmail** com IMAP habilitado
- **Senha de App do Google** (2FA necessário)

---

## 🚀 Início Rápido

### Opção 1: Docker (Recomendado)

```bash
# 1. Clonar repositório
git clone https://github.com/gustavoflandal/Gmail-Scanner.git
cd Gmail-Scanner

# 2. Iniciar com Docker
docker-compose up --build -d

# 3. Acessar aplicação
http://localhost:8080

# 4. Fazer login
# Email: seu.email@gmail.com
# Senha de App: (gere em myaccount.google.com/apppasswords)
```

### Opção 2: Desenvolvimento Local

```bash
# Backend (Terminal 1)
cd Gmail-Scanner
go mod download
go run ./cmd/api/main.go

# Frontend (Terminal 2)
cd Gmail-Scanner/web
npm install
npm run dev

# Acessar: http://localhost:5173
```

📖 **Guia completo:** [SETUP_IMAP.md](SETUP_IMAP.md)

---

## 🎯 Como Usar

### 1. **Primeiro Login**
- Acesse a aplicação
- Clique em "Fazer Login"
- Insira seu email e senha de app do Google
- Sistema conecta via IMAP (porta 993 SSL)

### 2. **Varredura de Emails**
- No Dashboard, clique em "Selecionar Pastas para Escanear"
- Escolha uma ou mais pastas (INBOX, DevOps, Sent, etc)
- Clique em "🚀 Iniciar Varredura Manual"
- Acompanhe progresso em tempo real
- Cancele a qualquer momento se necessário

### 3. **Visualizar Mensagens**
- Vá para aba "Mensagens"
- Use filtros por pasta ou busca por texto
- Clique em "Abrir" para ver email no Gmail
- Delete emails indexados se necessário

---

## 📡 API Endpoints

### Autenticação
```http
POST /api/auth/login          # Login com email + senha IMAP
POST /api/auth/logout         # Logout
```

### Varredura
```http
POST /api/scan                # Inicia varredura (body: {"folders": ["INBOX"]})
POST /api/scan-cancel         # Cancela varredura em andamento
GET  /api/scan-status         # Status da varredura
GET  /api/scan-progress       # Progresso detalhado (%, pasta atual, emails)
GET  /api/folders             # Lista todas as pastas IMAP disponíveis
```

### Mensagens
```http
GET    /api/messages?page=1&q=search    # Lista emails com paginação
DELETE /api/messages/{id}               # Deleta email do banco
```

### Sistema
```http
GET /api/health               # Health check
GET /api/stats                # Estatísticas do banco
```

### Exemplos de Uso

```bash
# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "seu@gmail.com", "password": "senha-app"}'

# Listar pastas disponíveis
curl http://localhost:8080/api/folders \
  -H "Authorization: Bearer YOUR_TOKEN"

# Iniciar varredura em múltiplas pastas
curl -X POST http://localhost:8080/api/scan \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"folders": ["INBOX", "DevOps", "[Gmail]/Sent Mail"]}'

# Ver progresso
curl http://localhost:8080/api/scan-progress \
  -H "Authorization: Bearer YOUR_TOKEN"

# Cancelar varredura
curl -X POST http://localhost:8080/api/scan-cancel \
  -H "Authorization: Bearer YOUR_TOKEN"

# Buscar emails
curl "http://localhost:8080/api/messages?q=invoice&page=1" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## 📁 Estrutura do Projeto

```
Gmail-Scanner/
├── cmd/
│   └── api/
│       └── main.go                    # Servidor HTTP + handlers
├── internal/
│   ├── auth/
│   │   └── simple.go                  # Autenticação JWT + IMAP
│   ├── imap/
│   │   └── client.go                  # Cliente IMAP (emersion/go-imap)
│   └── database/
│       └── db.go                      # SQLite com modernc.org/sqlite
├── web/
│   ├── src/
│   │   ├── pages/
│   │   │   ├── Dashboard.jsx         # Dashboard com seleção de pastas
│   │   │   ├── Messages.jsx          # Grid de mensagens com filtros
│   │   │   └── Login.jsx             # Tela de login IMAP
│   │   ├── services/
│   │   │   └── api.js                # Cliente API
│   │   └── utils/
│   │       └── storage.js            # Gerenciamento de tokens
│   ├── package.json
│   └── vite.config.js
├── data/
│   └── emails.db                      # Banco SQLite (criado automaticamente)
├── docker-compose.yml                 # Configuração Docker
├── Dockerfile                         # Multi-stage build
├── SETUP_IMAP.md                      # Guia de setup detalhado
└── README.md                          # Este arquivo
```

---

## 🔒 Segurança

### Senha de App vs Senha Principal

- ✅ **Senha de App**: Específica para esta aplicação
- ✅ **Revogável**: Pode revogar sem afetar outras apps
- ✅ **Escopo Limitado**: Acesso apenas a email via IMAP
- ✅ **2FA Obrigatório**: Requer verificação em 2 etapas ativada
- ✅ **Mais Seguro**: Não expõe sua senha principal do Google

### Gerar Senha de App

1. Ative 2FA: [myaccount.google.com/security](https://myaccount.google.com/security)
2. Gere senha: [myaccount.google.com/apppasswords](https://myaccount.google.com/apppasswords)
3. Copie a senha de 16 caracteres
4. Use no login da aplicação

### Revogar Acesso

- Acesse [myaccount.google.com/apppasswords](https://myaccount.google.com/apppasswords)
- Encontre "Gmail Scanner"
- Clique em "Remover"

---

## ⚙️ Configuração Avançada

### Variáveis de Ambiente (.env)

```env
# JWT Secret (mude em produção!)
JWT_SECRET=change-this-secret-in-production

# IMAP (fixo para Gmail)
IMAP_HOST=imap.gmail.com
IMAP_PORT=993

# Aplicação
APP_ENV=production
LOG_LEVEL=info
```

### Persistência de Dados

Os emails são armazenados em `./data/emails.db` (SQLite).

**Docker Volume:**
```yaml
volumes:
  - ./data:/app/data  # Persiste dados entre restarts
```

---

## 🐛 Solução de Problemas

### Erro: "Falha na autenticação"
- ✅ Verifique se IMAP está ativo no Gmail
- ✅ Gere uma nova senha de app
- ✅ Copie sem espaços
- ✅ Confirme que 2FA está ativo

### Erro: "Connection refused" (porta 993)
- ✅ Firewall bloqueando porta 993
- ✅ Libere IMAP SSL no antivírus
- ✅ Teste conexão: `telnet imap.gmail.com 993`

### Varredura em 0%
- ✅ Pasta pode estar vazia
- ✅ Verifique logs: `docker logs -f gmail-scanner`
- ✅ Teste com pasta INBOX primeiro

### Nenhuma mensagem aparece
- ✅ Verifique banco: `ls -la ./data/`
- ✅ Execute nova varredura
- ✅ Verifique logs do backend

---

## 🔧 Desenvolvimento

### Compilar Backend

```bash
# Linux/Mac
CGO_ENABLED=0 go build -o gmail-scanner ./cmd/api

# Windows
go build -o gmail-scanner.exe ./cmd/api/main.go
```

### Build Frontend

```bash
cd web
npm run build
# Saída em: web/dist/
```

### Testar Localmente

```bash
# Backend
go run ./cmd/api/main.go

# Frontend (dev server)
cd web && npm run dev
```

---

## 🐳 Docker

### Build Manual

```bash
docker build -t gmail-scanner .
docker run -p 8080:8080 -v $(pwd)/data:/app/data gmail-scanner
```

### Docker Compose

```bash
# Iniciar
docker-compose up -d

# Ver logs
docker-compose logs -f

# Parar
docker-compose down

# Rebuild
docker-compose up --build -d
```

---

## 📊 Banco de Dados

### Estrutura SQLite

```sql
CREATE TABLE emails (
    message_id TEXT PRIMARY KEY,
    from_addr TEXT,
    title TEXT,
    subject TEXT,
    link TEXT,
    folder TEXT,
    timestamp TEXT,
    snippet_preview TEXT,
    is_read BOOLEAN,
    created_at TEXT
);

CREATE INDEX idx_folder ON emails(folder);
CREATE INDEX idx_timestamp ON emails(timestamp);
CREATE INDEX idx_title ON emails(title);
```

### Consultar Manualmente

```bash
# Entrar no container
docker exec -it gmail-scanner sh

# Abrir banco
sqlite3 /app/data/emails.db

# Consultas
SELECT COUNT(*) FROM emails;
SELECT folder, COUNT(*) FROM emails GROUP BY folder;
SELECT * FROM emails WHERE folder = 'INBOX' LIMIT 10;
```

---

## 🔄 Migração OAuth → IMAP

Se você usava a versão OAuth antiga:

1. ✅ Nova autenticação é mais simples (sem Google Cloud Console)
2. ✅ Não precisa mais de Client ID / Client Secret
3. ✅ Apenas email + senha de app
4. ✅ Dados antigos continuam no banco (compatível)

---

## 📝 Comparação: OAuth vs IMAP

| Aspecto | OAuth (Antigo) | IMAP (Novo) |
|---------|----------------|-------------|
| Setup | 30+ minutos | 5 minutos |
| Google Cloud Console | ✅ Necessário | ❌ Não necessário |
| Aprovação Google | ✅ Testadores | ❌ Não precisa |
| Credenciais | Client ID + Secret | Email + Senha App |
| Complexidade | Alta | Baixa |
| Manutenção | Token expira | Estável |

---

## 📄 Licença

MIT License - veja [LICENSE](LICENSE) para detalhes.

---

## 🤝 Contribuições

Contribuições são bem-vindas!

1. Fork o repositório
2. Crie uma branch: `git checkout -b feature/nova-funcionalidade`
3. Commit: `git commit -m 'Adiciona nova funcionalidade'`
4. Push: `git push origin feature/nova-funcionalidade`
5. Abra um Pull Request

---

## 📞 Suporte

- 🐛 **Bugs**: Abra uma [issue](https://github.com/gustavoflandal/Gmail-Scanner/issues)
- 💡 **Features**: Sugira via [discussions](https://github.com/gustavoflandal/Gmail-Scanner/discussions)
- 📖 **Docs**: Veja [SETUP_IMAP.md](SETUP_IMAP.md) e [DEPLOY_DOCKER.md](DEPLOY_DOCKER.md)

---

## ⭐ Agradecimentos

- [emersion/go-imap](https://github.com/emersion/go-imap) - Cliente IMAP em Go
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) - SQLite em Go puro
- [React](https://react.dev) - Interface web
- [Tailwind CSS](https://tailwindcss.com) - Estilização

---

**Versão**: 0.4.0 (IMAP Complete Edition)  
**Última atualização**: 26 de novembro de 2025

### Com Docker (Recomendado)

```bash
# 1. Clonar repositório
git clone https://github.com/gustavoflandal/gmail-scanner.git
cd gmail-scanner

# 2. Configurar credenciais
cp .env.docker .env
# Editar .env com suas credenciais Google

# 3. Iniciar
./start.sh                    # Linux/Mac
# ou
start.bat                     # Windows

# 4. Acessar
# http://localhost:8080
```

Veja [QUICKSTART.md](QUICKSTART.md) para instruções detalhadas.

### Local (Desenvolvimento)

```bash
# Backend
cd gmail-scanner
go mod download
go run ./cmd/api

# Frontend (novo terminal)
cd gmail-scanner/web
npm install
npm run dev

# Serviços (novo terminal)
cd gmail-scanner
docker-compose up opensearch libretranslate
```

## Uso

### Endpoints da API

#### Autenticação
- `GET /api/auth/login` - Inicia fluxo de autenticação OAuth
- `GET /api/auth/callback` - Callback do OAuth (automático)

#### Varredura
- `POST /api/scan` - Inicia uma varredura manual
- `GET /api/scan-status` - Status da última varredura

#### Mensagens
- `GET /api/messages?page=1&q=search` - Lista emails com paginação e busca
- `DELETE /api/messages/{id}` - Deleta um email

#### Informações
- `GET /api/health` - Status de saúde da aplicação
- `GET /api/stats` - Estatísticas do banco de dados

### Exemplo de Uso

```bash
# Iniciar varredura
curl -X POST http://localhost:8080/api/scan

# Verificar status
curl http://localhost:8080/api/scan-status

# Buscar emails
curl "http://localhost:8080/api/messages?q=invoice"

# Ver estatísticas
curl http://localhost:8080/api/stats

# Verificar saúde
curl http://localhost:8080/api/health
```

## Estrutura do Projeto

```
gmail-scanner/
├── cmd/
│   └── api/
│       └── main.go              # Entrada principal da aplicação
├── internal/
│   ├── gmail/
│   │   ├── auth.go              # Autenticação OAuth 2.0
│   │   └── client.go            # Cliente da Gmail API
│   ├── elasticsearch/
│   │   └── client.go            # Cliente OpenSearch
│   ├── translation/
│   │   └── translator.go        # Cliente LibreTranslate
│   ├── config/
│   │   └── config.go            # Configurações da aplicação
│   └── scheduler/
│       └── scheduler.go         # Agendamento de varreduras
├── web/
│   ├── public/                  # Arquivos estáticos (futura interface)
│   └── src/                     # Código React (futuro)
├── docker-compose.yml           # Configuração Docker
├── Dockerfile                   # Build da aplicação
└── README.md                    # Este arquivo
```

## Recursos

- ✅ Autenticação segura com Google OAuth 2.0
- ✅ Varredura automática a cada 6 horas (configurável)
- ✅ Tradução de assuntos para português (offline com LibreTranslate)
- ✅ Armazenamento em OpenSearch para buscas rápidas
- ✅ API RESTful para integração
- ✅ Paginação e filtros avançados
- ✅ 100% gratuito - sem custos financeiros

## Configuração Avançada

### Mudar intervalo de varredura

Edite `.env`:

```env
SCAN_INTERVAL_HOURS=12  # Varrer a cada 12 horas
```

### Adicionar mais idiomas

Edite `docker-compose.yml`:

```yaml
libretranslate:
  environment:
    - LT_LOAD_ONLY=en,pt,es,fr,de,ja,zh
```

### Aumentar recurso de memória

Edite `docker-compose.yml`:

```yaml
opensearch:
  environment:
    - JAVA_OPTS=-Xms1024m -Xmx1024m
```

## Logs

Ver logs da aplicação:

```bash
docker-compose logs -f gmail-scanner-backend
```

Ver logs do OpenSearch:

```bash
docker-compose logs -f opensearch
```

## Solução de Problemas

### Erro: "Gmail client not initialized"
- Certifique-se de ter feito login via `http://localhost:8080/api/auth/login`
- Verifique se suas credenciais OAuth estão corretas em `.env`

### Erro: "OpenSearch connection refused"
- Verifique se o container está rodando: `docker-compose ps`
- Espere alguns segundos para o OpenSearch iniciar completamente

### Erro: "Translation service error"
- Verifique se LibreTranslate está rodando: `curl http://localhost:5000/languages`
- Reinicie o container: `docker-compose restart libretranslate`

## Desenvolvimento Local (sem Docker)

Para desenvolvimento local:

```bash
# Instalar dependências Go
go mod download

# Instalar e rodar OpenSearch localmente (opcional)
# Usar versão Docker é recomendado

# Rodar aplicação
go run ./cmd/api/main.go
```

## Licença

MIT License - veja LICENSE para detalhes

## Contribuições

Contribuições são bem-vindas! Por favor, faça um fork do repositório e crie um pull request.

## Suporte

Para relatar bugs ou sugerir features, abra uma issue no GitHub.
