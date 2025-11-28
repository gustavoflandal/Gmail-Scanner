package nosql

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

var log = logrus.New()

// Article representa um artigo salvo no NoSQL
type Article struct {
	ID          int64     `json:"id"`
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Domain      string    `json:"domain"`
	Newsletter  string    `json:"newsletter"`
	EmailDate   string    `json:"email_date"`
	Folder      string    `json:"folder"`
	Content     string    `json:"content"`      // Conteúdo HTML do artigo
	ContentType string    `json:"content_type"` // "html" ou "text"
	ImportedAt  time.Time `json:"imported_at"`
}

// NoSQLDB gerencia o banco de dados BBolt
type NoSQLDB struct {
	db *bolt.DB
	mu sync.RWMutex
}

const (
	bucketName = "articles"
)

// NewNoSQLDB cria uma nova instância do banco NoSQL
func NewNoSQLDB(dbPath string) (*NoSQLDB, error) {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("failed to open nosql database: %w", err)
	}

	// Criar bucket se não existir
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create bucket: %w", err)
	}

	log.Info("NoSQL database initialized successfully")
	return &NoSQLDB{db: db}, nil
}

// Close fecha o banco de dados
func (n *NoSQLDB) Close() error {
	if n.db != nil {
		return n.db.Close()
	}
	return nil
}

// ImportArticle importa um artigo para o banco NoSQL
func (n *NoSQLDB) ImportArticle(article Article) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	article.ImportedAt = time.Now()

	return n.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return fmt.Errorf("bucket not found")
		}

		data, err := json.Marshal(article)
		if err != nil {
			return fmt.Errorf("failed to marshal article: %w", err)
		}

		key := fmt.Sprintf("%d", article.ID)
		err = bucket.Put([]byte(key), data)
		if err != nil {
			return fmt.Errorf("failed to save article: %w", err)
		}

		log.Infof("Article imported: ID=%d, Title=%s", article.ID, article.Title)
		return nil
	})
}

// GetArticle recupera um artigo pelo ID
func (n *NoSQLDB) GetArticle(id int64) (*Article, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var article Article
	found := false

	err := n.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return nil
		}

		key := fmt.Sprintf("%d", id)
		data := bucket.Get([]byte(key))
		if data == nil {
			return nil
		}

		if err := json.Unmarshal(data, &article); err != nil {
			return fmt.Errorf("failed to unmarshal article: %w", err)
		}

		found = true
		return nil
	})

	if err != nil {
		return nil, err
	}

	if !found {
		return nil, nil
	}

	return &article, nil
}

// IsImported verifica se um artigo já foi importado
func (n *NoSQLDB) IsImported(id int64) bool {
	n.mu.RLock()
	defer n.mu.RUnlock()

	imported := false

	n.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return nil
		}

		key := fmt.Sprintf("%d", id)
		data := bucket.Get([]byte(key))
		imported = data != nil
		return nil
	})

	return imported
}

// GetImportedIDs retorna todos os IDs de artigos importados
func (n *NoSQLDB) GetImportedIDs() ([]int64, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var ids []int64

	err := n.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			log.Warn("GetImportedIDs: bucket not found")
			return nil
		}

		count := 0
		err := bucket.ForEach(func(k, v []byte) error {
			var article Article
			if err := json.Unmarshal(v, &article); err == nil {
				ids = append(ids, article.ID)
				count++
			}
			return nil
		})
		log.Infof("GetImportedIDs: found %d articles in bucket", count)
		return err
	})

	return ids, err
}

// GetAllImported retorna todos os artigos importados
func (n *NoSQLDB) GetAllImported() ([]Article, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var articles []Article

	err := n.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return nil
		}

		return bucket.ForEach(func(k, v []byte) error {
			var article Article
			if err := json.Unmarshal(v, &article); err == nil {
				articles = append(articles, article)
			}
			return nil
		})
	})

	return articles, err
}

// DeleteArticle remove um artigo do banco NoSQL
func (n *NoSQLDB) DeleteArticle(id int64) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	return n.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return nil
		}

		key := fmt.Sprintf("%d", id)
		return bucket.Delete([]byte(key))
	})
}

// GetStats retorna estatísticas do banco NoSQL
func (n *NoSQLDB) GetStats() (map[string]interface{}, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	stats := map[string]interface{}{
		"total_imported": 0,
	}

	n.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return nil
		}

		count := 0
		bucket.ForEach(func(k, v []byte) error {
			count++
			return nil
		})
		stats["total_imported"] = count
		return nil
	})

	return stats, nil
}
