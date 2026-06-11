// Package ai encapsula chamadas ao OpenRouter (compat. OpenAI Chat Completions).
// Gera 2 versões de título + descrição e escolhe a board mais adequada para um Pin.
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const openRouterURL = "https://openrouter.ai/api/v1/chat/completions"

// Client é o cliente HTTP para OpenRouter.
type Client struct {
	APIKey  string
	Model   string
	Referer string
	Title   string
	HTTP    *http.Client
}

// NewClient cria o client com timeout razoável.
func NewClient(apiKey, model, referer, title string) *Client {
	return &Client{
		APIKey:  apiKey,
		Model:   model,
		Referer: referer,
		Title:   title,
		HTTP:    &http.Client{Timeout: 60 * time.Second},
	}
}

// PinGenRequest é o input que o handler passa para o AI.
type PinGenRequest struct {
	ProductName        string `json:"productName"`
	AffiliateLink      string `json:"affiliateLink"`
	InfluencerHandle   string `json:"influencerHandle,omitempty"`
	ProductDescription string `json:"productDescription,omitempty"`
	ProductTags        string `json:"productTags,omitempty"`
	ProductImageURL    string `json:"productImageUrl"`
}

// PinGenResponse é o resultado estruturado que a AI deve devolver.
// Usamos JSON mode do OpenRouter para forçar esse formato.
type PinGenResponse struct {
	ProductName        string     `json:"productName"`
	ProductImageURL    string     `json:"productImageUrl"`
	ProductTags        string     `json:"productTags,omitempty"`
	ProductDescription string     `json:"productDescription,omitempty"`
	Board              string     `json:"board"` // slug da board
	VersionA           PinVersion `json:"versionA"`
	VersionB           PinVersion `json:"versionB"`
}

// PinVersion representa uma opção de título + descrição.
type PinVersion struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// Payloads e estruturas do OpenRouter (subset do que nos interessa).
type chatReq struct {
	Model          string       `json:"model"`
	Messages       []chatMsg    `json:"messages"`
	ResponseFormat *respFormat  `json:"response_format,omitempty"`
	Temperature    float64      `json:"temperature,omitempty"`
}

type chatMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type respFormat struct {
	Type string `json:"type"` // "json_object"
}

type chatResp struct {
	Choices []struct {
		Message chatMsg `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    any    `json:"code"`
	} `json:"error,omitempty"`
}

// Boards válidas (slug -> nome amigável). Só esses slugs são aceitos.
var validBoards = map[string]string{
	"viral-makeup-skincare-finds": "Viral Makeup & Skincare Finds",
	"wellness-health-essentials":  "Wellness & Health Essentials",
	"amazon-home-finds-hacks":     "Amazon Home Finds & Hacks",
	"aesthetic-self-care-routine": "Aesthetic Self-Care Routine",
	"genius-gadgets-viral-finds":  "Genius Gadgets & Viral Finds",
}

// GeneratePin chama o OpenRouter e devolve o PinGenResponse.
func (c *Client) GeneratePin(ctx context.Context, in PinGenRequest) (*PinGenResponse, error) {
	prompt := buildPrompt(in)

	body := chatReq{
		Model: c.Model,
		Messages: []chatMsg{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
		ResponseFormat: &respFormat{Type: "json_object"},
		Temperature:    0.8,
	}

	buf, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openRouterURL, bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")
	// OpenRouter pede esses headers para rastrear origem do tráfego.
	req.Header.Set("HTTP-Referer", c.Referer)
	req.Header.Set("X-Title", c.Title)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("openrouter http %d: %s", resp.StatusCode, truncate(string(raw), 400))
	}

	var cr chatResp
	if err := json.Unmarshal(raw, &cr); err != nil {
		return nil, fmt.Errorf("unmarshal chat resp: %w", err)
	}
	if cr.Error != nil {
		return nil, fmt.Errorf("openrouter error: %s", cr.Error.Message)
	}
	if len(cr.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned")
	}

	content := cr.Choices[0].Message.Content
	var out PinGenResponse
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return nil, fmt.Errorf("parse AI json: %w (content=%s)", err, truncate(content, 400))
	}

	// Enriquece com os dados originais (AI não precisa ecoar).
	out.ProductName = in.ProductName
	out.ProductImageURL = in.ProductImageURL
	if in.ProductTags != "" {
		out.ProductTags = in.ProductTags
	}
	if in.ProductDescription != "" {
		out.ProductDescription = in.ProductDescription
	}

	// Valida board
	if _, ok := validBoards[out.Board]; !ok {
		// Fallback: se AI devolver nome inválido, cai no "genius-gadgets"
		out.Board = "genius-gadgets-viral-finds"
	}

	// Valida tamanhos
	out.VersionA.Title = trim(out.VersionA.Title, 100)
	out.VersionB.Title = trim(out.VersionB.Title, 100)
	out.VersionA.Description = trim(out.VersionA.Description, 500)
	out.VersionB.Description = trim(out.VersionB.Description, 500)

	return &out, nil
}

const systemPrompt = `You are a Pinterest SEO and copywriting specialist for the US/CA/UK market, writing for the brand "Nuvea Finds" - a curated aesthetic of Amazon products (Beauty, Wellness, Home Gadgets).

Your job: given a product, produce TWO distinct versions of Pinterest Video Pin content, plus pick the best board.

RULES (strict):
- TITLE: max 100 chars. Hooky, keyword-rich (Amazon Finds, Must-Have, Viral, Hack). Do NOT use the literal product name - use the benefit or category.
- DESCRIPTION: max 500 chars. First 2-3 sentences matter most. Friendly, benefit-focused. MUST end with a clear CTA ("tap the link in our bio", "shop via the link", etc.). Include 5-7 English hashtags at the end.
- Version A: softer/beauty angle. Version B: punchier/scroll-stopping hook.
- Board: pick the best slug from this list:
  - viral-makeup-skincare-finds (makeup, skincare)
  - wellness-health-essentials (vitamins, supplements, fitness)
  - amazon-home-finds-hacks (kitchen, organization, decor)
  - aesthetic-self-care-routine (self-care, lifestyle, loungewear)
  - genius-gadgets-viral-finds (quirky gadgets, viral misc)

You MUST return ONLY a JSON object with this exact shape:
{
  "board": "<slug>",
  "versionA": {"title": "...", "description": "..."},
  "versionB": {"title": "...", "description": "..."}
}
No markdown, no prose, just the JSON.`

func buildPrompt(in PinGenRequest) string {
	var b strings.Builder
	b.WriteString("Product: ")
	b.WriteString(in.ProductName)
	b.WriteString("\nAmazon link: ")
	b.WriteString(in.AffiliateLink)
	if in.InfluencerHandle != "" {
		b.WriteString("\nOriginal TikTok creator: ")
		b.WriteString(in.InfluencerHandle)
	}
	if in.ProductTags != "" {
		b.WriteString("\nSEO keyword hints: ")
		b.WriteString(in.ProductTags)
	}
	if in.ProductDescription != "" {
		b.WriteString("\nProduct description from Amazon:\n")
		b.WriteString(in.ProductDescription)
	}
	b.WriteString("\n\nReturn the JSON now.")
	return b.String()
}

func trim(s string, max int) string {
	if len(s) <= max {
		return s
	}
	// Corta respeitando rune boundary.
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max])
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
