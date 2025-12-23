package rag

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/melegattip/giia-core-engine/pkg/logger"
	"github.com/melegattip/giia-core-engine/services/ai-intelligence-hub/internal/core/providers"
)

type SimpleKnowledgeRetriever struct {
	knowledgeBasePath string
	documents         map[string]string
	logger            logger.Logger
}

func NewSimpleKnowledgeRetriever(knowledgeBasePath string, logger logger.Logger) providers.RAGKnowledge {
	return &SimpleKnowledgeRetriever{
		knowledgeBasePath: knowledgeBasePath,
		documents:         make(map[string]string),
		logger:            logger,
	}
}

func (r *SimpleKnowledgeRetriever) Initialize(ctx context.Context) error {
	r.logger.Info(ctx, "Initializing knowledge base", logger.Tags{
		"path": r.knowledgeBasePath,
	})

	if _, err := os.Stat(r.knowledgeBasePath); os.IsNotExist(err) {
		r.logger.Warn(ctx, "Knowledge base path does not exist, using empty knowledge base", logger.Tags{
			"path": r.knowledgeBasePath,
		})
		return nil
	}

	err := filepath.Walk(r.knowledgeBasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".md") {
			content, err := os.ReadFile(path)
			if err != nil {
				r.logger.Warn(ctx, "Failed to read knowledge file", logger.Tags{
					"path": path,
					"error": err.Error(),
				})
				return nil
			}

			relPath, _ := filepath.Rel(r.knowledgeBasePath, path)
			r.documents[relPath] = string(content)

			r.logger.Debug(ctx, "Loaded knowledge document", logger.Tags{
				"document": relPath,
				"size":     len(content),
			})
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to load knowledge base: %w", err)
	}

	r.logger.Info(ctx, "Knowledge base initialized", logger.Tags{
		"documents_loaded": len(r.documents),
	})

	return nil
}

func (r *SimpleKnowledgeRetriever) Retrieve(ctx context.Context, query string, topK int) ([]string, error) {
	if len(r.documents) == 0 {
		r.logger.Warn(ctx, "No knowledge documents available", nil)
		return []string{}, nil
	}

	query = strings.ToLower(query)
	keywords := extractKeywords(query)

	type scoredDoc struct {
		path    string
		content string
		score   int
	}

	var scored []scoredDoc

	for path, content := range r.documents {
		score := 0
		contentLower := strings.ToLower(content)

		for _, keyword := range keywords {
			score += strings.Count(contentLower, keyword)
		}

		if score > 0 {
			scored = append(scored, scoredDoc{
				path:    path,
				content: content,
				score:   score,
			})
		}
	}

	for i := 0; i < len(scored); i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	var results []string
	limit := topK
	if limit > len(scored) {
		limit = len(scored)
	}

	for i := 0; i < limit; i++ {
		snippet := r.extractRelevantSnippet(scored[i].content, keywords, 1000)
		results = append(results, fmt.Sprintf("Document: %s\n\n%s", scored[i].path, snippet))
	}

	r.logger.Debug(ctx, "Retrieved knowledge documents", logger.Tags{
		"query":         query,
		"results_count": len(results),
	})

	return results, nil
}

func extractKeywords(query string) []string {
	stopWords := map[string]bool{
		"a": true, "an": true, "and": true, "are": true, "as": true, "at": true,
		"be": true, "by": true, "for": true, "from": true, "has": true, "he": true,
		"in": true, "is": true, "it": true, "its": true, "of": true, "on": true,
		"that": true, "the": true, "to": true, "was": true, "will": true, "with": true,
	}

	words := strings.Fields(query)
	var keywords []string

	for _, word := range words {
		word = strings.Trim(word, ".,!?;:")
		if len(word) > 2 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

func (r *SimpleKnowledgeRetriever) extractRelevantSnippet(content string, keywords []string, maxLength int) string {
	lines := strings.Split(content, "\n")

	type scoredLine struct {
		line  string
		score int
		index int
	}

	var scored []scoredLine

	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		score := 0
		lineLower := strings.ToLower(line)
		for _, keyword := range keywords {
			if strings.Contains(lineLower, keyword) {
				score += 10
			}
		}

		if score > 0 || (i > 0 && len(scored) > 0 && scored[len(scored)-1].index == i-1) {
			scored = append(scored, scoredLine{
				line:  line,
				score: score,
				index: i,
			})
		}
	}

	var snippet strings.Builder
	currentLength := 0

	for _, sl := range scored {
		if currentLength+len(sl.line) > maxLength {
			break
		}
		snippet.WriteString(sl.line)
		snippet.WriteString("\n")
		currentLength += len(sl.line) + 1
	}

	result := snippet.String()
	if len(result) == 0 && len(content) > 0 {
		if len(content) > maxLength {
			return content[:maxLength] + "..."
		}
		return content
	}

	return result
}
