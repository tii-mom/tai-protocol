package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

// TGUser represents a verified Telegram user from initData.
type TGUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	PhotoURL  string `json:"photo_url"`
}

// InitData contains all parsed fields from Telegram WebApp initData.
type InitData struct {
	User       TGUser
	QueryID    string
	AuthDate   int64
	Hash       string
	StartParam string
}

// VerifyInitData validates Telegram WebApp initData signature.
// Algorithm: https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app
//
// 1. Parse the query string into key=value pairs
// 2. Remove "hash" from the pairs
// 3. Sort remaining pairs alphabetically by key
// 4. Join with "\n" to form data_check_string
// 5. secret_key = HMAC-SHA256("WebAppData", bot_token)
// 6. computed_hash = HMAC-SHA256(secret_key, data_check_string)
// 7. Compare computed_hash with the provided hash
func VerifyInitData(initDataRaw string, botToken string) (*InitData, error) {
	// Parse query string
	values, err := url.ParseQuery(initDataRaw)
	if err != nil {
		return nil, fmt.Errorf("parse initData: %w", err)
	}

	hash := values.Get("hash")
	if hash == "" {
		return nil, fmt.Errorf("missing hash in initData")
	}

	// Build data_check_string: sorted key=value pairs excluding "hash"
	var pairs []string
	for key, vals := range values {
		if key == "hash" {
			continue
		}
		if len(vals) > 0 {
			pairs = append(pairs, key+"="+vals[0])
		}
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	// Compute secret key: HMAC-SHA256 with "WebAppData" as key
	secretKey := hmacSHA256([]byte("WebAppData"), []byte(botToken))

	// Compute hash: HMAC-SHA256 with secret_key
	computedHash := hmacSHA256(secretKey, []byte(dataCheckString))
	computedHashHex := hex.EncodeToString(computedHash)

	// Constant-time comparison
	if !hmac.Equal([]byte(computedHashHex), []byte(hash)) {
		return nil, fmt.Errorf("invalid initData signature")
	}

	// Check auth_date freshness (allow 24h window)
	var authDate int64
	if ad := values.Get("auth_date"); ad != "" {
		fmt.Sscanf(ad, "%d", &authDate)
	}
	if authDate > 0 && time.Now().Unix()-authDate > 86400 {
		return nil, fmt.Errorf("initData expired (auth_date too old)")
	}

	// Parse user JSON
	var user TGUser
	if userStr := values.Get("user"); userStr != "" {
		if err := json.Unmarshal([]byte(userStr), &user); err != nil {
			return nil, fmt.Errorf("parse user: %w", err)
		}
	}
	if user.ID == 0 {
		return nil, fmt.Errorf("no user in initData")
	}

	return &InitData{
		User:       user,
		QueryID:    values.Get("query_id"),
		AuthDate:   authDate,
		Hash:       hash,
		StartParam: values.Get("start_param"),
	}, nil
}

// VerifyBotToken validates a Telegram Bot API token format.
// Format: <bot_id>:<hash>
func VerifyBotTokenFormat(token string) error {
	parts := strings.SplitN(token, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid bot token format")
	}
	return nil
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
