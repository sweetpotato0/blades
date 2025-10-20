# RAG (Retrieval-Augmented Generation)

`rag/` 提供 RAG 核心接口、上下文构建工具以及可复用的提示增强中间件。开发者可以在自己的业务代码里实现 `rag.Indexer`、`rag.Retriever`、`rag.Reranker` 等接口，也可以参考 `rag.AugmentationMiddleware` 与 `examples/rag/` 的演示示例。

## 开始使用

```go
// rag.Document、rag.Indexer 等接口在 rag/rag.go 中导出，可直接在应用中实现。
type MyStore struct{}

func (s *MyStore) Add(ctx context.Context, docs []rag.Document) error    { /* ... */ }
func (s *MyStore) Delete(ctx context.Context, ids []string) error         { /* ... */ }
func (s *MyStore) Retrieve(ctx context.Context, query string, opts ...rag.RetrieveOption) ([]rag.Document, error) {
    /* ... */
}
```

想要快速体验，可以运行 `examples/rag/graph` 或 `examples/rag/middleware` 下的示例：
- `graph` 展示了如何通过 `flow.Graph` 串联分块、索引、检索、重排与生成。
- `middleware` 则展示了用 `Agent` 中间件 + `PromptTemplate` 动态注入检索上下文的做法。

## 贡献指南

- 欢迎提交新的示例或第三方集成，帮助社区演示不同的 RAG 策略。
- 如需在仓库中新增通用组件，请先在 issue 中讨论需求及维护计划。
