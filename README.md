# OpenAI Chat Completion and Embedding with PostgreSQL

This repository demonstrates how to use OpenAI's chat completion and text embedding capabilities in conjunction with a PostgreSQL database to store embeddings and perform similarity searches.

For further details, refer to my blog post: [Konsep RAG pada LLM dan Contoh Implementasinya](https://syahidfrd.com/blog/konsep-rag-llm-dan-contoh-implementasinya)

## Overview

This application allows you to:

1. Generate text embeddings from OpenAI's `text-embedding-3-small` model.
2. Store and retrieve these embeddings from PostgreSQL using the `pgvector` extension.
3. Perform similarity search on the stored embeddings to find relevant context for chat completion.

## Requirements

- Golang
- PostgreSQL with `pgvector` extension installed
- OpenAI API Key
