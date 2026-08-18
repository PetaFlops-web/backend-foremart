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
// InventoryPredictionRequest maps to POST /predict-inventory
type InventoryPredictionRequest struct {
	Store        string    `json:"store"`
	Item         string    `json:"item"`
	Date         string    `json:"date"`
	SalesHistory []float64 `json:"sales_history"`
}

// InventoryPredictionResponse maps to the ML service response
type InventoryPredictionResponse struct {
	Date           string `json:"date"`
	Store          string `json:"store"`
	Item           string `json:"item"`
	PredictedSales int    `json:"predicted_sales"`
}

// InventoryPredictionBatchRequest wraps multiple predictions for POST /predict-inventory/batch
type InventoryPredictionBatchRequest struct {
	Predictions []InventoryPredictionRequest `json:"predictions"`
}

type MLClient interface {
	TranscribeAndExtract(ctx context.Context, audioData []byte, filename string) (*MLExtractResponse, error)
	PredictInventory(ctx context.Context, req InventoryPredictionRequest) (*InventoryPredictionResponse, error)
	PredictInventoryBatch(ctx context.Context, req InventoryPredictionBatchRequest) ([]InventoryPredictionResponse, error)
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

func (c *httpMLClient) PredictInventory(ctx context.Context, req InventoryPredictionRequest) (*InventoryPredictionResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inventory prediction request: %w", err)
	}

	url := fmt.Sprintf("%s/predict-inventory", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call ML service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ML service returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result InventoryPredictionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode ML response: %w", err)
	}

	return &result, nil
}

func (c *httpMLClient) PredictInventoryBatch(ctx context.Context, req InventoryPredictionBatchRequest) ([]InventoryPredictionResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch inventory prediction request: %w", err)
	}

	url := fmt.Sprintf("%s/predict-inventory/batch", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call ML service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ML service returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var result []InventoryPredictionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode ML batch response: %w", err)
	}

	return result, nil
}
