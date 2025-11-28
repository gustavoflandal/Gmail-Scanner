package scraper

import (
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

// Lista de User-Agents para rotação
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
}

// ArticleContent representa o conteúdo extraído de um artigo
type ArticleContent struct {
	Title       string
	Content     string // HTML do conteúdo principal
	ContentType string // "html" ou "text"
}

// getRandomUserAgent retorna um User-Agent aleatório
func getRandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}

// createHTTPClient cria um cliente HTTP otimizado para scraping
func createHTTPClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
			MinVersion:         tls.VersionTLS12,
		},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
	}

	return &http.Client{
		Timeout:   45 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			// Copiar headers para o redirect
			for key, val := range via[0].Header {
				req.Header[key] = val
			}
			return nil
		},
	}
}

// FetchArticleContent busca e extrai o conteúdo principal de um artigo
func FetchArticleContent(originalURL string) (*ArticleContent, error) {
	log.Infof("Fetching article content from: %s", originalURL)

	// Detectar o tipo de site e usar estratégia apropriada
	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	host := strings.ToLower(parsedURL.Host)

	// Estratégias específicas por site
	if strings.Contains(host, "medium.com") || strings.Contains(host, "towardsdatascience.com") ||
		strings.Contains(host, "levelup.gitconnected.com") || strings.Contains(host, "betterprogramming.pub") {
		return fetchMediumArticle(originalURL)
	}

	if strings.Contains(host, "dev.to") {
		return fetchDevToArticle(originalURL)
	}

	if strings.Contains(host, "github.com") {
		return fetchGitHubContent(originalURL)
	}

	if strings.Contains(host, "substack.com") || strings.Contains(host, ".substack.com") {
		return fetchSubstackArticle(originalURL)
	}

	// Fallback para scraping genérico
	return fetchGenericArticle(originalURL)
}

// fetchMediumArticle busca artigo do Medium usando técnicas avançadas
func fetchMediumArticle(articleURL string) (*ArticleContent, error) {
	log.Infof("Using Medium-specific strategy for: %s", articleURL)

	// Tentar primeiro o endpoint de exportação do Medium (formato texto limpo)
	// Medium tem um endpoint ?format=json que às vezes funciona
	client := createHTTPClient()

	// Estratégia 1: Tentar via Freedium (proxy que remove paywall)
	freediumURL := strings.Replace(articleURL, "medium.com", "freedium.cfd", 1)
	freediumURL = strings.Replace(freediumURL, "towardsdatascience.com", "freedium.cfd", 1)

	content, err := tryFetchWithHeaders(client, freediumURL, map[string]string{
		"User-Agent":      getRandomUserAgent(),
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.9",
		"Accept-Encoding": "gzip, deflate",
		"Connection":      "keep-alive",
		"Referer":         "https://www.google.com/",
	})

	if err == nil && content != nil && len(content.Content) > 500 {
		log.Info("Successfully fetched via Freedium proxy")
		return content, nil
	}

	// Estratégia 2: Tentar scribe.rip (outro proxy para Medium)
	scribeURL := strings.Replace(articleURL, "medium.com", "scribe.rip", 1)
	scribeURL = strings.Replace(scribeURL, "towardsdatascience.com", "scribe.rip", 1)

	content, err = tryFetchWithHeaders(client, scribeURL, map[string]string{
		"User-Agent":      getRandomUserAgent(),
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.9",
		"Referer":         "https://www.google.com/",
	})

	if err == nil && content != nil && len(content.Content) > 500 {
		log.Info("Successfully fetched via Scribe.rip proxy")
		return content, nil
	}

	// Estratégia 3: Tentar direto com headers de cache do Google
	content, err = tryFetchWithHeaders(client, articleURL, map[string]string{
		"User-Agent":                getRandomUserAgent(),
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
		"Accept-Language":           "en-US,en;q=0.9",
		"Accept-Encoding":           "gzip, deflate, br",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "cross-site",
		"Sec-Fetch-User":            "?1",
		"Cache-Control":             "max-age=0",
		"Referer":                   "https://www.google.com/",
	})

	if err == nil && content != nil && len(content.Content) > 200 {
		return content, nil
	}

	// Estratégia 4: Usar Google Cache
	googleCacheURL := fmt.Sprintf("https://webcache.googleusercontent.com/search?q=cache:%s", url.QueryEscape(articleURL))
	content, err = tryFetchWithHeaders(client, googleCacheURL, map[string]string{
		"User-Agent": getRandomUserAgent(),
		"Accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
	})

	if err == nil && content != nil && len(content.Content) > 500 {
		log.Info("Successfully fetched via Google Cache")
		return content, nil
	}

	return nil, fmt.Errorf("could not fetch Medium article after trying multiple strategies")
}

// fetchDevToArticle busca artigo do Dev.to
func fetchDevToArticle(articleURL string) (*ArticleContent, error) {
	log.Infof("Using Dev.to-specific strategy for: %s", articleURL)

	client := createHTTPClient()

	// Dev.to geralmente funciona bem com headers simples
	content, err := tryFetchWithHeaders(client, articleURL, map[string]string{
		"User-Agent":      getRandomUserAgent(),
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.9",
		"Accept-Encoding": "gzip, deflate",
		"Connection":      "keep-alive",
	})

	if err != nil {
		return nil, err
	}

	return content, nil
}

// fetchGitHubContent busca conteúdo do GitHub (README, arquivos, etc)
func fetchGitHubContent(githubURL string) (*ArticleContent, error) {
	log.Infof("Using GitHub-specific strategy for: %s", githubURL)

	client := createHTTPClient()
	parsedURL, _ := url.Parse(githubURL)
	path := parsedURL.Path

	// Verificar se é um repositório (para pegar o README)
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(pathParts) >= 2 {
		owner := pathParts[0]
		repo := pathParts[1]

		// Tentar buscar README via API do GitHub (não tem rate limit tão restrito para leitura)
		readmeAPIURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/readme", owner, repo)

		req, err := http.NewRequest("GET", readmeAPIURL, nil)
		if err == nil {
			req.Header.Set("Accept", "application/vnd.github.html+json")
			req.Header.Set("User-Agent", "Gmail-Scanner-Bot/1.0")

			resp, err := client.Do(req)
			if err == nil && resp.StatusCode == http.StatusOK {
				defer resp.Body.Close()
				body, _ := io.ReadAll(resp.Body)

				// A API retorna JSON com o conteúdo em base64 ou HTML
				var result map[string]interface{}
				if json.Unmarshal(body, &result) == nil {
					if htmlContent, ok := result["content"].(string); ok {
						// Decodificar base64 se necessário
						return &ArticleContent{
							Title:       fmt.Sprintf("%s/%s README", owner, repo),
							Content:     htmlContent,
							ContentType: "html",
						}, nil
					}
				}
			}
		}
	}

	// Fallback: scraping normal
	content, err := tryFetchWithHeaders(client, githubURL, map[string]string{
		"User-Agent":      getRandomUserAgent(),
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.9",
	})

	return content, err
}

// fetchSubstackArticle busca artigo do Substack
func fetchSubstackArticle(articleURL string) (*ArticleContent, error) {
	log.Infof("Using Substack-specific strategy for: %s", articleURL)

	client := createHTTPClient()

	// Substack geralmente permite acesso ao conteúdo público
	content, err := tryFetchWithHeaders(client, articleURL, map[string]string{
		"User-Agent":      getRandomUserAgent(),
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.9",
		"Accept-Encoding": "gzip, deflate",
		"Connection":      "keep-alive",
		"Referer":         "https://substack.com/",
	})

	if err != nil {
		return nil, err
	}

	return content, nil
}

// fetchGenericArticle usa scraping genérico
func fetchGenericArticle(articleURL string) (*ArticleContent, error) {
	log.Infof("Using generic strategy for: %s", articleURL)

	client := createHTTPClient()

	content, err := tryFetchWithHeaders(client, articleURL, map[string]string{
		"User-Agent":                getRandomUserAgent(),
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
		"Accept-Language":           "en-US,en;q=0.9,pt-BR;q=0.8",
		"Accept-Encoding":           "gzip, deflate",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
	})

	return content, err
}

// tryFetchWithHeaders tenta buscar conteúdo com headers específicos
func tryFetchWithHeaders(client *http.Client, targetURL string, headers map[string]string) (*ArticleContent, error) {
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	// Ler o corpo (com suporte a gzip)
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	}

	body, err := io.ReadAll(reader)
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
	content := extractMainContent(doc, targetURL)

	if content == "" {
		return nil, fmt.Errorf("could not extract article content")
	}

	log.Infof("Successfully extracted article content (%d chars) from: %s", len(content), targetURL)

	return &ArticleContent{
		Title:       title,
		Content:     content,
		ContentType: "html",
	}, nil
}

// extractTitle extrai o título do artigo
func extractTitle(doc *goquery.Document) string {
	// Tentar meta tags primeiro (mais confiável)
	if title, exists := doc.Find("meta[property='og:title']").Attr("content"); exists && title != "" {
		return strings.TrimSpace(title)
	}

	if title, exists := doc.Find("meta[name='twitter:title']").Attr("content"); exists && title != "" {
		return strings.TrimSpace(title)
	}

	// Tentar h1 dentro do artigo
	if title := doc.Find("article h1").First().Text(); title != "" {
		return strings.TrimSpace(title)
	}

	// Para Medium/Scribe
	if title := doc.Find("h1[data-testid='storyTitle']").First().Text(); title != "" {
		return strings.TrimSpace(title)
	}

	// Para Dev.to
	if title := doc.Find(".crayons-article__header h1").First().Text(); title != "" {
		return strings.TrimSpace(title)
	}

	// Para Substack
	if title := doc.Find(".post-title").First().Text(); title != "" {
		return strings.TrimSpace(title)
	}

	// Tentar h1 genérico
	if title := doc.Find("h1").First().Text(); title != "" {
		return strings.TrimSpace(title)
	}

	// Usar tag title (último recurso)
	titleText := doc.Find("title").First().Text()
	// Remover sufixos comuns
	titleText = regexp.MustCompile(`\s*[-|–]\s*(Medium|Dev\.to|GitHub|Substack).*$`).ReplaceAllString(titleText, "")
	return strings.TrimSpace(titleText)
}

// extractMainContent extrai o conteúdo principal do artigo
func extractMainContent(doc *goquery.Document, pageURL string) string {
	// Remover elementos indesejados primeiro
	removeUnwantedElements(doc)

	var content string
	host := strings.ToLower(pageURL)

	// Estratégias específicas por domínio
	if strings.Contains(host, "medium.com") || strings.Contains(host, "towardsdatascience.com") ||
		strings.Contains(host, "freedium") || strings.Contains(host, "scribe.rip") ||
		strings.Contains(host, "levelup.gitconnected.com") || strings.Contains(host, "betterprogramming.pub") {
		content = extractMediumContent(doc)
	} else if strings.Contains(host, "dev.to") {
		content = extractDevToContent(doc)
	} else if strings.Contains(host, "github.com") {
		content = extractGitHubHTMLContent(doc)
	} else if strings.Contains(host, "substack.com") {
		content = extractSubstackContent(doc)
	} else {
		content = extractGenericContent(doc)
	}

	return cleanContent(content)
}

// removeUnwantedElements remove elementos que não fazem parte do conteúdo principal
func removeUnwantedElements(doc *goquery.Document) {
	selectorsToRemove := []string{
		"script", "style", "noscript", "iframe",
		"nav", "header", "footer", "aside",
		".sidebar", ".comments", ".social-share",
		".advertisement", ".ad", ".ads", ".adsbygoogle",
		".cookie-banner", ".newsletter-signup", ".newsletter-cta",
		".related-posts", ".author-bio", ".author-card",
		".share-buttons", ".social-buttons",
		".popup", ".modal", ".overlay",
		"[role='banner']", "[role='navigation']", "[role='complementary']",
		".metabar", ".reactions", ".reaction-button",
		".crayons-article__aside", // Dev.to sidebar
		".pw-multi-vote-icon", ".pw-post-body-paragraph-highlight", // Medium
		".js-postMetaLockup", ".js-stickyFooter", // Medium
		"#lite-post-promo", ".js-postActionsBar", // Medium
	}

	for _, selector := range selectorsToRemove {
		doc.Find(selector).Remove()
	}
}

// extractMediumContent extrai conteúdo do Medium e proxies
func extractMediumContent(doc *goquery.Document) string {
	// Tentar seletores específicos do Scribe.rip (mais limpo)
	if content, _ := doc.Find(".main-content article").Html(); content != "" && len(content) > 200 {
		return content
	}

	// Freedium
	if content, _ := doc.Find(".main-content").Html(); content != "" && len(content) > 200 {
		return content
	}

	// Medium original - section com o artigo
	if content, _ := doc.Find("article section").Html(); content != "" && len(content) > 200 {
		return content
	}

	// Medium - article tag
	if content, _ := doc.Find("article").Html(); content != "" && len(content) > 200 {
		return content
	}

	// Tentar pegar por parágrafos do Medium
	var paragraphs []string
	doc.Find("article p, article h1, article h2, article h3, article pre, article code, article ul, article ol, article blockquote, article figure").Each(func(i int, s *goquery.Selection) {
		if html, _ := s.Html(); html != "" {
			tagName := goquery.NodeName(s)
			paragraphs = append(paragraphs, fmt.Sprintf("<%s>%s</%s>", tagName, html, tagName))
		}
	})

	if len(paragraphs) > 3 {
		return strings.Join(paragraphs, "\n")
	}

	return ""
}

// extractDevToContent extrai conteúdo do Dev.to
func extractDevToContent(doc *goquery.Document) string {
	// Seletor principal do Dev.to
	if content, _ := doc.Find("#article-body").Html(); content != "" {
		return content
	}

	if content, _ := doc.Find(".crayons-article__body").Html(); content != "" {
		return content
	}

	if content, _ := doc.Find(".crayons-article__main").Html(); content != "" {
		return content
	}

	// Fallback
	if content, _ := doc.Find("article").Html(); content != "" {
		return content
	}

	return ""
}

// extractGitHubHTMLContent extrai conteúdo HTML do GitHub
func extractGitHubHTMLContent(doc *goquery.Document) string {
	// README renderizado
	if content, _ := doc.Find(".markdown-body").First().Html(); content != "" {
		return content
	}

	// Box do README
	if content, _ := doc.Find("#readme .Box-body").Html(); content != "" {
		return content
	}

	// Arquivo markdown
	if content, _ := doc.Find("[data-target='readme-toc.content']").Html(); content != "" {
		return content
	}

	// Issues/PRs
	if content, _ := doc.Find(".comment-body").First().Html(); content != "" {
		return content
	}

	return ""
}

// extractSubstackContent extrai conteúdo do Substack
func extractSubstackContent(doc *goquery.Document) string {
	// Conteúdo do post
	if content, _ := doc.Find(".body.markup").Html(); content != "" {
		return content
	}

	if content, _ := doc.Find(".post-content").Html(); content != "" {
		return content
	}

	if content, _ := doc.Find(".available-content").Html(); content != "" {
		return content
	}

	// Fallback
	if content, _ := doc.Find("article").Html(); content != "" {
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
		".post-body",
		".article-body",
		".story-body",
		".content-body",
		"#content",
		"#main-content",
		".main-content",
		".content",
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

	// Remover elementos indesejados que podem ter sobrado
	doc.Find("script, style, .hidden, [hidden], .visually-hidden, svg.hidden, template").Remove()

	// Remover atributos de tracking
	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		attrsToRemove := []string{
			"onclick", "onload", "onerror", "onmouseover", "onfocus",
			"data-tracking", "data-analytics", "data-testid",
			"data-action", "data-controller", "data-target",
			"jsaction", "jsname", "jscontroller",
		}
		for _, attr := range attrsToRemove {
			s.RemoveAttr(attr)
		}
	})

	// Limpar imagens do Medium (converter data-src para src se necessário)
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		if dataSrc, exists := s.Attr("data-src"); exists && dataSrc != "" {
			s.SetAttr("src", dataSrc)
		}
		// Remover atributos de lazy loading
		s.RemoveAttr("data-src")
		s.RemoveAttr("loading")
	})

	html, _ := doc.Html()

	// Remover excesso de whitespace
	html = regexp.MustCompile(`\s+`).ReplaceAllString(html, " ")
	html = regexp.MustCompile(`>\s+<`).ReplaceAllString(html, "><")

	return strings.TrimSpace(html)
}
