package llm

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

// ChatRequest OpenAI 兼容的请求体
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse OpenAI 兼容的响应体
type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Chat 调用 OpenAI 兼容的 Chat Completions API，返回助手回复的文本
func Chat(ctx context.Context, baseURL, apiKey, model, userPrompt string) (string, error) {
	baseURL = strings.TrimSuffix(baseURL, "/")
	url := baseURL + "/chat/completions"
	reqBody := ChatRequest{
		Model: model,
		Messages: []Message{
			{Role: "user", Content: userPrompt},
		},
		Temperature: 0.2,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("llm: marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("llm: new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm: request: %w", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("llm: read body: %w", err)
	}
	var chatResp ChatResponse
	if err := json.Unmarshal(b, &chatResp); err != nil {
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("llm: http %d: %s", resp.StatusCode, string(b))
		}
		return "", fmt.Errorf("llm: unmarshal response: %w", err)
	}
	if chatResp.Error != nil {
		return "", fmt.Errorf("llm: api error: %s", chatResp.Error.Message)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("llm: http %d: %s", resp.StatusCode, string(b))
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("llm: empty choices")
	}
	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}
