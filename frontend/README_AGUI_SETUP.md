# AGUI-Enabled RAG Agent with CopilotKit

This is a complete implementation of a Retrieval-Augmented Generation (RAG) agent using Pydantic AI with AG-UI protocol support, integrated with CopilotKit as the frontend client.

## Architecture Overview

- **Backend**: Pydantic AI agent with AG-UI support running as an ASGI server
- **Frontend**: Next.js with CopilotKit for the user interface
- **Protocol**: AG-UI for real-time agent-user interaction with shared state
- **Database**: PostgreSQL with pgvector for vector similarity search

## Key Features

1. **Shared State Management**: Real-time synchronization between agent and UI
2. **RAG Capabilities**: Semantic and hybrid search through knowledge base
3. **Visual Chunk Display**: Retrieved chunks are displayed in the UI with metadata
4. **Interactive Exploration**: Users can expand chunks, view metadata, and filter results
5. **Search History**: Tracks previous queries for context
6. **Custom Events**: Supports custom UI updates through AG-UI events

## Setup Instructions

### Prerequisites

- PostgreSQL with pgvector extension installed
- Node.js 18+ and npm/yarn
- Python 3.9+
- OpenAI API key (or compatible LLM provider)

### Backend Setup

1. **Install Python dependencies**:
```bash
cd agent
pip install -r requirements.txt
```

2. **Set up environment variables**:
Create a `.env` file in the `agent` directory:
```env
# LLM Configuration
LLM_PROVIDER=openai
LLM_MODEL=gpt-4
LLM_API_KEY=your-openai-api-key
LLM_BASE_URL=https://api.openai.com/v1  # Optional for OpenAI

# Database Configuration
DATABASE_URL=postgresql://user:password@localhost:5432/ragdb
DB_POOL_MIN_SIZE=5
DB_POOL_MAX_SIZE=20

# Embeddings
EMBEDDING_MODEL=text-embedding-3-small
EMBEDDING_DIMENSIONS=1536

# Optional: LLM Settings
LLM_TEMPERATURE=0.7
LLM_MAX_TOKENS=4096
```

3. **Initialize the database**:
```bash
python -m agent.utils.db_utils init
```

4. **Ingest documents** (if you have documents to index):
```bash
python -m agent.ingestion.ingest --documents ./documents --clean
```

5. **Start the AGUI server**:
```bash
python agent/agent.py
```
The server will start on http://localhost:8000

### Frontend Setup

1. **Install Node dependencies**:
```bash
npm install
# or
yarn install
```

2. **Start the development server**:
```bash
npm run dev
# or
yarn dev
```
The frontend will be available at http://localhost:3000

## Usage

1. Open http://localhost:3000 in your browser
2. The RAG Assistant will appear in the right sidebar
3. Ask questions to search the knowledge base
4. Retrieved chunks will appear in the left panel
5. Click on chunks to expand and view full content and metadata
6. Use the filter box to search within retrieved chunks

## Shared State Structure

The shared state between the agent and frontend includes:

```typescript
type RAGState = {
  retrieved_chunks: RetrievedChunk[];      // Chunks from knowledge base
  current_query: SearchQuery | null;       // Current search query
  search_history: SearchQuery[];           // History of searches
  selected_chunk_id: string | null;        // Currently highlighted chunk
  total_chunks_in_kb: number;              // Total chunks in database
  knowledge_base_status: string;           // Status of the knowledge base
}
```

## Agent Tools

The RAG agent provides several tools:

- `search_knowledge_base`: Performs semantic/hybrid search and updates shared state
- `clear_search_results`: Clears the current search results
- `select_chunk`: Highlights a specific chunk in the UI
- `get_knowledge_base_stats`: Gets statistics about the knowledge base
- `display_search_results`: Triggers UI display of search results

## Frontend Actions

The frontend provides actions that the agent can call:

- `highlightChunk`: Highlights and expands a specific chunk
- `setThemeColor`: Changes the UI theme color

## Customization

### Modifying the Agent

Edit `agent/agent.py` to:
- Add new tools with `@rag_agent.tool` decorator
- Modify the RAGState model for different shared state
- Adjust search behavior or scoring

### Modifying the Frontend

Edit `src/app/page.tsx` to:
- Change the UI layout and components
- Add new visualizations for chunks
- Customize the chunk display format
- Add new frontend actions

### Ingestion Pipeline

The ingestion pipeline in `agent/ingestion/` can be customized to:
- Support additional document formats
- Extract custom metadata
- Modify chunking strategies
- Add document preprocessing steps

## Troubleshooting

### Common Issues

1. **Database connection errors**: Verify PostgreSQL is running and credentials are correct
2. **Agent not responding**: Check the agent server is running on port 8000
3. **No chunks displayed**: Verify documents have been ingested into the database
4. **AGUI errors**: Ensure pydantic-ai[agui] is properly installed

### Debug Mode

To enable debug logging:
```bash
# Backend
PYTHONUNBUFFERED=1 python agent/agent.py

# Frontend
npm run dev -- --verbose
```

## Architecture Benefits

1. **Real-time Sync**: AG-UI protocol ensures frontend and backend stay synchronized
2. **Type Safety**: Pydantic models provide type validation for shared state
3. **Scalability**: ASGI server can handle multiple concurrent connections
4. **Flexibility**: Easy to swap LLM providers or add new tools
5. **User Experience**: Interactive chunk exploration enhances understanding

## Next Steps

- Add authentication and user sessions
- Implement chunk feedback and ratings
- Add document upload through the UI
- Create custom visualization for different document types
- Implement incremental indexing for new documents
- Add export functionality for search results