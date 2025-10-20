# RAG 示例

此目录收录了基于 `github.com/go-kratos/blades` 工具集的检索增强生成（RAG）示例，演示多种集成方式。

## 目录结构
- `graph/`：使用 `flow.Graph` 将多个 RAG 节点串联成流水线。
- `middleware/`：通过中间件在运行时向 Agent 注入检索上下文。
- `shared/`：示例复用的辅助组件（句子切分器、内存存储、简单重排器）。

## 准备条件
- Go 1.24 及以上版本。
- `contrib/openai` 支持的 LLM 服务密钥（在环境变量中设置 `OPENAI_API_KEY`）。

## 运行示例

```bash
# 使用 flow.Graph 组织的流水线
go run ./examples/rag/graph

# 运行时在中间件中增强提示词
go run ./examples/rag/middleware
```

示例会在终端输出执行日志，并打印最终生成的答案。

## 核心概念
- **切分与索引**：`shared.SentenceChunker` 按句子划分文本并避免空块；`shared.SimpleMemoryStore` 用于索引和检索。
- **检索与重排**：`SimpleMemoryStore.Retrieve` 与 `SimpleReranker` 提供适用于演示的轻量打分逻辑。
- **答案生成**：`blades.Agent` 封装模型提供方，根据检索到的上下文生成回复。

欢迎根据业务需要扩展这些组件，或替换成真实的向量库、嵌入模型、重排模型等，以构建生产级 RAG 流程。
