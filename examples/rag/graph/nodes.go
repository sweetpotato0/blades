package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-kratos/blades"
	"github.com/go-kratos/blades/examples/rag/shared"
	"github.com/go-kratos/blades/rag"
)

// RAGState represents the state passed between nodes
type RAGState struct {
	Query        string
	OriginalDoc  string
	Chunks       []string
	Documents    []rag.Document
	RerankedDocs []rag.Document
	FinalAnswer  string
}

// ChunkingNode is responsible for splitting long documents into chunks
type ChunkingNode struct {
	chunker *shared.SentenceChunker
}

func NewChunkingNode() *ChunkingNode {
	return &ChunkingNode{
		chunker: shared.NewSentenceChunker(150),
	}
}

func (n *ChunkingNode) Name() string {
	return "chunking"
}

func (n *ChunkingNode) Run(ctx context.Context, state *RAGState, opts ...blades.ModelOption) (*RAGState, error) {
	log.Println("[Chunking] Splitting document into chunks...")
	state.Chunks = n.chunker.Split(state.OriginalDoc)
	log.Printf("[Chunking] Created %d chunks\n", len(state.Chunks))
	return state, nil
}

func (n *ChunkingNode) RunStream(ctx context.Context, state *RAGState, opts ...blades.ModelOption) (blades.Streamable[*RAGState], error) {
	result, err := n.Run(ctx, state, opts...)
	if err != nil {
		return nil, err
	}
	pipe := blades.NewStreamPipe[*RAGState]()
	pipe.Send(result)
	pipe.Close()
	return pipe, nil
}

// IndexingNode is responsible for indexing document chunks into the store
type IndexingNode struct {
	store rag.Indexer
}

func NewIndexingNode(store rag.Indexer) *IndexingNode {
	return &IndexingNode{store: store}
}

func (n *IndexingNode) Name() string {
	return "indexing"
}

func (n *IndexingNode) Run(ctx context.Context, state *RAGState, opts ...blades.ModelOption) (*RAGState, error) {
	log.Println("[Indexing] Adding chunks to document store...")

	docs := make([]rag.Document, len(state.Chunks))
	for i, chunk := range state.Chunks {
		docs[i] = rag.Document{
			ID:       fmt.Sprintf("doc-%d", i),
			Content:  chunk,
			Metadata: map[string]any{"source": "commute_guide", "chunk": i},
		}
	}

	if err := n.store.Add(ctx, docs); err != nil {
		return nil, fmt.Errorf("indexing failed: %w", err)
	}

	log.Printf("[Indexing] Indexed %d documents\n", len(docs))
	return state, nil
}

func (n *IndexingNode) RunStream(ctx context.Context, state *RAGState, opts ...blades.ModelOption) (blades.Streamable[*RAGState], error) {
	result, err := n.Run(ctx, state, opts...)
	if err != nil {
		return nil, err
	}
	pipe := blades.NewStreamPipe[*RAGState]()
	pipe.Send(result)
	pipe.Close()
	return pipe, nil
}

// RetrievalNode is responsible for retrieving relevant documents
type RetrievalNode struct {
	retriever rag.Retriever
}

func NewRetrievalNode(retriever rag.Retriever) *RetrievalNode {
	return &RetrievalNode{retriever: retriever}
}

func (n *RetrievalNode) Name() string {
	return "retrieval"
}

func (n *RetrievalNode) Run(ctx context.Context, state *RAGState, opts ...blades.ModelOption) (*RAGState, error) {
	log.Printf("[Retrieval] Searching for: %s\n", state.Query)

	docs, err := n.retriever.Retrieve(ctx, state.Query, rag.WithTopK(3))
	if err != nil {
		return nil, fmt.Errorf("retrieval failed: %w", err)
	}

	state.Documents = docs
	log.Printf("[Retrieval] Found %d documents\n", len(docs))
	for i, doc := range docs {
		log.Printf("  %d. [Score: %.3f] %s\n", i+1, doc.Score, doc.Content)
	}

	return state, nil
}

func (n *RetrievalNode) RunStream(ctx context.Context, state *RAGState, opts ...blades.ModelOption) (blades.Streamable[*RAGState], error) {
	result, err := n.Run(ctx, state, opts...)
	if err != nil {
		return nil, err
	}
	pipe := blades.NewStreamPipe[*RAGState]()
	pipe.Send(result)
	pipe.Close()
	return pipe, nil
}

// RerankingNode is responsible for reordering retrieval results
type RerankingNode struct {
	reranker rag.Reranker
}

func NewRerankingNode(reranker rag.Reranker) *RerankingNode {
	return &RerankingNode{reranker: reranker}
}

func (n *RerankingNode) Name() string {
	return "reranking"
}

func (n *RerankingNode) Run(ctx context.Context, state *RAGState, opts ...blades.ModelOption) (*RAGState, error) {
	log.Println("[Reranking] Reordering documents...")

	reranked, err := n.reranker.Rerank(ctx, state.Query, state.Documents)
	if err != nil {
		return nil, fmt.Errorf("reranking failed: %w", err)
	}

	state.RerankedDocs = reranked
	log.Printf("[Reranking] Top %d documents after reranking:\n", len(reranked))
	for i, doc := range reranked {
		log.Printf("  %d. [Score: %.3f] %s\n", i+1, doc.Score, doc.Content)
	}

	return state, nil
}

func (n *RerankingNode) RunStream(ctx context.Context, state *RAGState, opts ...blades.ModelOption) (blades.Streamable[*RAGState], error) {
	result, err := n.Run(ctx, state, opts...)
	if err != nil {
		return nil, err
	}
	pipe := blades.NewStreamPipe[*RAGState]()
	pipe.Send(result)
	pipe.Close()
	return pipe, nil
}

// GenerationNode is responsible for generating answers using LLM
type GenerationNode struct {
	agent *blades.Agent
}

func NewGenerationNode(agent *blades.Agent) *GenerationNode {
	return &GenerationNode{agent: agent}
}

func (n *GenerationNode) Name() string {
	return "generation"
}

func (n *GenerationNode) Run(ctx context.Context, state *RAGState, opts ...blades.ModelOption) (*RAGState, error) {
	log.Println("[Generation] Generating answer with LLM...")

	// Build context
	contextText := rag.BuildContext(state.RerankedDocs)

	// Use system message to provide context, user message contains only the question
	prompt := &blades.Prompt{
		Messages: []*blades.Message{
			blades.SystemMessage(fmt.Sprintf("You are a helpful assistant. Use the following context to answer the user's question.\n\nContext:\n%s", contextText)),
			blades.UserMessage(state.Query),
		},
	}

	response, err := n.agent.Run(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	state.FinalAnswer = response.Text()
	log.Println("[Generation] Answer generated successfully")

	return state, nil
}

func (n *GenerationNode) RunStream(ctx context.Context, state *RAGState, opts ...blades.ModelOption) (blades.Streamable[*RAGState], error) {
	result, err := n.Run(ctx, state, opts...)
	if err != nil {
		return nil, err
	}
	pipe := blades.NewStreamPipe[*RAGState]()
	pipe.Send(result)
	pipe.Close()
	return pipe, nil
}
