package threeapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client communicates with 3api.shop's internal pet API.
// The TAI backend uses this to provision pet compute accounts
// and convert TAI tokens into API call credits.
type Client struct {
	baseURL    string
	secret     string
	httpClient *http.Client
}

// Config holds 3api connection settings.
type Config struct {
	BaseURL string // e.g. "https://3api.shop"
	Secret  string // X-Internal-Secret shared key
	Timeout time.Duration
}

func NewClient(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &Client{
		baseURL: cfg.BaseURL,
		secret:  cfg.Secret,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// ─── Request/Response Types ────────────────────────────────────────

type ProvisionRequest struct {
	PetID     string `json:"pet_id"`
	OwnerTgID string `json:"owner_tg_id"`
	PetName   string `json:"pet_name,omitempty"`
}

type ProvisionResponse struct {
	PetID   string `json:"pet_id"`
	UserID  int64  `json:"user_id"`
	APIKey  string `json:"api_key"`
	GroupID int64  `json:"group_id"`
	Status  string `json:"status"`
}

type CreditRequest struct {
	PetID          string  `json:"pet_id"`
	TAIAmount      float64 `json:"tai_amount"`
	CreditAmount   float64 `json:"credit_amount"`
	IdempotencyKey string  `json:"idempotency_key"`
}

type CreditResponse struct {
	PetID         string  `json:"pet_id"`
	Credited      float64 `json:"credited"`
	NewBalance    float64 `json:"new_balance"`
	TAISpentTotal float64 `json:"tai_spent_total"`
}

type PetStatus struct {
	PetID        string  `json:"pet_id"`
	Status       string  `json:"status"`
	Balance      float64 `json:"balance"`
	TAISpentTotal float64 `json:"tai_spent_total"`
	DailyTAIUsed float64 `json:"daily_tai_used"`
	DailyTALimit float64 `json:"daily_tai_limit"`
	APIKeyStatus string  `json:"api_key_status"`
}

type BatchUsageResponse struct {
	Pets []PetUsageItem `json:"pets"`
}

type PetUsageItem struct {
	PetID        string  `json:"pet_id"`
	Balance      float64 `json:"balance"`
	DailyTAIUsed float64 `json:"daily_tai_used"`
	Status       string  `json:"status"`
}

// ─── API Methods ───────────────────────────────────────────────────

// Provision creates a 3api account + API key for a new pet.
func (c *Client) Provision(ctx context.Context, petID, ownerTgID, petName string) (*ProvisionResponse, error) {
	var resp ProvisionResponse
	err := c.post(ctx, "/api/v1/internal/pet/provision", ProvisionRequest{
		PetID:     petID,
		OwnerTgID: ownerTgID,
		PetName:   petName,
	}, &resp)
	if err != nil {
		return nil, fmt.Errorf("provision pet %s: %w", petID, err)
	}
	return &resp, nil
}

// Credit converts TAI tokens into 3api compute balance.
// taiAmount: how many TAI the pet is spending.
// creditAmount: how much 3api balance to grant (based on exchange rate).
func (c *Client) Credit(ctx context.Context, petID string, taiAmount, creditAmount float64, idempotencyKey string) (*CreditResponse, error) {
	var resp CreditResponse
	err := c.post(ctx, "/api/v1/internal/pet/credit", CreditRequest{
		PetID:          petID,
		TAIAmount:      taiAmount,
		CreditAmount:   creditAmount,
		IdempotencyKey: idempotencyKey,
	}, &resp)
	if err != nil {
		return nil, fmt.Errorf("credit pet %s: %w", petID, err)
	}
	return &resp, nil
}

// GetStatus returns a pet's 3api account status and balance.
func (c *Client) GetStatus(ctx context.Context, petID string) (*PetStatus, error) {
	var resp PetStatus
	err := c.get(ctx, fmt.Sprintf("/api/v1/internal/pet/status/%s", petID), &resp)
	if err != nil {
		return nil, fmt.Errorf("get status pet %s: %w", petID, err)
	}
	return &resp, nil
}

// Suspend disables a pet's compute access.
func (c *Client) Suspend(ctx context.Context, petID string) error {
	return c.post(ctx, fmt.Sprintf("/api/v1/internal/pet/suspend/%s", petID), nil, nil)
}

// Reactivate re-enables a pet's compute access.
func (c *Client) Reactivate(ctx context.Context, petID string) error {
	return c.post(ctx, fmt.Sprintf("/api/v1/internal/pet/reactivate/%s", petID), nil, nil)
}

// BatchUsage queries usage for multiple pets.
func (c *Client) BatchUsage(ctx context.Context, petIDs []string) (*BatchUsageResponse, error) {
	var resp BatchUsageResponse
	err := c.post(ctx, "/api/v1/internal/pet/usage/batch", map[string][]string{"pet_ids": petIDs}, &resp)
	if err != nil {
		return nil, fmt.Errorf("batch usage: %w", err)
	}
	return &resp, nil
}

// ─── HTTP Helpers ──────────────────────────────────────────────────

func (c *Client) post(ctx context.Context, path string, body interface{}, out interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Secret", c.secret)

	return c.do(req, out)
}

func (c *Client) get(ctx context.Context, path string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Internal-Secret", c.secret)

	return c.do(req, out)
}

func (c *Client) do(req *http.Request, out interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("3api returned %d: %s", resp.StatusCode, string(data))
	}

	if out != nil {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}
