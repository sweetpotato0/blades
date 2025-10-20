package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-kratos/blades"
	"github.com/go-kratos/blades/contrib/openai"
	"github.com/go-kratos/blades/examples/rag/shared"
	"github.com/go-kratos/blades/flow"
	"github.com/go-kratos/blades/rag"
)

func main() {
	ctx := context.Background()

	// 1. Create store and components (shared example implementations are in the shared package)
	store := shared.NewSimpleMemoryStore()

	// 2. Create custom scoring function for reranking
	scorer := func(ctx context.Context, query string, doc rag.Document) (float64, error) {
		// Simple example: score based on content length
		return float64(len(doc.Content)) / 100.0, nil
	}
	reranker := shared.NewSimpleReranker(scorer)

	// 3. Create LLM Agent (using OpenAI)
	provider := openai.NewChatProvider()
	agent := blades.NewAgent(
		"rag-graph-assistant",
		blades.WithProvider(provider),
		blades.WithModel("gpt-4o-mini"),
	)

	// 4. Create nodes
	chunkingNode := NewChunkingNode()
	indexingNode := NewIndexingNode(store)
	retrievalNode := NewRetrievalNode(store)
	rerankingNode := NewRerankingNode(reranker)
	generationNode := NewGenerationNode(agent)

	// 5. Define state transition handler (passes RAGState between nodes)
	transitionHandler := func(ctx context.Context, transition flow.Transition, output *RAGState) (*RAGState, error) {
		log.Printf("[Transition] %s -> %s\n", transition.From, transition.To)
		return output, nil
	}

	// 6. Build Graph
	g := flow.NewGraph[*RAGState, *RAGState, blades.ModelOption]("rag-graph-pipeline", transitionHandler)

	// Add nodes
	g.AddNode(chunkingNode)
	g.AddNode(indexingNode)
	g.AddNode(retrievalNode)
	g.AddNode(rerankingNode)
	g.AddNode(generationNode)

	// Add edges: chunking -> indexing -> retrieval -> reranking -> generation
	g.AddStart(chunkingNode)
	g.AddEdge(chunkingNode, indexingNode)
	g.AddEdge(indexingNode, retrievalNode)
	g.AddEdge(retrievalNode, rerankingNode)
	g.AddEdge(rerankingNode, generationNode)

	// 7. Compile Graph
	runner, err := g.Compile()
	if err != nil {
		log.Fatalf("Failed to compile graph: %v", err)
	}

	// 8. Prepare initial state
	longDoc := `Rainy weather requires special preparation for your commute.
	First, always carry a waterproof jacket in your bag, as weather can change unexpectedly.
	Second, check the forecast before leaving home to plan your route accordingly.
	Third, a compact umbrella fits in most bags and provides essential protection.
	Fourth, choose covered walkways or sheltered routes when possible to minimize exposure.
	Finally, leave earlier than usual to account for slower traffic during heavy rain.`

	initialState := &RAGState{
		Query:       "How do I prepare for a rainy commute?",
		OriginalDoc: longDoc,
	}

	// 9. Run Graph
	fmt.Println("=== Starting RAG Graph Pipeline ===")
	fmt.Printf("Question: %s\n\n", initialState.Query)

	finalState, err := runner.Run(ctx, initialState)
	if err != nil {
		log.Fatalf("Pipeline execution failed: %v", err)
	}

	// 10. Output final result
	fmt.Println("\n=== Final Answer ===")
	fmt.Println(finalState.FinalAnswer)
}
