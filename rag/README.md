# RAG (Retrieval-Augmented Generation)

The `rag/` package provides core RAG interfaces, context-building utilities, and reusable prompt augmentation middleware. Developers can implement `rag.Indexer`, `rag.Retriever`, `rag.Reranker` and other interfaces in their own business code, or refer to `rag.AugmentationMiddleware` and the demonstration examples in `examples/rag/`.

## Getting Started

```go
// rag.Document, rag.Indexer and other interfaces are exported in rag/types.go,
// and can be directly implemented in your application.
type MyStore struct{}

func (s *MyStore) Add(ctx context.Context, docs []rag.Document) error    { /* ... */ }
func (s *MyStore) Delete(ctx context.Context, ids []string) error         { /* ... */ }
func (s *MyStore) Retrieve(ctx context.Context, query string, opts ...rag.RetrieveOption) ([]rag.Document, error) {
    /* ... */
}
```

For a quick hands-on experience, run the examples in `examples/rag/graph` or `examples/rag/middleware`:
- `graph` demonstrates how to chain chunking, indexing, retrieval, reranking, and generation using `flow.Graph`.
- `middleware` shows how to use `Agent` middleware + `PromptTemplate` to dynamically inject retrieval context.

## Contributing

- Contributions of new examples or third-party integrations are welcome to help the community demonstrate different RAG strategies.
- If you need to add generic components to the repository, please discuss requirements and maintenance plans in an issue first.
