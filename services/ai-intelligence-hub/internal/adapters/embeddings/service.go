package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/melegattip/giia-core-engine/pkg/logger"
)

// EmbeddingProvider represents different embedding providers
type EmbeddingProvider string

const (
	ProviderOpenAI   EmbeddingProvider = "openai"
	ProviderVoyageAI EmbeddingProvider = "voyage"
	ProviderLocal    EmbeddingProvider = "local"
)

// Service provides text embedding generation capabilities
type Service struct {
	provider   EmbeddingProvider
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
	logger     logger.Logger
	dimension  int
}

// Config holds configuration for the embedding service
type Config struct {
	Provider  EmbeddingProvider
	APIKey    string
	Model     string
	BaseURL   string
	Timeout   time.Duration
	Dimension int
}

// OpenAIEmbeddingRequest represents an OpenAI embedding request
type OpenAIEmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// OpenAIEmbeddingResponse represents an OpenAI embedding response
type OpenAIEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Index     int       `json:"index"`
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// NewService creates a new embedding service
func NewService(config Config, log logger.Logger) *Service {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.Dimension == 0 {
		config.Dimension = 1536 // Default for OpenAI text-embedding-3-small
	}

	// Set defaults based on provider
	switch config.Provider {
	case ProviderOpenAI:
		if config.Model == "" {
			config.Model = "text-embedding-3-small"
		}
		if config.BaseURL == "" {
			config.BaseURL = "https://api.openai.com/v1/embeddings"
		}
	case ProviderVoyageAI:
		if config.Model == "" {
			config.Model = "voyage-2"
		}
		if config.BaseURL == "" {
			config.BaseURL = "https://api.voyageai.com/v1/embeddings"
		}
	case ProviderLocal:
		if config.BaseURL == "" {
			config.BaseURL = "http://localhost:11434/api/embeddings" // Ollama default
		}
		if config.Model == "" {
			config.Model = "nomic-embed-text"
		}
	}

	return &Service{
		provider:   config.Provider,
		apiKey:     config.APIKey,
		model:      config.Model,
		baseURL:    config.BaseURL,
		dimension:  config.Dimension,
		httpClient: &http.Client{Timeout: config.Timeout},
		logger:     log,
	}
}

// Generate creates an embedding for the given text
func (s *Service) Generate(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	s.logger.Debug(ctx, "Generating embedding", logger.Tags{
		"provider":    string(s.provider),
		"model":       s.model,
		"text_length": len(text),
	})

	switch s.provider {
	case ProviderOpenAI:
		return s.generateOpenAI(ctx, text)
	case ProviderVoyageAI:
		return s.generateVoyageAI(ctx, text)
	case ProviderLocal:
		return s.generateLocal(ctx, text)
	default:
		// Fallback to local hash-based embedding for development
		return s.generateHashEmbedding(text), nil
	}
}

// GenerateBatch creates embeddings for multiple texts
func (s *Service) GenerateBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	s.logger.Debug(ctx, "Generating batch embeddings", logger.Tags{
		"provider": string(s.provider),
		"count":    len(texts),
	})

	switch s.provider {
	case ProviderOpenAI:
		return s.generateOpenAIBatch(ctx, texts)
	default:
		// Fall back to sequential generation
		embeddings := make([][]float32, len(texts))
		for i, text := range texts {
			emb, err := s.Generate(ctx, text)
			if err != nil {
				return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
			}
			embeddings[i] = emb
		}
		return embeddings, nil
	}
}

// generateOpenAI generates embeddings using OpenAI API
func (s *Service) generateOpenAI(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := s.generateOpenAIBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return embeddings[0], nil
}

// generateOpenAIBatch generates embeddings for multiple texts using OpenAI API
func (s *Service) generateOpenAIBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if s.apiKey == "" {
		s.logger.Warn(ctx, "OpenAI API key not configured, using hash embedding", nil)
		embeddings := make([][]float32, len(texts))
		for i, text := range texts {
			embeddings[i] = s.generateHashEmbedding(text)
		}
		return embeddings, nil
	}

	req := OpenAIEmbeddingRequest{
		Model: s.model,
		Input: texts,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var embResp OpenAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	embeddings := make([][]float32, len(embResp.Data))
	for _, data := range embResp.Data {
		embeddings[data.Index] = data.Embedding
	}

	s.logger.Debug(ctx, "Generated OpenAI embeddings", logger.Tags{
		"count":       len(embeddings),
		"tokens_used": embResp.Usage.TotalTokens,
	})

	return embeddings, nil
}

// generateVoyageAI generates embeddings using Voyage AI
func (s *Service) generateVoyageAI(ctx context.Context, text string) ([]float32, error) {
	if s.apiKey == "" {
		s.logger.Warn(ctx, "Voyage AI API key not configured, using hash embedding", nil)
		return s.generateHashEmbedding(text), nil
	}

	req := map[string]interface{}{
		"model": s.model,
		"input": []string{text},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Voyage AI: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Voyage AI error (%d): %s", resp.StatusCode, string(respBody))
	}

	var voyageResp struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&voyageResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(voyageResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return voyageResp.Data[0].Embedding, nil
}

// generateLocal generates embeddings using a local model (Ollama)
func (s *Service) generateLocal(ctx context.Context, text string) ([]float32, error) {
	req := map[string]interface{}{
		"model":  s.model,
		"prompt": text,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		// Fall back to hash embedding if local server is not available
		s.logger.Warn(ctx, "Local embedding server unavailable, using hash embedding", logger.Tags{
			"error": err.Error(),
		})
		return s.generateHashEmbedding(text), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.logger.Warn(ctx, "Local embedding failed, using hash embedding", nil)
		return s.generateHashEmbedding(text), nil
	}

	var localResp struct {
		Embedding []float32 `json:"embedding"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&localResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return localResp.Embedding, nil
}

// generateHashEmbedding creates a deterministic embedding based on text hashing
// This is a fallback for development/testing when no embedding API is available
func (s *Service) generateHashEmbedding(text string) []float32 {
	embedding := make([]float32, s.dimension)

	// Normalize text
	text = strings.ToLower(strings.TrimSpace(text))
	words := strings.Fields(text)

	// Create a simple bag-of-words style embedding
	for i, word := range words {
		// Hash each word to a position in the embedding
		hash := hashString(word)
		pos := int(hash % uint64(s.dimension))

		// Add weighted contribution
		weight := 1.0 / float32(i+1) // Earlier words have higher weight
		embedding[pos] += weight

		// Also affect neighboring positions for smoothing
		if pos > 0 {
			embedding[pos-1] += weight * 0.3
		}
		if pos < s.dimension-1 {
			embedding[pos+1] += weight * 0.3
		}
	}

	// Normalize the embedding
	return normalizeEmbedding(embedding)
}

// hashString creates a simple hash of a string
func hashString(s string) uint64 {
	var hash uint64 = 5381
	for _, c := range s {
		hash = ((hash << 5) + hash) + uint64(c)
	}
	return hash
}

// normalizeEmbedding normalizes an embedding to unit length
func normalizeEmbedding(embedding []float32) []float32 {
	var sum float64
	for _, v := range embedding {
		sum += float64(v * v)
	}

	if sum == 0 {
		return embedding
	}

	norm := float32(math.Sqrt(sum))
	result := make([]float32, len(embedding))
	for i, v := range embedding {
		result[i] = v / norm
	}

	return result
}

// CosineSimilarity calculates the cosine similarity between two embeddings
func CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// GetDimension returns the embedding dimension
func (s *Service) GetDimension() int {
	return s.dimension
}

// GetProvider returns the embedding provider
func (s *Service) GetProvider() EmbeddingProvider {
	return s.provider
}

// HealthCheck verifies the embedding service is operational
func (s *Service) HealthCheck(ctx context.Context) error {
	_, err := s.Generate(ctx, "health check")
	return err
}
