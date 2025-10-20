package rag

import "context"

// Document represents a retrieved document or chunk with content, score, and custom metadata.
type Document struct {
	ID        string
	Content   string
	Score     float64
	Metadata  map[string]any
	Embedding []float64 // Vector representation
}

// Indexer is responsible for adding or deleting documents in the underlying storage for retrieval.
type Indexer interface {
	Add(ctx context.Context, docs []Document) error
	Delete(ctx context.Context, docIDs []string) error
}

// RetrieveOptions contains optional parameters for retrieval.
type RetrieveOptions struct {
	TopK    int
	Filters map[string]string
}

// RetrieveOption is a function type for configuring retrieval options.
type RetrieveOption func(*RetrieveOptions)

// WithTopK sets the maximum number of documents to return.
func WithTopK(topK int) RetrieveOption {
	return func(o *RetrieveOptions) {
		o.TopK = topK
	}
}

// WithFilters sets filter conditions.
func WithFilters(filters map[string]string) RetrieveOption {
	return func(o *RetrieveOptions) {
		o.Filters = filters
	}
}

// WithFilter adds a single filter condition.
func WithFilter(key, value string) RetrieveOption {
	return func(o *RetrieveOptions) {
		if o.Filters == nil {
			o.Filters = make(map[string]string)
		}
		o.Filters[key] = value
	}
}

// Retriever interface is responsible for retrieving relevant documents based on the query.
type Retriever interface {
	Retrieve(ctx context.Context, query string, opts ...RetrieveOption) ([]Document, error)
}

// Reranker interface is responsible for reordering initial retrieval results to improve relevance.
type Reranker interface {
	Rerank(ctx context.Context, query string, docs []Document) ([]Document, error)
}
