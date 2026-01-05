package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OllamaClient struct {
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

func NewOllamaClient(baseURL, model string) *OllamaClient {
	return &OllamaClient{
		BaseURL:    baseURL,
		Model:      model,
		HTTPClient: &http.Client{Timeout: 120 * time.Second},
	}
}

// --- DTO ---

type ollamaOptions struct {
	Temperature float64  `json:"temperature"`
	NumPredict  int      `json:"num_predict"`
	NumCtx      int      `json:"num_ctx"`
	Stop        []string `json:"stop"`
	TopP        float64  `json:"top_p"`
}

type ollamaChatRequest struct {
	Model     string        `json:"model"`
	Messages  []ChatMessage `json:"messages"`
	Stream    bool          `json:"stream"`
	KeepAlive string        `json:"keep_alive"`
	Options   ollamaOptions `json:"options"`
}

type ollamaChatResponse struct {
	Message ChatMessage `json:"message"`
}

// --- Public Methods ---

// 1. АНАЛИЗ (Шаг 1)
func (c *OllamaClient) AnalyzeFax(req FaxRequest) (FaxAnalysisResponse, error) {
	msg := BuildAnalyzeMessages(req)
	// Temp 0.1 - нужна точность
	raw, err := c.chatRaw(msg, 0.1, 150)
	if err != nil {
		return FaxAnalysisResponse{}, err
	}

	summary := parseField(raw, "Summary:")
	urgency := parseField(raw, "Urgency:")
	if summary == "" {
		summary = raw
	}
	if urgency == "" {
		urgency = "Low"
	}

	return FaxAnalysisResponse{Summary: summary, Urgency: urgency}, nil
}

// 2. ОТВЕТ (Шаг 2)
func (c *OllamaClient) GenerateReply(req FaxReplyRequest) (string, error) {
	msg := BuildReplyMessages(req)
	// Было 300, ставим 450. Этого хватит даже на длинную бюрократию.
	return c.chatRaw(msg, 0.8, 400)
}

// --- Helpers ---

func (c *OllamaClient) chatRaw(msgs []ChatMessage, temp float64, tokens int) (string, error) {
	reqBody := ollamaChatRequest{
		Model:     c.Model,
		Messages:  msgs,
		Stream:    false,
		KeepAlive: "-1m",
		Options: ollamaOptions{
			Temperature: temp,
			NumPredict:  tokens,
			NumCtx:      2048,
			TopP:        0.9,
			Stop:        []string{"User:", "Sender:", "Original Fax:"},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	resp, err := c.HTTPClient.Post(c.BaseURL+"/api/chat", "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("ollama post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama error %d: %s", resp.StatusCode, string(body))
	}

	var ollamaResp ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}

	return cleanOutput(ollamaResp.Message.Content), nil
}

func parseField(text, key string) string {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), key) {
			val := strings.TrimPrefix(strings.TrimSpace(line), key)
			return strings.TrimSpace(val)
		}
	}
	return ""
}

func cleanOutput(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 1 && strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		s = s[1 : len(s)-1]
	}
	return s
}
