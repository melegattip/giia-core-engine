package entities

import (
	"time"

	"github.com/google/uuid"
)

// DocumentType represents the type of knowledge document
type DocumentType string

const (
	DocumentTypeDDMRPGuide     DocumentType = "ddmrp_guide"
	DocumentTypeBestPractice   DocumentType = "best_practice"
	DocumentTypeCaseStudy      DocumentType = "case_study"
	DocumentTypePolicy         DocumentType = "policy"
	DocumentTypeProcedure      DocumentType = "procedure"
	DocumentTypeSupplierInfo   DocumentType = "supplier_info"
	DocumentTypeProductInfo    DocumentType = "product_info"
	DocumentTypeHistoricalData DocumentType = "historical_data"
)

// KnowledgeDocument represents a document in the knowledge base
type KnowledgeDocument struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	Title          string
	Content        string
	Type           DocumentType
	Source         string
	Metadata       DocumentMetadata
	Embedding      []float32
	ChunkIndex     int
	TotalChunks    int
	ParentID       *uuid.UUID // For chunked documents, reference to parent
	CreatedAt      time.Time
	UpdatedAt      time.Time
	IndexedAt      *time.Time
}

// DocumentMetadata contains additional information about a document
type DocumentMetadata struct {
	Author      string            `json:"author,omitempty"`
	Version     string            `json:"version,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Language    string            `json:"language,omitempty"`
	Category    string            `json:"category,omitempty"`
	Subcategory string            `json:"subcategory,omitempty"`
	Keywords    []string          `json:"keywords,omitempty"`
	Custom      map[string]string `json:"custom,omitempty"`
}

// EmbeddingStatus represents the status of document embedding
type EmbeddingStatus string

const (
	EmbeddingStatusPending    EmbeddingStatus = "pending"
	EmbeddingStatusProcessing EmbeddingStatus = "processing"
	EmbeddingStatusCompleted  EmbeddingStatus = "completed"
	EmbeddingStatusFailed     EmbeddingStatus = "failed"
)

// DocumentChunk represents a chunk of a larger document for RAG
type DocumentChunk struct {
	ID         uuid.UUID
	DocumentID uuid.UUID
	Content    string
	ChunkIndex int
	Embedding  []float32
	StartChar  int
	EndChar    int
	Overlap    int
}

// RetrievalResult represents a document retrieved from semantic search
type RetrievalResult struct {
	Document       *KnowledgeDocument
	Score          float64
	MatchedContent string
	Highlights     []string
}

// NewKnowledgeDocument creates a new knowledge document
func NewKnowledgeDocument(
	organizationID uuid.UUID,
	title string,
	content string,
	docType DocumentType,
	source string,
) *KnowledgeDocument {
	now := time.Now()
	return &KnowledgeDocument{
		ID:             uuid.New(),
		OrganizationID: organizationID,
		Title:          title,
		Content:        content,
		Type:           docType,
		Source:         source,
		Metadata:       DocumentMetadata{Tags: []string{}, Keywords: []string{}},
		CreatedAt:      now,
		UpdatedAt:      now,
		ChunkIndex:     0,
		TotalChunks:    1,
	}
}

// MarkAsIndexed marks the document as indexed in the vector store
func (d *KnowledgeDocument) MarkAsIndexed() {
	now := time.Now()
	d.IndexedAt = &now
}

// SetEmbedding sets the embedding vector for the document
func (d *KnowledgeDocument) SetEmbedding(embedding []float32) {
	d.Embedding = embedding
}

// IsIndexed returns true if the document has been indexed
func (d *KnowledgeDocument) IsIndexed() bool {
	return d.IndexedAt != nil
}

// HasEmbedding returns true if the document has an embedding
func (d *KnowledgeDocument) HasEmbedding() bool {
	return len(d.Embedding) > 0
}

// ChunkDocument splits a document into smaller chunks for better retrieval
func (d *KnowledgeDocument) ChunkDocument(chunkSize, overlap int) []*DocumentChunk {
	if len(d.Content) <= chunkSize {
		return []*DocumentChunk{
			{
				ID:         uuid.New(),
				DocumentID: d.ID,
				Content:    d.Content,
				ChunkIndex: 0,
				StartChar:  0,
				EndChar:    len(d.Content),
				Overlap:    0,
			},
		}
	}

	var chunks []*DocumentChunk
	start := 0
	chunkIndex := 0

	for start < len(d.Content) {
		end := start + chunkSize
		if end > len(d.Content) {
			end = len(d.Content)
		}

		// Find a good break point (end of sentence or paragraph)
		if end < len(d.Content) {
			breakPoints := []byte{'.', '\n', '!', '?'}
			for i := end - 1; i > start+chunkSize/2; i-- {
				for _, bp := range breakPoints {
					if d.Content[i] == bp {
						end = i + 1
						break
					}
				}
				if end < start+chunkSize {
					break
				}
			}
		}

		chunks = append(chunks, &DocumentChunk{
			ID:         uuid.New(),
			DocumentID: d.ID,
			Content:    d.Content[start:end],
			ChunkIndex: chunkIndex,
			StartChar:  start,
			EndChar:    end,
			Overlap:    overlap,
		})

		start = end - overlap
		if start < 0 {
			start = 0
		}
		chunkIndex++

		// Prevent infinite loop
		if end >= len(d.Content) {
			break
		}
	}

	d.TotalChunks = len(chunks)
	return chunks
}
