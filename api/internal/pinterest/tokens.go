package pinterest

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

const oauthTokenURL = "https://api.pinterest.com/v5/oauth/token"

type tokenFile struct {
	AccessToken      string    `json:"access_token"`
	RefreshToken     string    `json:"refresh_token"`
	ExpiresAt        time.Time `json:"expires_at"`
	RefreshExpiresAt time.Time `json:"refresh_expires_at"`
}

type TokenStore struct {
	mu sync.RWMutex

	accessToken      string
	refreshToken     string
	expiresAt        time.Time
	refreshExpiresAt time.Time

	clientID     string
	clientSecret string
	filePath     string
}

func NewTokenStore(clientID, clientSecret, accessToken, refreshToken, filePath string) *TokenStore {
	return &TokenStore{
		clientID:     clientID,
		clientSecret: clientSecret,
		accessToken:  accessToken,
		refreshToken: refreshToken,
		filePath:     filePath,
	}
}

func (s *TokenStore) Load() {
	raw, err := os.ReadFile(s.filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("token-store: error reading %s: %v", s.filePath, err)
		}
		return
	}

	var tf tokenFile
	if err := json.Unmarshal(raw, &tf); err != nil {
		log.Printf("token-store: error parsing %s: %v", s.filePath, err)
		return
	}

	s.mu.Lock()
	if tf.AccessToken != "" {
		s.accessToken = tf.AccessToken
	}
	if tf.RefreshToken != "" {
		s.refreshToken = tf.RefreshToken
	}
	if !tf.ExpiresAt.IsZero() {
		s.expiresAt = tf.ExpiresAt
	}
	if !tf.RefreshExpiresAt.IsZero() {
		s.refreshExpiresAt = tf.RefreshExpiresAt
	}
	s.mu.Unlock()

	log.Println("token-store: loaded from", s.filePath)
}

func (s *TokenStore) save() error {
	s.mu.RLock()
	tf := tokenFile{
		AccessToken:      s.accessToken,
		RefreshToken:     s.refreshToken,
		ExpiresAt:        s.expiresAt,
		RefreshExpiresAt: s.refreshExpiresAt,
	}
	s.mu.RUnlock()

	raw, err := json.MarshalIndent(tf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, raw, 0600)
}

func (s *TokenStore) AccessToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.accessToken
}

func (s *TokenStore) RefreshToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.refreshToken
}

func (s *TokenStore) CanRefresh() bool {
	return s.clientID != "" && s.clientSecret != "" && s.refreshToken != ""
}

func (s *TokenStore) Refresh(httpClient *http.Client) error {
	if !s.CanRefresh() {
		return fmt.Errorf("cannot refresh: missing client_id, client_secret or refresh_token")
	}

	creds := base64.StdEncoding.EncodeToString([]byte(s.clientID + ":" + s.clientSecret))

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {s.RefreshToken()},
	}

	req, err := http.NewRequest(http.MethodPost, oauthTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("build refresh request: %w", err)
	}
	req.Header.Set("Authorization", "Basic "+creds)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("refresh request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("refresh http %d: %s", resp.StatusCode, truncate(string(body), 400))
	}

	var tok struct {
		AccessToken           string `json:"access_token"`
		RefreshToken          string `json:"refresh_token"`
		ExpiresIn             int    `json:"expires_in"`
		RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"`
	}
	if err := json.Unmarshal(body, &tok); err != nil {
		return fmt.Errorf("parse refresh response: %w", err)
	}
	if tok.AccessToken == "" {
		return fmt.Errorf("refresh returned empty access_token")
	}

	now := time.Now()
	s.mu.Lock()
	s.accessToken = tok.AccessToken
	if tok.RefreshToken != "" {
		s.refreshToken = tok.RefreshToken
	}
	s.expiresAt = now.Add(time.Duration(tok.ExpiresIn) * time.Second)
	if tok.RefreshTokenExpiresIn > 0 {
		s.refreshExpiresAt = now.Add(time.Duration(tok.RefreshTokenExpiresIn) * time.Second)
	}
	s.mu.Unlock()

	if err := s.save(); err != nil {
		log.Printf("token-store: warning: failed to persist: %v", err)
	}

	log.Printf("token-store: refreshed (expires %s, refresh_expires %s)",
		s.expiresAt.Format(time.RFC3339),
		s.refreshExpiresAt.Format(time.RFC3339),
	)
	return nil
}

func (s *TokenStore) timeUntilRefresh() time.Duration {
	s.mu.RLock()
	exp := s.expiresAt
	hasExp := !exp.IsZero()
	s.mu.RUnlock()

	if !hasExp {
		return 7 * 24 * time.Hour
	}
	remaining := time.Until(exp)
	if remaining <= 0 {
		return 0
	}
	wait := time.Duration(float64(remaining) * 0.85)
	if wait < time.Hour {
		return time.Hour
	}
	return wait
}

func (s *TokenStore) RunRefresher(ctx context.Context, httpClient *http.Client) {
	if !s.CanRefresh() {
		log.Println("token-store: auto-refresh disabled (missing credentials)")
		return
	}
	log.Println("token-store: background refresher started")

	for {
		delay := s.timeUntilRefresh()
		if delay <= 0 {
			log.Println("token-store: refreshing now")
			if err := s.Refresh(httpClient); err != nil {
				log.Printf("token-store: refresh failed: %v (retry 1h)", err)
				delay = time.Hour
			} else {
				delay = s.timeUntilRefresh()
			}
		}

		select {
		case <-ctx.Done():
			log.Println("token-store: refresher stopped")
			return
		case <-time.After(delay):
		}
	}
}

func (s *TokenStore) Info() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info := map[string]any{
		"has_access_token":  s.accessToken != "",
		"has_refresh_token": s.refreshToken != "",
		"can_auto_refresh":  s.CanRefresh(),
	}
	if !s.expiresAt.IsZero() {
		info["expires_at"] = s.expiresAt.Format(time.RFC3339)
		info["expires_in"] = time.Until(s.expiresAt).Round(time.Hour).String()
	}
	if !s.refreshExpiresAt.IsZero() {
		info["refresh_expires_at"] = s.refreshExpiresAt.Format(time.RFC3339)
		info["refresh_token_expires_in"] = time.Until(s.refreshExpiresAt).Round(time.Hour).String()
	}
	return info
}
