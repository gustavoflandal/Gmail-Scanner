package imap

import (
	"crypto/tls"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

// Client representa um cliente IMAP conectado
type Client struct {
	conn  *client.Client
	email string
}

// Message representa uma mensagem de email
type Message struct {
	MessageID      string
	From           string
	Subject        string
	Date           time.Time
	Body           string
	SnippetPreview string
	Folder         string
	IsRead         bool
}

// Connect estabelece conexão com servidor IMAP do Gmail
func Connect(email, password string) (*Client, error) {
	log.Infof("Connecting to IMAP server for %s", email)

	// Conectar ao Gmail IMAP (SSL/TLS)
	conn, err := client.DialTLS("imap.gmail.com:993", &tls.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to IMAP: %w", err)
	}

	// Autenticar
	if err := conn.Login(email, password); err != nil {
		conn.Logout()
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	log.Infof("Successfully authenticated as %s", email)

	return &Client{
		conn:  conn,
		email: email,
	}, nil
}

// Close fecha a conexão IMAP
func (c *Client) Close() error {
	if c.conn != nil {
		c.conn.Logout()
	}
	return nil
}

// ListFolders retorna lista de todas as pastas/labels
func (c *Client) ListFolders() ([]string, error) {
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.conn.List("", "*", mailboxes)
	}()

	var folders []string
	for m := range mailboxes {
		folders = append(folders, m.Name)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to list folders: %w", err)
	}

	log.Infof("Found %d folders", len(folders))
	return folders, nil
}

// FetchMessages busca mensagens de uma pasta específica
// Se limit = 0, busca TODAS as mensagens
func (c *Client) FetchMessages(folder string, limit uint32) ([]*Message, error) {
	// Selecionar pasta
	mbox, err := c.conn.Select(folder, false)
	if err != nil {
		return nil, fmt.Errorf("failed to select folder %s: %w", folder, err)
	}

	if mbox.Messages == 0 {
		log.Infof("No messages in folder %s", folder)
		return []*Message{}, nil
	}

	// Determinar range de mensagens
	from := uint32(1)
	to := mbox.Messages

	// Se limit > 0, buscar apenas as últimas 'limit' mensagens
	if limit > 0 && mbox.Messages > limit {
		from = mbox.Messages - limit + 1
	}

	log.Infof("Fetching messages %d:%d from folder %s (total: %d)", from, to, folder, mbox.Messages)

	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)

	go func() {
		done <- c.conn.Fetch(seqset, []imap.FetchItem{
			imap.FetchEnvelope,
			imap.FetchFlags,
			imap.FetchUid,
			imap.FetchBodyStructure,
		}, messages)
	}()

	var result []*Message
	for msg := range messages {
		if msg == nil || msg.Envelope == nil {
			continue
		}

		// Construir mensagem
		message := &Message{
			MessageID: msg.Envelope.MessageId,
			Subject:   msg.Envelope.Subject,
			Date:      msg.Envelope.Date,
			Folder:    folder,
			IsRead:    false,
		}

		// Verificar se está lida
		for _, flag := range msg.Flags {
			if flag == imap.SeenFlag {
				message.IsRead = true
				break
			}
		}

		// Extrair remetente
		if len(msg.Envelope.From) > 0 {
			from := msg.Envelope.From[0]
			if from.PersonalName != "" {
				message.From = fmt.Sprintf("%s <%s@%s>", from.PersonalName, from.MailboxName, from.HostName)
			} else {
				message.From = fmt.Sprintf("%s@%s", from.MailboxName, from.HostName)
			}
		}

		// Usar subject como snippet por enquanto (mais rápido)
		// TODO: Implementar fetch de snippet em batch ou de forma assíncrona
		message.SnippetPreview = msg.Envelope.Subject
		if len(message.SnippetPreview) > 200 {
			message.SnippetPreview = message.SnippetPreview[:200] + "..."
		}

		result = append(result, message)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}

	log.Infof("Fetched %d messages from folder %s", len(result), folder)
	return result, nil
}

// fetchSnippet busca um preview do corpo da mensagem
func (c *Client) fetchSnippet(uid uint32, folder string) (string, error) {
	// Re-selecionar pasta se necessário
	c.conn.Select(folder, true)

	seqset := new(imap.SeqSet)
	seqset.AddNum(uid)

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message, 1)
	done := make(chan error, 1)

	go func() {
		done <- c.conn.UidFetch(seqset, items, messages)
	}()

	msg := <-messages
	if msg == nil {
		return "", fmt.Errorf("no message found")
	}

	r := msg.GetBody(section)
	if r == nil {
		return "", fmt.Errorf("no body found")
	}

	body, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	// Extrair texto simples (primeiros 200 caracteres)
	text := string(body)

	// Tentar encontrar conteúdo text/plain
	if strings.Contains(text, "Content-Type: text/plain") {
		parts := strings.Split(text, "\n\n")
		if len(parts) > 1 {
			text = parts[1]
		}
	}

	// Limpar e truncar
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\r", "")
	text = strings.ReplaceAll(text, "\n", " ")

	if len(text) > 200 {
		text = text[:200] + "..."
	}

	return text, nil
}

// FetchAllMessages busca mensagens de todas as pastas importantes
func (c *Client) FetchAllMessages(limit uint32) ([]*Message, error) {
	// Pastas principais do Gmail
	folders := []string{
		"INBOX",
		"[Gmail]/Sent Mail",
		"[Gmail]/Important",
		"[Gmail]/Starred",
	}

	var allMessages []*Message

	for _, folder := range folders {
		messages, err := c.FetchMessages(folder, limit)
		if err != nil {
			log.Warnf("Failed to fetch from folder %s: %v", folder, err)
			continue
		}
		allMessages = append(allMessages, messages...)
	}

	log.Infof("Fetched total of %d messages from all folders", len(allMessages))
	return allMessages, nil
}

// TestConnection testa se as credenciais são válidas
func TestConnection(email, password string) error {
	client, err := Connect(email, password)
	if err != nil {
		return err
	}
	defer client.Close()

	log.Info("Connection test successful")
	return nil
}
