# RAG Examples

This folder contains Retrieval-Augmented Generation demos that showcase different integration patterns built on top of the `github.com/go-kratos/blades` toolkit. For a Chinese version of this guide, see `README_zh.md`.

## Layout
- `graph/`: pipelines a set of RAG nodes with `flow.Graph`.
- `middleware/`: injects retrieved context inside agent middleware.
- `shared/`: helper components (sentence chunker, in-memory store, simple reranker) used by the samples.

## Prerequisites
- Go 1.24 or newer.
- An LLM provider key supported by `contrib/openai` (set `OPENAI_API_KEY` in your environment).

## Running the examples

```bash
# Pipeline orchestrated via flow.Graph
go run ./examples/rag/graph

# Agent middleware that augments prompts on the fly
go run ./examples/rag/middleware
```

Each example logs its progress and prints the generated answer to stdout.

## Key Concepts
- **Chunking & Indexing**: `shared.SentenceChunker` splits source text while avoiding empty chunks; `shared.SimpleMemoryStore` indexes them for retrieval.
- **Retrieval & Reranking**: `SimpleMemoryStore.Retrieve` and `SimpleReranker` provide lightweight scoring suitable for demos.
- **Generation**: A `blades.Agent` wraps the selected provider to synthesize answers from the retrieved context.

Feel free to adapt the shared components or swap in real vector stores, embeddings, or rerankers for production-grade workflows.
