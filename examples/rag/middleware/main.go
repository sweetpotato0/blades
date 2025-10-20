package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-kratos/blades"
	"github.com/go-kratos/blades/contrib/openai"
	"github.com/go-kratos/blades/examples/rag/shared"
	"github.com/go-kratos/blades/rag"
)

func main() {
	ctx := context.Background()

	// 1. 构建示例文档索引
	store := shared.NewSimpleMemoryStore()
	chunker := shared.NewSentenceChunker(150)

	longDoc := `Rainy weather requires special preparation for your commute.
	First, always carry a waterproof jacket in your bag, as weather can change unexpectedly.
	Second, check the forecast before leaving home to plan your route accordingly.
	Third, a compact umbrella fits in most bags and provides essential protection.
	Fourth, choose covered walkways or sheltered routes when possible to minimize exposure.
	Finally, leave earlier than usual to account for slower traffic during heavy rain.`

	chunks := chunker.Split(longDoc)
	documents := make([]rag.Document, len(chunks))
	for i, chunk := range chunks {
		documents[i] = rag.Document{
			ID:       fmt.Sprintf("doc-%d", i),
			Content:  chunk,
			Metadata: map[string]any{"source": "commute_guide", "chunk": i},
		}
	}

	if err := store.Add(ctx, documents); err != nil {
		log.Fatalf("failed to index documents: %v", err)
	}

	// 2. 创建带有中间件的 Agent
	provider := openai.NewChatProvider()
	systemTemplate := "You are a helpful assistant. Use the context below to answer the question.\n\nContext:\n{{.Context}}"
	userTemplate := "Question: {{.Question}}"

	agent := blades.NewAgent(
		"rag-middleware-assistant",
		blades.WithProvider(provider),
		blades.WithModel("gpt-4o-mini"),
		blades.WithMiddleware(rag.AugmentationMiddleware(store, systemTemplate, userTemplate)),
	)

	// 3. 使用 PromptTemplate 构建初始问题（中间件会在运行时注入上下文）
	params := map[string]any{
		"Question": "How do I prepare for a rainy commute?",
	}

	prompt, err := blades.NewPromptTemplate().
		User("{{.Question}}", params).
		Build()
	if err != nil {
		log.Fatalf("failed to build prompt: %v", err)
	}

	fmt.Println("=== Running RAG Middleware Example ===")

	response, err := agent.Run(ctx, prompt)
	if err != nil {
		log.Fatalf("agent run failed: %v", err)
	}

	fmt.Println("\n=== Final Answer ===")
	fmt.Println(response.Text())
}
