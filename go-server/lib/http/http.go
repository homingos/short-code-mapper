package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type Client struct {
	lgr *zap.SugaredLogger
	cli http.Client
}

func NewClient(lgr *zap.SugaredLogger) *Client {
	return &Client{lgr, http.Client{}}
}

func (client *Client) DoPost(url string, data interface{}, headers map[string]string) ([]byte, error) {
	payload := &bytes.Buffer{}
	enc := json.NewEncoder(payload)
	enc.SetEscapeHTML(false)
	err := enc.Encode(data)
	// payload, err := json.Marshal(data)
	// fmt.Println("payload: ", payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := client.cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		var responseMap map[string]interface{}
		if err := json.Unmarshal(body, &responseMap); err != nil {
			return body, fmt.Errorf("failed to consume credit")
		}

		if message, ok := responseMap["message"].(string); ok {
			return body, fmt.Errorf(message)
		}

		return body, fmt.Errorf("response: %s", string(body))

	}
	return body, nil
}

func (client *Client) DoFdbPost(url string, data interface{}, headers map[string]string) ([]byte, error) {
	payload := &bytes.Buffer{}
	enc := json.NewEncoder(payload)
	enc.SetEscapeHTML(false)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := client.cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		var responseMap map[string]interface{}
		if message, ok := responseMap["message"].(string); ok {
			return body, fmt.Errorf(message)
		}
		return body, fmt.Errorf("response: %s", string(body))

	}
	return body, nil
}
