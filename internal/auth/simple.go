package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gustavoflandal/gmail-scanner/internal/imap"
)

var (
	jwtSecret     []byte
	sessionsMutex sync.RWMutex
	sessions      = make(map[string]*Session)
)

// Session representa uma sessão de usuário autenticado
type Session struct {
	Email        string
	Password     string // Armazenado em memória para reconexões IMAP
	CreatedAt    time.Time
	LastActivity time.Time
	Token        string
}

// Claims representa as claims JWT
type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// LoginRequest representa dados de login
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse representa resposta de login
type LoginResponse struct {
	Token   string `json:"token"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// Init inicializa o sistema de autenticação
func Init(secret string) {
	if secret == "" {
		secret = "your-secret-key-change-in-production"
	}
	jwtSecret = []byte(secret)
}

// Authenticate valida credenciais IMAP e retorna token JWT
func Authenticate(email, password string) (*LoginResponse, error) {
	// Testar conexão IMAP
	if err := imap.TestConnection(email, password); err != nil {
		return nil, fmt.Errorf("falha na autenticação: credenciais inválidas ou IMAP não habilitado")
	}

	// Gerar token JWT
	token, err := generateJWT(email)
	if err != nil {
		return nil, fmt.Errorf("falha ao gerar token: %w", err)
	}

	// Criar sessão
	session := &Session{
		Email:        email,
		Password:     password, // Mantido para reconexões
		CreatedAt:    time.Now(),
		LastActivity: time.Now(),
		Token:        token,
	}

	sessionsMutex.Lock()
	sessions[token] = session
	sessionsMutex.Unlock()

	return &LoginResponse{
		Token:   token,
		Email:   email,
		Message: "Autenticação realizada com sucesso",
	}, nil
}

// generateJWT gera um token JWT para o usuário
func generateJWT(email string) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour) // 7 dias
	claims := &Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken valida um token JWT e retorna a sessão
func ValidateToken(tokenString string) (*Session, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token inválido: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token expirado ou inválido")
	}

	// Buscar sessão
	sessionsMutex.RLock()
	session, exists := sessions[tokenString]
	sessionsMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("sessão não encontrada")
	}

	// Atualizar última atividade
	sessionsMutex.Lock()
	session.LastActivity = time.Now()
	sessionsMutex.Unlock()

	return session, nil
}

// GetSession retorna sessão ativa do token
func GetSession(tokenString string) (*Session, error) {
	sessionsMutex.RLock()
	defer sessionsMutex.RUnlock()

	session, exists := sessions[tokenString]
	if !exists {
		return nil, fmt.Errorf("sessão não encontrada")
	}

	return session, nil
}

// Logout remove a sessão
func Logout(tokenString string) {
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()
	delete(sessions, tokenString)
}

// GetIMAPClient retorna um cliente IMAP conectado para a sessão
func (s *Session) GetIMAPClient() (*imap.Client, error) {
	return imap.Connect(s.Email, s.Password)
}

// SetAuthCookie define um cookie de autenticação
func SetAuthCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Mude para true em produção (HTTPS)
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7, // 7 dias
	}
	http.SetCookie(w, cookie)
}

// GetAuthCookie obtém o token de autenticação do cookie
func GetAuthCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// GetAuthToken obtém token do header Authorization ou cookie
func GetAuthToken(r *http.Request) (string, error) {
	// Tentar Authorization header primeiro
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Formato: "Bearer <token>"
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			return authHeader[7:], nil
		}
	}

	// Tentar cookie
	return GetAuthCookie(r)
}

// ClearAuthCookie remove o cookie de autenticação
func ClearAuthCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "auth_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
}

// CleanupExpiredSessions remove sessões inativas (executar periodicamente)
func CleanupExpiredSessions() {
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()

	now := time.Now()
	for token, session := range sessions {
		// Remover sessões inativas por mais de 7 dias
		if now.Sub(session.LastActivity) > 7*24*time.Hour {
			delete(sessions, token)
		}
	}
}

// GenerateState gera um estado aleatório para CSRF protection (mantido para compatibilidade)
func GenerateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
