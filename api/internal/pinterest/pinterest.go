package pinterest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

const productionBaseURL = "https://api.pinterest.com/v5"
const sandboxBaseURL = "https://api-sandbox.pinterest.com/v5"

type Client struct {
	tokenFunc func() string
	BaseURL   string
	HTTP      *http.Client
}

func NewClient(token string) *Client {
	return NewClientWithTokenFunc(func() string { return token })
}

func NewClientWithTokenFunc(fn func() string) *Client {
	return &Client{
		tokenFunc: fn,
		BaseURL:   productionBaseURL,
		HTTP:      &http.Client{Timeout: 5 * time.Minute},
	}
}

func (c *Client) SetSandbox(enabled bool) {
	if enabled {
		c.BaseURL = sandboxBaseURL
	} else {
		c.BaseURL = productionBaseURL
	}
}

func (c *Client) IsSandbox() bool {
	return c.BaseURL == sandboxBaseURL
}

func (c *Client) token() string {
	if c.tokenFunc != nil {
		return c.tokenFunc()
	}
	return ""
}

type RegisterVideoResponse struct {
	MediaID          string            `json:"media_id"`
	MediaType        string            `json:"media_type"`
	UploadURL        string            `json:"upload_url"`
	UploadParameters map[string]string `json:"upload_parameters"`
}

func (c *Client) RegisterVideo(ctx context.Context) (*RegisterVideoResponse, error) {
	body := map[string]string{"media_type": "video"}
	buf, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/media", bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("register video: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("pinterest register video http %d: %s", resp.StatusCode, string(raw))
	}

	var out RegisterVideoResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("parse register resp: %w", err)
	}
	return &out, nil
}

func (c *Client) UploadToS3(ctx context.Context, uploadURL string, params map[string]string, filename, contentType string, file io.Reader) error {
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	for k, v := range params {
		if err := mw.WriteField(k, v); err != nil {
			return fmt.Errorf("write field %s: %w", k, err)
		}
	}

	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		return fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(fw, file); err != nil {
		return fmt.Errorf("copy file: %w", err)
	}
	if err := mw.Close(); err != nil {
		return fmt.Errorf("close multipart: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.ContentLength = int64(buf.Len())

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("s3 upload: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("s3 upload http %d: %s", resp.StatusCode, truncate(string(body), 400))
	}
	return nil
}

type MediaStatus struct {
	MediaID   string `json:"media_id"`
	MediaType string `json:"media_type"`
	Status    string `json:"status"`
}

func (c *Client) WaitForMediaReady(ctx context.Context, mediaID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("media %s not ready after %s", mediaID, timeout)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/media/"+mediaID, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.token())

		resp, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 400 {
			return fmt.Errorf("pinterest media status http %d: %s", resp.StatusCode, string(raw))
		}

		var ms MediaStatus
		if err := json.Unmarshal(raw, &ms); err != nil {
			return fmt.Errorf("parse media status: %w", err)
		}

		switch ms.Status {
		case "succeeded":
			return nil
		case "failed":
			return fmt.Errorf("media %s processing failed", mediaID)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}
}

type CreatePinRequest struct {
	BoardID     string
	Title       string
	Description string
	Link        string
	MediaID     string
	CoverImage  string
}

type CreatePinResponse struct {
	ID          string `json:"id"`
	Link        string `json:"link"`
	Title       string `json:"title"`
	Description string `json:"description"`
	BoardID     string `json:"board_id"`
}

func (c *Client) CreateVideoPin(ctx context.Context, in CreatePinRequest) (*CreatePinResponse, error) {
	payload := map[string]any{
		"board_id":    in.BoardID,
		"title":       in.Title,
		"description": in.Description,
		"link":        in.Link,
		"media_source": map[string]any{
			"source_type":     "video_id",
			"cover_image_url": in.CoverImage,
			"media_id":        in.MediaID,
		},
	}

	buf, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/pins", bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("create pin: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("pinterest create pin http %d: %s", resp.StatusCode, string(raw))
	}

	var out CreatePinResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("parse create pin resp: %w", err)
	}
	return &out, nil
}

type Board struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ListBoardsResponse struct {
	Items []Board `json:"items"`
}

func (c *Client) ListBoards(ctx context.Context) ([]Board, error) {
	var all []Board
	for cursor := c.BaseURL + "/boards"; cursor != ""; {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, cursor, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+c.token())

		resp, err := c.HTTP.Do(req)
		if err != nil {
			return nil, fmt.Errorf("list boards: %w", err)
		}
		raw, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("pinterest list boards http %d: %s", resp.StatusCode, string(raw))
		}

		var page ListBoardsResponse
		if err := json.Unmarshal(raw, &page); err != nil {
			return nil, fmt.Errorf("parse boards resp: %w", err)
		}
		all = append(all, page.Items...)

		if len(page.Items) == 0 || resp.Header.Get("Link") == "" {
			break
		}
		next, err := extractNextPage(resp.Header.Get("Link"))
		if err != nil || next == "" {
			break
		}
		cursor = next
	}
	return all, nil
}

func extractNextPage(linkHeader string) (string, error) {
	if linkHeader == "" {
		return "", nil
	}
	for _, part := range strings.Split(linkHeader, ",") {
		part = strings.TrimSpace(part)
		if strings.Contains(part, `rel="next"`) {
			start := strings.Index(part, "<")
			end := strings.Index(part, ">")
			if start >= 0 && end > start {
				return part[start+1 : end], nil
			}
		}
	}
	return "", nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
