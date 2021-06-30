package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	return &Client{
		httpClient: httpClient,
	}
}

func (c *Client) SendMessageToWebhook(ctx context.Context, message, webhookURL string) error {
	if c.httpClient == nil {
		return fmt.Errorf("httpClient is not defined")
	}

	postBody, err := json.Marshal(map[string]string{
		"text": message,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal body: %v", err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(postBody))
	if err != nil {
		return fmt.Errorf("failed to create http request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request to slack webhook URL: %v", err)
	}
	defer resp.Body.Close()

	return nil
}
