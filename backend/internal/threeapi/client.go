package threeapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Client is the TAI Protocol's proxy to 3api.shop.
// Phase 0 model: ONE platform API key, all pet calls proxied through here.
// Per-pet tracking is done internally; 3api only sees one account.
type Client struct {
	baseURL    string
	platformKey string // single 3api API key for the tai-pets group
	httpClient *http.Client

	// Internal usage tracking (flushed to DB periodically)
	mu       sync.Mutex
	usageMap map[string]*PetUsage // petID → accumulated usage
}

// PetUsage tracks per-pet API consumption in memory before DB flush.
type PetUsage struct {
	PetID       string
	Calls       int64
	TokensUsed  int64
	TAISpent    float64
	LastCallAt  time.Time
}

// Config holds 3api connection settings.
type Config struct {
	BaseURL     string // e.g. "https://api.3api.shop"
	PlatformKey string // the tai-pets group API key
	Timeout     time.Duration
}

func NewClient(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 60 * time.Second // AI calls can be slow
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.3api.shop"
	}
	return &Client{
		baseURL:     cfg.BaseURL,
		platformKey: cfg.PlatformKey,
		httpClient:  &http.Client{Timeout: cfg.Timeout},
		usageMap:    make(map[string]*PetUsage),
	}
}

// ─── Chat Completion (main proxy method) ───────────────────────────

type ChatRequest struct {
	PetID    string          `json:"pet_id"`
	Model    string          `json:"model"`
	Messages []ChatMessage   `json:"messages"`
	MaxTokens int            `json:"max_tokens,omitempty"`
	Temperature float64      `json:"temperature,omitempty"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Content    string `json:"content"`
	Model      string `json:"model"`
	TokensUsed int64  `json:"tokens_used"`
	TAICost    float64 `json:"tai_cost"`
	Duration   int64  `json:"duration_ms"`
}

// ChatCompletion proxies an OpenAI-compatible chat call to 3api,
// using the platform key, and records per-pet usage internally.
func (c *Client) ChatCompletion(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	start := time.Now()

	// Build the OpenAI-compatible request body
	body := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
	}
	if req.MaxTokens > 0 {
		body["max_tokens"] = req.MaxTokens
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}

	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	// Call 3api with platform key
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/v1/chat/completions", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.platformKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("3api request failed: %w", err)
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("3api returned %d: %s", resp.StatusCode, string(respData))
	}

	// Parse OpenAI response
	var aiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int64 `json:"total_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}
	if err := json.Unmarshal(respData, &aiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	content := ""
	if len(aiResp.Choices) > 0 {
		content = aiResp.Choices[0].Message.Content
	}

	tokensUsed := aiResp.Usage.TotalTokens
	taiCost := CalculateTAICost(req.Model, tokensUsed)
	duration := time.Since(start).Milliseconds()

	// Record per-pet usage
	c.recordUsage(req.PetID, tokensUsed, taiCost)

	return &ChatResponse{
		Content:    content,
		Model:      aiResp.Model,
		TokensUsed: tokensUsed,
		TAICost:    taiCost,
		Duration:   duration,
	}, nil
}

// ─── Usage Tracking ────────────────────────────────────────────────

func (c *Client) recordUsage(petID string, tokens int64, taiCost float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	u, ok := c.usageMap[petID]
	if !ok {
		u = &PetUsage{PetID: petID}
		c.usageMap[petID] = u
	}
	u.Calls++
	u.TokensUsed += tokens
	u.TAISpent += taiCost
	u.LastCallAt = time.Now()
}

// FlushUsage returns accumulated usage and resets the map.
// Called periodically by the backend to persist to DB.
func (c *Client) FlushUsage() []*PetUsage {
	c.mu.Lock()
	defer c.mu.Unlock()

	result := make([]*PetUsage, 0, len(c.usageMap))
	for _, u := range c.usageMap {
		result = append(result, u)
	}
	c.usageMap = make(map[string]*PetUsage)
	return result
}

// GetPetUsage returns current accumulated usage for a pet (without flushing).
func (c *Client) GetPetUsage(petID string) *PetUsage {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.usageMap[petID]
}

// ─── Platform Account Health ───────────────────────────────────────

// CheckBalance calls 3api internal API to verify platform account health.
func (c *Client) CheckBalance(ctx context.Context, internalSecret string) (float64, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/api/v1/internal/tai/balance-check", nil)
	if err != nil {
		return 0, false, err
	}
	req.Header.Set("X-Internal-Secret", internalSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, false, err
	}
	defer resp.Body.Close()

	var result struct {
		Balance    float64 `json:"balance"`
		Sufficient bool    `json:"sufficient"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, false, err
	}
	return result.Balance, result.Sufficient, nil
}
