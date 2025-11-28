package scraper

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

// ArticleContent representa o conteúdo extraído de um artigo
type ArticleContent struct {
	Title       string
	Content     string // HTML do conteúdo principal
	ContentType string // "html" ou "text"
}

// FetchArticleContent busca e extrai o conteúdo principal de um artigo
func FetchArticleContent(url string) (*ArticleContent, error) {
	log.Infof("Fetching article content from: %s", url)

	// Cliente HTTP com timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Headers para simular um navegador
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch article: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extrair título
	title := extractTitle(doc)

	// Extrair conteúdo principal
	content := extractMainContent(doc, url)

	if content == "" {
		return nil, fmt.Errorf("could not extract article content")
	}

	log.Infof("Successfully extracted article content (%d chars) from: %s", len(content), url)

	return &ArticleContent{
		Title:       title,
		Content:     content,
		ContentType: "html",
	}, nil
}

// extractTitle extrai o título do artigo
func extractTitle(doc *goquery.Document) string {
	// Tentar meta tags primeiro
	if title, exists := doc.Find("meta[property='og:title']").Attr("content"); exists && title != "" {
		return strings.TrimSpace(title)
	}

	// Tentar h1 dentro do artigo
	if title := doc.Find("article h1").First().Text(); title != "" {
		return strings.TrimSpace(title)
	}

	// Tentar h1 genérico
	if title := doc.Find("h1").First().Text(); title != "" {
		return strings.TrimSpace(title)
	}

	// Usar tag title
	return strings.TrimSpace(doc.Find("title").First().Text())
}

// extractMainContent extrai o conteúdo principal do artigo
func extractMainContent(doc *goquery.Document, url string) string {
	// Remover elementos indesejados
	doc.Find("script, style, nav, header, footer, aside, .sidebar, .comments, .social-share, .advertisement, .ad, .ads, .cookie-banner, .newsletter-signup, .related-posts, .author-bio, noscript, iframe").Remove()

	var content string

	// Estratégias específicas por domínio
	if strings.Contains(url, "medium.com") || strings.Contains(url, "towardsdatascience.com") {
		content = extractMediumContent(doc)
	} else if strings.Contains(url, "dev.to") {
		content = extractDevToContent(doc)
	} else if strings.Contains(url, "github.com") {
		content = extractGitHubContent(doc)
	} else if strings.Contains(url, "substack.com") {
		content = extractSubstackContent(doc)
	} else {
		content = extractGenericContent(doc)
	}

	return cleanContent(content)
}

// extractMediumContent extrai conteúdo do Medium
func extractMediumContent(doc *goquery.Document) string {
	// Medium usa article tag
	if article, _ := doc.Find("article").Html(); article != "" {
		return article
	}
	return ""
}

// extractDevToContent extrai conteúdo do Dev.to
func extractDevToContent(doc *goquery.Document) string {
	// Dev.to usa .crayons-article__main
	if content, _ := doc.Find(".crayons-article__main, .crayons-article__body, #article-body").First().Html(); content != "" {
		return content
	}
	if article, _ := doc.Find("article").Html(); article != "" {
		return article
	}
	return ""
}

// extractGitHubContent extrai conteúdo do GitHub (README, etc)
func extractGitHubContent(doc *goquery.Document) string {
	// GitHub README
	if readme, _ := doc.Find(".markdown-body, .readme").First().Html(); readme != "" {
		return readme
	}
	return ""
}

// extractSubstackContent extrai conteúdo do Substack
func extractSubstackContent(doc *goquery.Document) string {
	if content, _ := doc.Find(".post-content, .body").First().Html(); content != "" {
		return content
	}
	return ""
}

// extractGenericContent tenta extrair conteúdo de forma genérica
func extractGenericContent(doc *goquery.Document) string {
	// Ordem de prioridade para encontrar o conteúdo principal
	selectors := []string{
		"article",
		"[role='main']",
		"main",
		".post-content",
		".article-content",
		".entry-content",
		".content",
		".post-body",
		".article-body",
		".story-body",
		"#content",
		"#main-content",
		".main-content",
	}

	for _, selector := range selectors {
		if content, _ := doc.Find(selector).First().Html(); content != "" && len(content) > 500 {
			return content
		}
	}

	// Última tentativa: pegar o body inteiro (limitado)
	if content, _ := doc.Find("body").Html(); content != "" {
		return content
	}

	return ""
}

// cleanContent limpa o HTML extraído
func cleanContent(content string) string {
	if content == "" {
		return ""
	}

	// Criar novo documento para limpar
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return content
	}

	// Remover mais elementos indesejados que podem ter sobrado
	doc.Find("script, style, .hidden, [hidden], .visually-hidden").Remove()

	// Remover atributos de tracking e estilos inline excessivos
	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		// Manter apenas atributos essenciais
		for _, attr := range []string{"onclick", "onload", "onerror", "data-tracking", "data-analytics"} {
			s.RemoveAttr(attr)
		}
	})

	html, _ := doc.Html()
	return strings.TrimSpace(html)
}
