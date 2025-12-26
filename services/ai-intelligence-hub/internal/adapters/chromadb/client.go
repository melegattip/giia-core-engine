package chromadb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/pkg/errors"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/domain/entities"
)

const (
	DefaultBaseURL    = "http://localhost:8000"
	DefaultCollection = "knowledge_base"
	DefaultTimeout    = 30 * time.Second
)

// ChromaDBClient provides vector database operations using ChromaDB
type ChromaDBClient struct {
	baseURL     string
	httpClient  *http.Client
	collection  string
	logger      logger.Logger
	embeddingFn EmbeddingFunction
}

// ChromaDBConfig holds configuration for the ChromaDB client
type ChromaDBConfig struct {
	BaseURL    string
	Collection string
	Timeout    time.Duration
}

// EmbeddingFunction is a function that generates embeddings for text
type EmbeddingFunction func(ctx context.Context, text string) ([]float32, error)

// Collection represents a ChromaDB collection
type Collection struct {
	Name     string                 `json:"name"`
	ID       string                 `json:"id"`
	Metadata map[string]interface{} `json:"metadata"`
}

// AddDocumentsRequest represents a request to add documents
type AddDocumentsRequest struct {
	IDs        []string                 `json:"ids"`
	Embeddings [][]float32              `json:"embeddings"`
	Documents  []string                 `json:"documents"`
	Metadatas  []map[string]interface{} `json:"metadatas"`
}

// QueryRequest represents a query request
type QueryRequest struct {
	QueryEmbeddings [][]float32 `json:"query_embeddings"`
	NResults        int         `json:"n_results"`
	Include         []string    `json:"include"`
}

// QueryResponse represents a query response
type QueryResponse struct {
	IDs        [][]string                 `json:"ids"`
	Documents  [][]string                 `json:"documents"`
	Metadatas  [][]map[string]interface{} `json:"metadatas"`
	Distances  [][]float64                `json:"distances"`
	Embeddings [][][]float32              `json:"embeddings,omitempty"`
}

// NewChromaDBClient creates a new ChromaDB client
func NewChromaDBClient(config ChromaDBConfig, embeddingFn EmbeddingFunction, log logger.Logger) *ChromaDBClient {
	if config.BaseURL == "" {
		config.BaseURL = DefaultBaseURL
	}
	if config.Collection == "" {
		config.Collection = DefaultCollection
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultTimeout
	}

	return &ChromaDBClient{
		baseURL:    config.BaseURL,
		collection: config.Collection,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		logger:      log,
		embeddingFn: embeddingFn,
	}
}

// Initialize creates the collection if it doesn't exist
func (c *ChromaDBClient) Initialize(ctx context.Context) error {
	c.logger.Info(ctx, "Initializing ChromaDB collection", logger.Tags{
		"collection": c.collection,
		"base_url":   c.baseURL,
	})

	// Check if collection exists
	exists, err := c.collectionExists(ctx)
	if err != nil {
		return fmt.Errorf("failed to check collection: %w", err)
	}

	if !exists {
		// Create collection
		if err := c.createCollection(ctx); err != nil {
			return fmt.Errorf("failed to create collection: %w", err)
		}
		c.logger.Info(ctx, "Created ChromaDB collection", logger.Tags{
			"collection": c.collection,
		})
	} else {
		c.logger.Debug(ctx, "ChromaDB collection already exists", logger.Tags{
			"collection": c.collection,
		})
	}

	return nil
}

// collectionExists checks if the collection exists
func (c *ChromaDBClient) collectionExists(ctx context.Context) (bool, error) {
	url := fmt.Sprintf("%s/api/v1/collections/%s", c.baseURL, c.collection)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK, nil
}

// createCollection creates a new collection
func (c *ChromaDBClient) createCollection(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1/collections", c.baseURL)

	body, _ := json.Marshal(map[string]interface{}{
		"name": c.collection,
		"metadata": map[string]interface{}{
			"description": "GIIA Knowledge Base for DDMRP and supply chain documents",
			"created_at":  time.Now().Format(time.RFC3339),
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create collection: %s", string(respBody))
	}

	return nil
}

// AddDocument adds a single document to the collection
func (c *ChromaDBClient) AddDocument(ctx context.Context, doc *entities.KnowledgeDocument) error {
	// Generate embedding if not present
	if !doc.HasEmbedding() && c.embeddingFn != nil {
		embedding, err := c.embeddingFn(ctx, doc.Content)
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}
		doc.SetEmbedding(embedding)
	}

	return c.AddDocuments(ctx, []*entities.KnowledgeDocument{doc})
}

// AddDocuments adds multiple documents to the collection
func (c *ChromaDBClient) AddDocuments(ctx context.Context, docs []*entities.KnowledgeDocument) error {
	if len(docs) == 0 {
		return nil
	}

	c.logger.Debug(ctx, "Adding documents to ChromaDB", logger.Tags{
		"count": len(docs),
	})

	// Prepare request
	ids := make([]string, len(docs))
	embeddings := make([][]float32, len(docs))
	documents := make([]string, len(docs))
	metadatas := make([]map[string]interface{}, len(docs))

	for i, doc := range docs {
		// Generate embedding if not present
		if !doc.HasEmbedding() && c.embeddingFn != nil {
			embedding, err := c.embeddingFn(ctx, doc.Content)
			if err != nil {
				return fmt.Errorf("failed to generate embedding for doc %s: %w", doc.ID, err)
			}
			doc.SetEmbedding(embedding)
		}

		ids[i] = doc.ID.String()
		embeddings[i] = doc.Embedding
		documents[i] = doc.Content
		metadatas[i] = map[string]interface{}{
			"title":           doc.Title,
			"type":            string(doc.Type),
			"source":          doc.Source,
			"organization_id": doc.OrganizationID.String(),
			"chunk_index":     doc.ChunkIndex,
			"total_chunks":    doc.TotalChunks,
			"created_at":      doc.CreatedAt.Format(time.RFC3339),
		}
	}

	req := AddDocumentsRequest{
		IDs:        ids,
		Embeddings: embeddings,
		Documents:  documents,
		Metadatas:  metadatas,
	}

	return c.addToCollection(ctx, req)
}

// addToCollection sends the add request to ChromaDB
func (c *ChromaDBClient) addToCollection(ctx context.Context, req AddDocumentsRequest) error {
	url := fmt.Sprintf("%s/api/v1/collections/%s/add", c.baseURL, c.collection)

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to add documents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add documents: %s", string(respBody))
	}

	c.logger.Debug(ctx, "Successfully added documents to ChromaDB", logger.Tags{
		"count": len(req.IDs),
	})

	return nil
}

// Query performs a semantic search on the collection
func (c *ChromaDBClient) Query(ctx context.Context, query string, topK int) ([]*entities.RetrievalResult, error) {
	if c.embeddingFn == nil {
		return nil, errors.NewInternalServerError("embedding function not configured")
	}

	c.logger.Debug(ctx, "Querying ChromaDB", logger.Tags{
		"query": query,
		"top_k": topK,
	})

	// Generate query embedding
	embedding, err := c.embeddingFn(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	return c.QueryWithEmbedding(ctx, embedding, topK)
}

// QueryWithEmbedding performs a semantic search using a pre-computed embedding
func (c *ChromaDBClient) QueryWithEmbedding(ctx context.Context, embedding []float32, topK int) ([]*entities.RetrievalResult, error) {
	url := fmt.Sprintf("%s/api/v1/collections/%s/query", c.baseURL, c.collection)

	req := QueryRequest{
		QueryEmbeddings: [][]float32{embedding},
		NResults:        topK,
		Include:         []string{"documents", "metadatas", "distances"},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("query failed: %s", string(respBody))
	}

	var queryResp QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return c.parseQueryResponse(queryResp), nil
}

// parseQueryResponse converts the ChromaDB response to RetrievalResults
func (c *ChromaDBClient) parseQueryResponse(resp QueryResponse) []*entities.RetrievalResult {
	var results []*entities.RetrievalResult

	if len(resp.IDs) == 0 || len(resp.IDs[0]) == 0 {
		return results
	}

	for i, id := range resp.IDs[0] {
		docID, _ := uuid.Parse(id)

		var orgID uuid.UUID
		docType := entities.DocumentTypeDDMRPGuide
		title := ""
		source := ""

		if i < len(resp.Metadatas[0]) {
			meta := resp.Metadatas[0][i]
			if t, ok := meta["title"].(string); ok {
				title = t
			}
			if dt, ok := meta["type"].(string); ok {
				docType = entities.DocumentType(dt)
			}
			if s, ok := meta["source"].(string); ok {
				source = s
			}
			if oid, ok := meta["organization_id"].(string); ok {
				orgID, _ = uuid.Parse(oid)
			}
		}

		content := ""
		if i < len(resp.Documents[0]) {
			content = resp.Documents[0][i]
		}

		distance := 0.0
		if i < len(resp.Distances[0]) {
			distance = resp.Distances[0][i]
		}

		// Convert distance to similarity score (ChromaDB uses L2 distance by default)
		score := 1.0 / (1.0 + distance)

		doc := &entities.KnowledgeDocument{
			ID:             docID,
			OrganizationID: orgID,
			Title:          title,
			Content:        content,
			Type:           docType,
			Source:         source,
		}

		results = append(results, &entities.RetrievalResult{
			Document:       doc,
			Score:          score,
			MatchedContent: content,
		})
	}

	return results
}

// DeleteDocument deletes a document from the collection
func (c *ChromaDBClient) DeleteDocument(ctx context.Context, docID uuid.UUID) error {
	return c.DeleteDocuments(ctx, []uuid.UUID{docID})
}

// DeleteDocuments deletes multiple documents from the collection
func (c *ChromaDBClient) DeleteDocuments(ctx context.Context, docIDs []uuid.UUID) error {
	if len(docIDs) == 0 {
		return nil
	}

	url := fmt.Sprintf("%s/api/v1/collections/%s/delete", c.baseURL, c.collection)

	ids := make([]string, len(docIDs))
	for i, id := range docIDs {
		ids[i] = id.String()
	}

	body, _ := json.Marshal(map[string]interface{}{
		"ids": ids,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete documents: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed: %s", string(respBody))
	}

	c.logger.Debug(ctx, "Deleted documents from ChromaDB", logger.Tags{
		"count": len(docIDs),
	})

	return nil
}

// Count returns the number of documents in the collection
func (c *ChromaDBClient) Count(ctx context.Context) (int, error) {
	url := fmt.Sprintf("%s/api/v1/collections/%s/count", c.baseURL, c.collection)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to get count: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get count: status %d", resp.StatusCode)
	}

	var count int
	if err := json.NewDecoder(resp.Body).Decode(&count); err != nil {
		return 0, fmt.Errorf("failed to decode count: %w", err)
	}

	return count, nil
}

// HealthCheck verifies the ChromaDB connection
func (c *ChromaDBClient) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/v1/heartbeat", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ChromaDB health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ChromaDB unhealthy: status %d", resp.StatusCode)
	}

	return nil
}
