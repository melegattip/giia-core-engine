package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/adapters/chromadb"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/adapters/embeddings"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/domain/entities"
)

// RAGRetriever performs retrieval-augmented generation operations
type RAGRetriever struct {
	chromaClient     *chromadb.ChromaDBClient
	embeddingService *embeddings.Service
	simpleRetriever  providers.RAGKnowledge
	logger           logger.Logger
	useVectorStore   bool
}

// RAGConfig holds configuration for the RAG retriever
type RAGConfig struct {
	ChromaDBURL       string
	CollectionName    string
	EmbeddingProvider embeddings.EmbeddingProvider
	EmbeddingAPIKey   string
	EmbeddingModel    string
	UseVectorStore    bool
	KnowledgeBasePath string
}

// NewRAGRetriever creates a new RAG retriever
func NewRAGRetriever(config RAGConfig, log logger.Logger) (*RAGRetriever, error) {
	retriever := &RAGRetriever{
		logger:         log,
		useVectorStore: config.UseVectorStore,
	}

	if config.UseVectorStore {
		// Initialize embedding service
		embConfig := embeddings.Config{
			Provider: config.EmbeddingProvider,
			APIKey:   config.EmbeddingAPIKey,
			Model:    config.EmbeddingModel,
		}
		retriever.embeddingService = embeddings.NewService(embConfig, log)

		// Initialize ChromaDB client with embedding function
		chromaConfig := chromadb.ChromaDBConfig{
			BaseURL:    config.ChromaDBURL,
			Collection: config.CollectionName,
		}
		retriever.chromaClient = chromadb.NewChromaDBClient(
			chromaConfig,
			retriever.embeddingService.Generate,
			log,
		)
	}

	return retriever, nil
}

// SetSimpleRetriever sets the fallback simple retriever
func (r *RAGRetriever) SetSimpleRetriever(retriever providers.RAGKnowledge) {
	r.simpleRetriever = retriever
}

// Initialize initializes the RAG system
func (r *RAGRetriever) Initialize(ctx context.Context) error {
	r.logger.Info(ctx, "Initializing RAG retriever", logger.Tags{
		"use_vector_store": r.useVectorStore,
	})

	if r.useVectorStore && r.chromaClient != nil {
		if err := r.chromaClient.Initialize(ctx); err != nil {
			r.logger.Warn(ctx, "Failed to initialize ChromaDB, falling back to simple retriever", logger.Tags{
				"error": err.Error(),
			})
			r.useVectorStore = false
		}
	}

	// Initialize simple retriever as fallback
	if r.simpleRetriever != nil {
		if err := r.simpleRetriever.Initialize(ctx); err != nil {
			return fmt.Errorf("failed to initialize simple retriever: %w", err)
		}
	}

	return nil
}

// Retrieve performs semantic search to find relevant documents
func (r *RAGRetriever) Retrieve(ctx context.Context, query string, topK int) ([]string, error) {
	if topK <= 0 {
		topK = 5
	}

	r.logger.Debug(ctx, "Retrieving knowledge documents", logger.Tags{
		"query":            query,
		"top_k":            topK,
		"use_vector_store": r.useVectorStore,
	})

	// Try vector store first
	if r.useVectorStore && r.chromaClient != nil {
		results, err := r.chromaClient.Query(ctx, query, topK)
		if err == nil && len(results) > 0 {
			return r.formatResults(results), nil
		}
		if err != nil {
			r.logger.Warn(ctx, "Vector search failed, falling back to simple retriever", logger.Tags{
				"error": err.Error(),
			})
		}
	}

	// Fall back to simple retriever
	if r.simpleRetriever != nil {
		return r.simpleRetriever.Retrieve(ctx, query, topK)
	}

	return []string{}, nil
}

// RetrieveWithContext retrieves documents with additional context filtering
func (r *RAGRetriever) RetrieveWithContext(
	ctx context.Context,
	query string,
	topK int,
	organizationID uuid.UUID,
	docTypes []entities.DocumentType,
) ([]*entities.RetrievalResult, error) {
	if !r.useVectorStore || r.chromaClient == nil {
		// Fall back to simple retriever without metadata filtering
		docs, err := r.simpleRetriever.Retrieve(ctx, query, topK)
		if err != nil {
			return nil, err
		}

		results := make([]*entities.RetrievalResult, len(docs))
		for i, doc := range docs {
			results[i] = &entities.RetrievalResult{
				Document: &entities.KnowledgeDocument{
					ID:      uuid.New(),
					Content: doc,
				},
				Score:          1.0 - float64(i)*0.1,
				MatchedContent: doc,
			}
		}
		return results, nil
	}

	// Use vector store with full capability
	results, err := r.chromaClient.Query(ctx, query, topK*2) // Get more to filter
	if err != nil {
		return nil, err
	}

	// Filter by organization and document type
	var filtered []*entities.RetrievalResult
	for _, result := range results {
		// Check organization filter
		if organizationID != uuid.Nil && result.Document.OrganizationID != organizationID {
			continue
		}

		// Check document type filter
		if len(docTypes) > 0 {
			typeMatch := false
			for _, dt := range docTypes {
				if result.Document.Type == dt {
					typeMatch = true
					break
				}
			}
			if !typeMatch {
				continue
			}
		}

		filtered = append(filtered, result)
		if len(filtered) >= topK {
			break
		}
	}

	return filtered, nil
}

// IndexDocument indexes a single document into the vector store
func (r *RAGRetriever) IndexDocument(ctx context.Context, doc *entities.KnowledgeDocument) error {
	if !r.useVectorStore || r.chromaClient == nil {
		return fmt.Errorf("vector store not enabled")
	}

	r.logger.Debug(ctx, "Indexing document", logger.Tags{
		"doc_id": doc.ID.String(),
		"title":  doc.Title,
		"type":   string(doc.Type),
	})

	// Chunk the document if it's large
	if len(doc.Content) > 2000 {
		chunks := doc.ChunkDocument(1500, 200)
		for _, chunk := range chunks {
			chunkDoc := &entities.KnowledgeDocument{
				ID:             chunk.ID,
				OrganizationID: doc.OrganizationID,
				Title:          fmt.Sprintf("%s (Part %d)", doc.Title, chunk.ChunkIndex+1),
				Content:        chunk.Content,
				Type:           doc.Type,
				Source:         doc.Source,
				Metadata:       doc.Metadata,
				ChunkIndex:     chunk.ChunkIndex,
				TotalChunks:    len(chunks),
				ParentID:       &doc.ID,
				CreatedAt:      doc.CreatedAt,
				UpdatedAt:      doc.UpdatedAt,
			}

			if err := r.chromaClient.AddDocument(ctx, chunkDoc); err != nil {
				return fmt.Errorf("failed to index chunk %d: %w", chunk.ChunkIndex, err)
			}
		}
	} else {
		if err := r.chromaClient.AddDocument(ctx, doc); err != nil {
			return fmt.Errorf("failed to index document: %w", err)
		}
	}

	doc.MarkAsIndexed()

	r.logger.Info(ctx, "Document indexed successfully", logger.Tags{
		"doc_id": doc.ID.String(),
		"title":  doc.Title,
	})

	return nil
}

// IndexDocuments indexes multiple documents
func (r *RAGRetriever) IndexDocuments(ctx context.Context, docs []*entities.KnowledgeDocument) error {
	for _, doc := range docs {
		if err := r.IndexDocument(ctx, doc); err != nil {
			return err
		}
	}
	return nil
}

// DeleteDocument removes a document from the vector store
func (r *RAGRetriever) DeleteDocument(ctx context.Context, docID uuid.UUID) error {
	if !r.useVectorStore || r.chromaClient == nil {
		return fmt.Errorf("vector store not enabled")
	}

	return r.chromaClient.DeleteDocument(ctx, docID)
}

// GetDocumentCount returns the number of indexed documents
func (r *RAGRetriever) GetDocumentCount(ctx context.Context) (int, error) {
	if !r.useVectorStore || r.chromaClient == nil {
		return 0, nil
	}

	return r.chromaClient.Count(ctx)
}

// formatResults converts RetrievalResults to formatted strings
func (r *RAGRetriever) formatResults(results []*entities.RetrievalResult) []string {
	formatted := make([]string, len(results))
	for i, result := range results {
		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("## Document: %s\n", result.Document.Title))
		builder.WriteString(fmt.Sprintf("Type: %s | Relevance: %.2f\n\n", result.Document.Type, result.Score))
		builder.WriteString(result.MatchedContent)
		formatted[i] = builder.String()
	}
	return formatted
}

// BuildContextForAnalysis builds a context string for AI analysis
func (r *RAGRetriever) BuildContextForAnalysis(ctx context.Context, query string, topK int) ([]string, error) {
	results, err := r.Retrieve(ctx, query, topK)
	if err != nil {
		return nil, err
	}

	r.logger.Debug(ctx, "Built context for analysis", logger.Tags{
		"query":         query,
		"results_count": len(results),
	})

	return results, nil
}

// HealthCheck verifies the RAG system is operational
func (r *RAGRetriever) HealthCheck(ctx context.Context) error {
	if r.useVectorStore && r.chromaClient != nil {
		if err := r.chromaClient.HealthCheck(ctx); err != nil {
			return fmt.Errorf("ChromaDB health check failed: %w", err)
		}
	}

	if r.embeddingService != nil {
		if err := r.embeddingService.HealthCheck(ctx); err != nil {
			return fmt.Errorf("embedding service health check failed: %w", err)
		}
	}

	return nil
}

// Ensure RAGRetriever implements RAGKnowledge interface
var _ providers.RAGKnowledge = (*RAGRetriever)(nil)
