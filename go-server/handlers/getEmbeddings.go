package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/homingos/flam-go-common/errors"
	"github.com/homingos/campaign-svc/config"
	"github.com/homingos/campaign-svc/dtos"
)

func GetEmbeddings(text string) ([]float32, error) {
	fmt.Println("getting embedding")
	apiURL := config.LoadConfig().EmbeddingModel.URL
	req, err := http.NewRequest("POST", apiURL, nil)
	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", config.LoadConfig().EmbeddingModel.APIKey)
	payload := map[string]string{
		"text": text,
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}
	var response *dtos.EmbeddingResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.InternalServerError(err.Error())
	}
	if response == nil {
		return nil, errors.InternalServerError("response is nil")
	}
	if response.Embedding == nil {
		return nil, errors.InternalServerError("embedding is nil")
	}
	return response.Embedding, nil
}