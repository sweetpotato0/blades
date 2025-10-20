package rag

import "github.com/go-kratos/blades"

// Document 表示检索得到的文档或片段，包含内容、评分及自定义元数据。
type Document = blades.Document

// Indexer 负责向底层存储中添加或删除文档，以便后续检索使用。
type Indexer = blades.Indexer

// RetrieveOptions 包含检索时的可选参数。
type RetrieveOptions = blades.RetrieveOptions

// RetrieveOption 是用于配置检索选项的函数类型。
type RetrieveOption = blades.RetrieveOption

// WithTopK 设置返回的最大文档数量。
func WithTopK(topK int) RetrieveOption {
	return blades.WithTopK(topK)
}

// WithFilters 设置过滤条件。
func WithFilters(filters map[string]string) RetrieveOption {
	return blades.WithFilters(filters)
}

// WithFilter 添加单个过滤条件。
func WithFilter(key, value string) RetrieveOption {
	return blades.WithFilter(key, value)
}

// Retriever 接口负责根据请求检索相关文档。
type Retriever = blades.Retriever

// Reranker 接口负责对初检索结果进行重排序，提升相关性。
type Reranker = blades.Reranker
