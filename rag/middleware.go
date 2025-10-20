package rag

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/go-kratos/blades"
)

// BuildContext converts documents into a numbered context block.
func BuildContext(docs []Document) string {
	if len(docs) == 0 {
		return ""
	}

	var builder strings.Builder
	for i, doc := range docs {
		builder.WriteString(fmt.Sprintf("[%d] %s\n", i+1, doc.Content))
	}
	return builder.String()
}

// AugmentationMiddleware injects retrieved context into the prompt before it reaches the provider.
func AugmentationMiddleware(store Retriever, systemTemplate, userTemplate string, opts ...AugmentationOption) blades.Middleware {
	if store == nil {
		return blades.Unary(func(next blades.RunHandler) blades.RunHandler {
			return func(ctx context.Context, prompt *blades.Prompt, modelOpts ...blades.ModelOption) (*blades.Message, error) {
				return next(ctx, prompt, modelOpts...)
			}
		})
	}

	options := augmentationOptions{
		topK:      3,
		formatter: BuildContext,
		logger:    log.Printf,
	}
	for _, opt := range opts {
		opt(&options)
	}

	return blades.Unary(func(next blades.RunHandler) blades.RunHandler {
		return func(ctx context.Context, prompt *blades.Prompt, modelOpts ...blades.ModelOption) (*blades.Message, error) {
			if prompt == nil {
				return next(ctx, prompt, modelOpts...)
			}

			latest := prompt.Latest()
			if latest == nil || latest.Role != blades.RoleUser {
				return next(ctx, prompt, modelOpts...)
			}

			question := strings.TrimSpace(latest.Text())
			if question == "" {
				return next(ctx, prompt, modelOpts...)
			}

			retrieveOpts := []RetrieveOption{}
			if options.topK > 0 {
				retrieveOpts = append(retrieveOpts, WithTopK(options.topK))
			}
			if len(options.filters) > 0 {
				filters := make(map[string]string, len(options.filters))
				for key, value := range options.filters {
					filters[key] = value
				}
				retrieveOpts = append(retrieveOpts, WithFilters(filters))
			}

			docs, err := store.Retrieve(ctx, question, retrieveOpts...)
			if err != nil {
				return nil, fmt.Errorf("retrieve context: %w", err)
			}
			if len(docs) == 0 {
				return next(ctx, prompt, modelOpts...)
			}

			contextText := options.formatter(docs)
			templateParams := map[string]any{
				"Context":  contextText,
				"Question": question,
			}

			messages := make([]*blades.Message, 0, len(prompt.Messages)+2)
			if n := len(prompt.Messages); n > 0 {
				messages = append(messages, prompt.Messages[:n-1]...)
			}

			if systemTemplate != "" {
				systemMsg, err := blades.NewTemplateMessage(blades.RoleSystem, systemTemplate, templateParams)
				if err != nil {
					return nil, fmt.Errorf("build augmented system prompt: %w", err)
				}
				messages = append(messages, systemMsg)
			}

			userMsg, err := blades.NewTemplateMessage(blades.RoleUser, userTemplate, templateParams)
			if err != nil {
				return nil, fmt.Errorf("build augmented user prompt: %w", err)
			}
			messages = append(messages, userMsg)

			if options.logger != nil {
				options.logger("[RAG Middleware] Retrieved %d documents for query %q", len(docs), question)
			}

			return next(ctx, blades.NewPrompt(messages...), modelOpts...)
		}
	})
}

type augmentationOptions struct {
	topK      int
	filters   map[string]string
	formatter func([]Document) string
	logger    func(string, ...any)
}

// AugmentationOption configures middleware behaviour.
type AugmentationOption func(*augmentationOptions)

// WithAugmentationTopK limits retrieved documents.
func WithAugmentationTopK(topK int) AugmentationOption {
	return func(opts *augmentationOptions) {
		opts.topK = topK
	}
}

// WithAugmentationFilters applies fetch-time filters.
func WithAugmentationFilters(filters map[string]string) AugmentationOption {
	return func(opts *augmentationOptions) {
		if len(filters) == 0 {
			return
		}
		if opts.filters == nil {
			opts.filters = make(map[string]string, len(filters))
		}
		for key, value := range filters {
			opts.filters[key] = value
		}
	}
}

// WithAugmentationFormatter overrides the default context formatter.
func WithAugmentationFormatter(formatter func([]Document) string) AugmentationOption {
	return func(opts *augmentationOptions) {
		if formatter != nil {
			opts.formatter = formatter
		}
	}
}

// WithAugmentationLogger overrides the default logger. Pass nil to disable logging.
func WithAugmentationLogger(logger func(string, ...any)) AugmentationOption {
	return func(opts *augmentationOptions) {
		opts.logger = logger
	}
}
