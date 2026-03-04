package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jamesphm04/splose-clone-be/internal/models/dtos/open_ai_client"
	"go.uber.org/zap"
)

type SploseCloneAIClient struct {
	apiKey  string
	baseURL string
	log     *zap.Logger
}

func NewSploseCloneAIClient(apiKey string, baseURL string, log *zap.Logger) *SploseCloneAIClient {
	return &SploseCloneAIClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		log:     log,
	}
}

func (c *SploseCloneAIClient) getHeaders() http.Header {
	return http.Header{
		"X-API-Key":    []string{c.apiKey},
		"Content-Type": []string{"application/json"},
	}
}

func (c *SploseCloneAIClient) SendMessage(ctx context.Context, request open_ai_client.SendMessageRequest) (string, error) {
	headers := c.getHeaders()
	body, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/conversations/send-message", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header = headers

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to send message: %s", resp.Status)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response open_ai_client.SendMessageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	return response.Message, nil
}
