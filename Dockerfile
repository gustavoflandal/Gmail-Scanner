# Frontend build stage
FROM node:20-alpine AS frontend-builder

WORKDIR /app/web

# Copiar arquivos do frontend
COPY web/package*.json ./
RUN npm ci

COPY web ./
RUN npm run build

# Backend build stage
FROM golang:1.24-alpine AS backend-builder

WORKDIR /app

# Instalar dependências de build
RUN apk add --no-cache git

# Copiar go.mod e go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fonte
COPY . .

# Build com CGO desabilitado para usar modernc.org/sqlite
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o gmail-scanner ./cmd/api

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl tzdata

WORKDIR /app

# Copiar binary do backend
COPY --from=backend-builder /app/gmail-scanner .

# Copiar arquivos estáticos do frontend
COPY --from=frontend-builder /app/web/dist ./web/public

# Criar diretórios necessários
RUN mkdir -p ./data

# Expor porta
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:8080/api/health || exit 1

# Executar aplicação
CMD ["./gmail-scanner"]
