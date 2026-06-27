package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const openaiAPI = "https://api.openai.com/v1/chat/completions"
const defaultModel = "gpt-4o"

type Client struct {
	apiKey string
	http   *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		http:   &http.Client{Timeout: 120 * time.Second},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model     string        `json:"model"`
	Messages  []chatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens"`
}

type chatChoice struct {
	Message chatMessage `json:"message"`
}

type chatResponse struct {
	Choices []chatChoice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) Complete(prompt string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY não configurada")
	}

	body, _ := json.Marshal(chatRequest{
		Model:     defaultModel,
		MaxTokens: 8192,
		Messages:  []chatMessage{{Role: "user", Content: prompt}},
	})

	req, err := http.NewRequest(http.MethodPost, openaiAPI, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("erro na chamada à API: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errResp chatResponse
		_ = json.Unmarshal(raw, &errResp)
		if errResp.Error != nil {
			return "", fmt.Errorf("OpenAI API %d: %s", resp.StatusCode, errResp.Error.Message)
		}
		return "", fmt.Errorf("OpenAI API %d: %s", resp.StatusCode, string(raw))
	}

	var result chatResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 || result.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("resposta vazia da OpenAI API")
	}
	return result.Choices[0].Message.Content, nil
}
