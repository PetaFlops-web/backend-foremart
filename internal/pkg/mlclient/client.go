package mlclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type ExtractedItem struct {
	Item          string  `json:"item"`
	Qty           float64 `json:"qty"`
	Harga         int64   `json:"harga"`
	SumberHarga   string  `json:"sumber_harga"`
	ProdukKatalog string  `json:"produk_katalog"`
	SkorCocok     float64 `json:"skor_cocok"`
	StatusCocok   string  `json:"status_cocok"`
}

type MLExtractResponse struct {
	SumberTranskrip string          `json:"sumber_transkrip"`
	RawText         string          `json:"raw_text"`
	Items           []ExtractedItem `json:"items"`
}

type MLClient interface {
	TranscribeAndExtract(ctx context.Context, audioData []byte, filename string) (*MLExtractResponse, error)
}
type httpMLClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewMLClient(baseURL string) MLClient {
	return &httpMLClient{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (c *httpMLClient) TranscribeAndExtract(ctx context.Context, audioData []byte, filename string) (*MLExtractResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := part.Write(audioData); err != nil {
		return nil, fmt.Errorf("failed to write audio data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	url := fmt.Sprintf("%s/transcribe", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call ML service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ML service returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result MLExtractResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode ML response: %w", err)
	}

	return &result, nil
}
