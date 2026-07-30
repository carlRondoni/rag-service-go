# RAG Service Go

A Retrieval-Augmented Generation (RAG) service built in Go, following Clean Architecture principles. The project integrates with Qdrant as the vector database and Ollama for local LLM inference and embeddings.

## Quickstart

1. Start the local infrastructure (Qdrant, Ollama, Observability tools) using Docker Compose:
   ```bash
    make local-complete-first-start
   ```

2. Stop the full infra:
   ```bash
   make local-docker-stop
   ```

## API Endpoints

### Health Checks

- **`GET /health`**
  Checks the general availability of the HTTP server.
  *Success Response (200 OK)*

- **`GET /db/health`**
  Checks the connectivity and health of the Qdrant vector database.
  *Success Response (200 OK):*
  ```json
  {
    "status": "ok",
    "db": "up"
  }
  ```
  *Failure Response (503 Service Unavailable):*
  ```json
  {
    "status": "degraded",
    "db": "down"
  }
  ```

- **`GET /llm/health`**
  Checks the connectivity and health of the Ollama LLM provider.
  *Success Response (200 OK):*
  ```json
  {
    "status": "ok",
    "llm": "up"
  }
  ```
  *Failure Response (503 Service Unavailable):*
  ```json
  {
    "status": "degraded",
    "llm": "down"
  }
  ```

### Core Endpoints

- **`POST /ingest`**
  Ingests a text document, chunks it, generates embeddings via Ollama, and stores them in Qdrant.
  *Request Body:*
  ```json
  {
    "text": "Your long document text goes here...",
    "documentId": "doc-123" // Optional. Defaults to "default-doc"
  }
  ```
  *Success Response (202 Accepted)*

- **`POST /generate`**
  Performs a similarity search in Qdrant for the given prompt, constructs a context, and generates an answer using Ollama.
  *Request Body:*
  ```json
  {
    "prompt": "What is the main topic of the document?"
  }
  ```
  *Success Response (200 OK):*
  ```json
  {
    "response": "The document primarily discusses..."
  }
  ```
## Architecture & Tech Stack

> **[TODO: Add an architecture diagram here if you have one. You can use Mermaid or insert an image link to show the flow.]**

This project adheres to **Clean Architecture** to ensure testability and decoupling of business logic from external frameworks:
- **Domain Layer**: Core business models (e.g., `Chunk`, `Embedding`) and interfaces (`LLMClient`, `VectorStore`).
- **Application Layer**: Use cases and orchestration (e.g., Document Ingestion, Answer Generation, Health Checks).
- **Infrastructure Layer**: Integration with external tools (Qdrant, Ollama) and HTTP Handlers.

**Key Technologies:**
- **Go** (Golang)
- **[Qdrant](https://qdrant.tech/)** (Vector Search Engine)
- **[Ollama](https://ollama.com/)** (Local LLM Inference and Embeddings)
- **Observability**: [Loki](https://grafana.com/oss/loki/) & Alloy (for structured logging & metrics)
- **Testing**: `httptest`, `testify/mock`, and Integration Tests.

## Features

> **[TODO: Customize and expand this list based on what is completely finished in the project.]**

- ✅ Local, private RAG pipeline with zero external API dependencies.
- ✅ Robust health-checking for external services (LLM, Vector DB).
- ✅ Document Ingestion pipeline (Chunking & Embedding).
- ✅ Question-Answering using similarity search.
- ✅ Comprehensive unit and integration testing.

## Future Improvements / Roadmap

> **[TODO: Mention what you plan to add next to show recruiters your thought process, ambition, and understanding of production systems.]**

- Add support for alternative cloud LLM providers (e.g., OpenAI, Anthropic) via the Domain interface.
- Implement an authentication and rate-limiting middleware.
- Integrate OpenTelemetry for distributed tracing.
- CI/CD pipeline using GitHub Actions.

## License

MIT License