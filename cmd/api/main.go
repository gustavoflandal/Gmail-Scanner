package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gustavoflandal/gmail-scanner/internal/auth"
	"github.com/gustavoflandal/gmail-scanner/internal/database"
	"github.com/sirupsen/logrus"
)

var (
	log          *logrus.Logger
	db           *database.Database
	scanMutex    sync.Mutex
	scanStatus   *ScanStatus
	isScanning   bool
	cancelScan   chan bool
	scanProgress *ScanProgress
)

// ScanStatus representa o estado da varredura
type ScanStatus struct {
	IsRunning         bool      `json:"is_running"`
	LastScanTime      time.Time `json:"last_scan_time,omitempty"`
	LastEmailsScanned int       `json:"last_emails_scanned"`
	LastError         string    `json:"last_error,omitempty"`
}

// ScanProgress representa o progresso detalhado da varredura
type ScanProgress struct {
	CurrentFolder    string `json:"current_folder"`
	FoldersTotal     int    `json:"folders_total"`
	FoldersProcessed int    `json:"folders_processed"`
	EmailsTotal      int    `json:"emails_total"`
	EmailsProcessed  int    `json:"emails_processed"`
	PercentComplete  int    `json:"percent_complete"`
	Status           string `json:"status"`
}

// ScanRequest representa os parâmetros de varredura
type ScanRequest struct {
	Folders []string `json:"folders"`
}

func init() {
	log = logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})

	scanStatus = &ScanStatus{
		IsRunning: false,
	}

	scanProgress = &ScanProgress{
		Status: "idle",
	}

	cancelScan = make(chan bool, 1)
}

func main() {
	// Create data directory if needed
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		os.Mkdir("./data", 0755)
	}

	// Inicializar autenticação simples
	jwtSecret := os.Getenv("JWT_SECRET")
	auth.Init(jwtSecret)

	var err error
	db, err = database.NewDatabase("./data/emails.db")
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	defer db.Close()

	// Add some sample data if database is empty
	stats, _ := db.GetStats()
	if stats["total_emails"].(int) == 0 {
		addSampleData()
	}

	router := mux.NewRouter()
	router.Use(corsMiddleware)

	// Auth routes (públicas)
	router.HandleFunc("/api/auth/login", handleLogin).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/auth/logout", handleLogout).Methods("POST", "OPTIONS")

	// API routes (requerem autenticação)
	router.HandleFunc("/api/messages", authMiddleware(getMessages)).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/messages/{id}", authMiddleware(deleteMessage)).Methods("DELETE", "OPTIONS")
	router.HandleFunc("/api/scan", authMiddleware(startScan)).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/scan-status", authMiddleware(getScanStatus)).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/scan-progress", authMiddleware(getScanProgress)).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/scan-cancel", authMiddleware(cancelScanHandler)).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/folders", authMiddleware(getFolders)).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/stats", authMiddleware(getStats)).Methods("GET", "OPTIONS")

	// API routes públicas
	router.HandleFunc("/api/health", getHealth).Methods("GET", "OPTIONS")

	// Static files
	fs := http.FileServer(http.Dir("./web/public"))
	router.PathPrefix("/").Handler(fs)

	// Cleanup de sessões expiradas a cada hora
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			auth.CleanupExpiredSessions()
		}
	}()

	port := ":8080"
	log.Infof("Server listening on %s", port)
	log.Infof("Login endpoint: POST /api/auth/login")
	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// authMiddleware verifica autenticação
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetAuthToken(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "não autorizado"})
			return
		}

		session, err := auth.ValidateToken(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "token inválido"})
			return
		}

		// Adicionar email ao contexto (opcional)
		_ = session.Email

		next.ServeHTTP(w, r)
	}
}

func addSampleData() {
	emails := []*database.Email{
		{
			MessageID:      "1",
			From:           "user@example.com",
			Title:          "Welcome to Gmail Scanner",
			Subject:        "Welcome",
			Link:           "https://mail.google.com/mail/u/0/#inbox/1",
			Folder:         "INBOX",
			Timestamp:      "2025-01-20T10:30:00Z",
			SnippetPreview: "Welcome to Gmail Scanner. This is a powerful tool for scanning and organizing your emails.",
			IsRead:         false,
			CreatedAt:      "2025-01-20T10:30:00Z",
		},
		{
			MessageID:      "2",
			From:           "admin@example.com",
			Title:          "Project Status Update",
			Subject:        "Status",
			Link:           "https://mail.google.com/mail/u/0/#inbox/2",
			Folder:         "INBOX",
			Timestamp:      "2025-01-19T14:15:00Z",
			SnippetPreview: "Here is the latest status on the project. Everything is progressing well.",
			IsRead:         true,
			CreatedAt:      "2025-01-19T14:15:00Z",
		},
		{
			MessageID:      "3",
			From:           "team@example.com",
			Title:          "Meeting Notes",
			Subject:        "Notes",
			Link:           "https://mail.google.com/mail/u/0/#inbox/3",
			Folder:         "INBOX",
			Timestamp:      "2025-01-18T09:00:00Z",
			SnippetPreview: "Notes from today's team meeting. Action items and next steps are listed below.",
			IsRead:         false,
			CreatedAt:      "2025-01-18T09:00:00Z",
		},
	}

	for _, email := range emails {
		if err := db.IndexEmail(email); err != nil {
			log.Errorf("Failed to index sample email: %v", err)
		}
	}

	log.Infof("Added %d sample emails to database", len(emails))
}

func getMessages(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	query := r.URL.Query().Get("q")
	emails, total, err := db.SearchEmails(query, page, 20)
	if err != nil {
		log.Errorf("search error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "search failed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"emails":      emails,
		"total":       total,
		"page":        page,
		"page_size":   20,
		"total_pages": (total + 19) / 20,
	})
}

func deleteMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if err := db.DeleteEmail(id); err != nil {
		log.Errorf("delete error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "delete failed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted", "id": id})
}

func getHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func getStats(w http.ResponseWriter, r *http.Request) {
	stats, err := db.GetStats()
	if err != nil {
		log.Errorf("stats error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "stats failed"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// startScan inicia uma varredura manual de emails
func startScan(w http.ResponseWriter, r *http.Request) {
	scanMutex.Lock()
	if isScanning {
		scanMutex.Unlock()
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "varredura já em andamento"})
		return
	}
	isScanning = true
	scanStatus.IsRunning = true
	scanMutex.Unlock()

	// Parse request body para obter pastas selecionadas
	var scanReq ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&scanReq); err != nil {
		// Se não houver body, usar pastas padrão
		scanReq.Folders = []string{"INBOX"}
	}

	// Se nenhuma pasta foi especificada, usar INBOX
	if len(scanReq.Folders) == 0 {
		scanReq.Folders = []string{"INBOX"}
	}

	// Obter token e sessão
	token, err := auth.GetAuthToken(r)
	if err != nil {
		scanMutex.Lock()
		isScanning = false
		scanStatus.IsRunning = false
		scanMutex.Unlock()

		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "não autorizado"})
		return
	}

	session, err := auth.GetSession(token)
	if err != nil {
		scanMutex.Lock()
		isScanning = false
		scanStatus.IsRunning = false
		scanMutex.Unlock()

		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "sessão inválida"})
		return
	}

	// Limpar canal de cancelamento
	select {
	case <-cancelScan:
	default:
	}

	// Responder imediatamente
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "started",
		"message": "varredura iniciada",
		"folders": scanReq.Folders,
	})

	// Executar varredura em goroutine
	go performScan(session, scanReq.Folders)
}

// performScan executa a varredura de emails
func performScan(session *auth.Session, folders []string) {
	defer func() {
		scanMutex.Lock()
		isScanning = false
		scanStatus.IsRunning = false
		scanStatus.LastScanTime = time.Now()
		scanProgress.Status = "completed"
		scanMutex.Unlock()
	}()

	log.Infof("Starting email scan for %s in folders: %v", session.Email, folders)

	// Atualizar progresso inicial
	scanMutex.Lock()
	scanProgress.CurrentFolder = ""
	scanProgress.FoldersTotal = len(folders)
	scanProgress.FoldersProcessed = 0
	scanProgress.EmailsTotal = 0
	scanProgress.EmailsProcessed = 0
	scanProgress.PercentComplete = 0
	scanProgress.Status = "connecting"
	scanMutex.Unlock()

	// Conectar IMAP
	imapClient, err := session.GetIMAPClient()
	if err != nil {
		scanMutex.Lock()
		scanStatus.LastError = fmt.Sprintf("Falha ao conectar IMAP: %v", err)
		scanProgress.Status = "error"
		scanMutex.Unlock()
		log.Errorf("IMAP connection failed: %v", err)
		return
	}
	defer imapClient.Close()

	scanMutex.Lock()
	scanProgress.Status = "scanning"
	scanMutex.Unlock()

	totalEmailCount := 0

	// Processar cada pasta
	for i, folder := range folders {
		// Verificar se foi cancelado
		select {
		case <-cancelScan:
			log.Infof("Scan cancelled by user")
			scanMutex.Lock()
			scanStatus.LastError = "Varredura cancelada pelo usuário"
			scanProgress.Status = "cancelled"
			scanMutex.Unlock()
			return
		default:
		}

		scanMutex.Lock()
		scanProgress.CurrentFolder = folder
		scanProgress.FoldersProcessed = i
		scanProgress.PercentComplete = (i * 100) / len(folders)
		scanMutex.Unlock()

		log.Infof("Scanning folder: %s (%d/%d)", folder, i+1, len(folders))

		// Buscar TODAS as mensagens da pasta (limit = 0)
		messages, err := imapClient.FetchMessages(folder, 0)
		if err != nil {
			log.Warnf("Failed to fetch messages from %s: %v", folder, err)
			continue
		}

		log.Infof("Fetched %d messages from folder %s", len(messages), folder)

		scanMutex.Lock()
		scanProgress.EmailsTotal += len(messages)
		scanMutex.Unlock()

		// Salvar no banco de dados
		for j, msg := range messages {
			// Verificar cancelamento a cada 10 emails
			if j%10 == 0 {
				select {
				case <-cancelScan:
					log.Infof("Scan cancelled by user")
					scanMutex.Lock()
					scanStatus.LastError = "Varredura cancelada pelo usuário"
					scanProgress.Status = "cancelled"
					scanMutex.Unlock()
					return
				default:
				}
			}

			// Validar MessageID
			messageID := msg.MessageID
			if messageID == "" {
				// Gerar ID único se MessageID estiver vazio
				messageID = fmt.Sprintf("%s-%s-%d", folder, msg.Date.Format("20060102150405"), j)
				log.Warnf("Empty MessageID, generated: %s", messageID)
			}

			email := &database.Email{
				MessageID:      messageID,
				From:           msg.From,
				Title:          msg.Subject,
				Subject:        msg.Subject,
				Link:           fmt.Sprintf("https://mail.google.com/mail/u/0/#inbox/%s", msg.MessageID),
				Folder:         msg.Folder,
				Timestamp:      msg.Date.Format(time.RFC3339),
				SnippetPreview: msg.SnippetPreview,
				IsRead:         msg.IsRead,
				CreatedAt:      time.Now().Format(time.RFC3339),
			}

			if err := db.IndexEmail(email); err != nil {
				log.Warnf("Failed to index email %s: %v", email.MessageID, err)
				continue
			}

			totalEmailCount++
			scanMutex.Lock()
			scanProgress.EmailsProcessed++
			scanMutex.Unlock()

			// Log a cada 10 emails salvos
			if totalEmailCount%10 == 0 {
				log.Infof("Indexed %d emails so far...", totalEmailCount)
			}
		}
	}

	scanMutex.Lock()
	scanStatus.LastEmailsScanned = totalEmailCount
	scanStatus.LastError = ""
	scanProgress.FoldersProcessed = len(folders)
	scanProgress.PercentComplete = 100
	scanProgress.Status = "completed"
	scanMutex.Unlock()

	log.Infof("Scan completed: %d emails processed from %d folders", totalEmailCount, len(folders))
}

// getScanStatus retorna o status da varredura
func getScanStatus(w http.ResponseWriter, r *http.Request) {
	scanMutex.Lock()
	status := *scanStatus
	scanMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleLogin processa login com email e senha
func handleLogin(w http.ResponseWriter, r *http.Request) {
	var loginReq auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "requisição inválida"})
		return
	}

	if loginReq.Email == "" || loginReq.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "email e senha são obrigatórios"})
		return
	}

	log.Infof("Login attempt for %s", loginReq.Email)

	response, err := auth.Authenticate(loginReq.Email, loginReq.Password)
	if err != nil {
		log.Errorf("Authentication failed for %s: %v", loginReq.Email, err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// Definir cookie
	auth.SetAuthCookie(w, response.Token)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	log.Infof("User authenticated successfully: %s", loginReq.Email)
}

// handleLogout faz logout do usuário
func handleLogout(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetAuthToken(r)
	if err == nil {
		auth.Logout(token)
	}

	auth.ClearAuthCookie(w)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "logged_out"})
	log.Info("User logged out")
}

// getScanProgress retorna o progresso detalhado da varredura
func getScanProgress(w http.ResponseWriter, r *http.Request) {
	scanMutex.Lock()
	progress := *scanProgress
	scanMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(progress)
}

// cancelScanHandler cancela a varredura em andamento
func cancelScanHandler(w http.ResponseWriter, r *http.Request) {
	scanMutex.Lock()
	if !isScanning {
		scanMutex.Unlock()
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "nenhuma varredura em andamento"})
		return
	}
	scanMutex.Unlock()

	// Enviar sinal de cancelamento
	select {
	case cancelScan <- true:
		log.Info("Scan cancellation requested")
	default:
		// Canal já tem sinal pendente
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "cancelling",
		"message": "cancelamento solicitado",
	})
}

// getFolders retorna lista de pastas IMAP disponíveis
func getFolders(w http.ResponseWriter, r *http.Request) {
	// Obter token e sessão
	token, err := auth.GetAuthToken(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "não autorizado"})
		return
	}

	session, err := auth.GetSession(token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "sessão inválida"})
		return
	}

	// Conectar IMAP
	imapClient, err := session.GetIMAPClient()
	if err != nil {
		log.Errorf("IMAP connection failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "falha ao conectar IMAP"})
		return
	}
	defer imapClient.Close()

	// Listar pastas
	folders, err := imapClient.ListFolders()
	if err != nil {
		log.Errorf("Failed to list folders: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "falha ao listar pastas"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"folders": folders,
	})
}
