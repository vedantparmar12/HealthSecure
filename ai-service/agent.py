"""Main AGUI-enabled RAG agent implementation with shared state."""

from pydantic_ai import Agent, RunContext
from pydantic import BaseModel, Field
from typing import List, Dict, Any, Optional
from datetime import datetime
from textwrap import dedent
import json

from ag_ui.core import CustomEvent, EventType, StateSnapshotEvent, StateDeltaEvent
from pydantic_ai.ag_ui import StateDeps

from providers import get_llm_model
from dependencies import AgentDependencies
from prompts import MAIN_SYSTEM_PROMPT
from tools import semantic_search, hybrid_search


class RetrievedChunk(BaseModel):
    """Model for a retrieved chunk with metadata."""
    chunk_id: str = Field(description="Unique identifier for the chunk")
    document_id: str = Field(description="ID of the source document")
    content: str = Field(description="The actual text content of the chunk")
    similarity: float = Field(description="Similarity score to the query")
    metadata: Dict[str, Any] = Field(default_factory=dict, description="Additional metadata")
    document_title: str = Field(description="Title of the source document")
    document_source: str = Field(description="Source/path of the document")
    highlight: Optional[str] = Field(default=None, description="Highlighted matching text")


class SearchQuery(BaseModel):
    """Model for a search query."""
    query: str = Field(description="The search query text")
    timestamp: str = Field(description="When the query was made")
    match_count: int = Field(default=10, description="Number of results requested")
    search_type: str = Field(default="semantic", description="Type of search performed")


class RAGState(BaseModel):
    """Shared state for the RAG agent."""
    retrieved_chunks: List[RetrievedChunk] = Field(
        default_factory=list,
        description="List of chunks retrieved from the knowledge base"
    )
    current_query: Optional[SearchQuery] = Field(
        default=None,
        description="The current search query being processed"
    )
    search_history: List[SearchQuery] = Field(
        default_factory=list,
        description="History of search queries"
    )
    selected_chunk_id: Optional[str] = Field(
        default=None,
        description="ID of the currently selected/highlighted chunk"
    )
    total_chunks_in_kb: int = Field(
        default=0,
        description="Total number of chunks in the knowledge base"
    )
    knowledge_base_status: str = Field(
        default="ready",
        description="Status of the knowledge base (ready, indexing, error)"
    )


# Create the RAG agent with AGUI support
rag_agent = Agent(
    get_llm_model(),
    deps_type=StateDeps[RAGState],
    system_prompt=MAIN_SYSTEM_PROMPT
)


@rag_agent.tool
async def search_knowledge_base(
    ctx: RunContext[StateDeps[RAGState]],
    query: str,
    match_count: Optional[int] = None,
    search_type: Optional[str] = "semantic"
) -> StateSnapshotEvent:
    """
    Search the knowledge base and update shared state with results.

    This tool performs semantic or hybrid search on the knowledge base,
    retrieves relevant chunks, and updates the shared state so the frontend
    can display the results.

    Args:
        ctx: Agent runtime context with state dependencies
        query: Search query text
        match_count: Number of results to return (default: 10)
        search_type: Type of search - "semantic" or "hybrid" (default: semantic)

    Returns:
        StateSnapshotEvent with updated retrieved chunks
    """
    # Create search query record
    search_query = SearchQuery(
        query=query,
        timestamp=datetime.now().isoformat(),
        match_count=match_count or 10,
        search_type=search_type
    )

    # Update current query in state
    ctx.deps.state.current_query = search_query
    ctx.deps.state.search_history.append(search_query)

    # Keep only last 10 queries in history
    if len(ctx.deps.state.search_history) > 10:
        ctx.deps.state.search_history = ctx.deps.state.search_history[-10:]

    try:
        # Get the underlying dependencies for database access
        # We need to extract the AgentDependencies from the StateDeps wrapper
        agent_deps = AgentDependencies()
        await agent_deps.initialize()

        # Create a context wrapper for the search tools
        class DepsWrapper:
            def __init__(self, deps):
                self.deps = deps

        deps_ctx = DepsWrapper(agent_deps)

        # Perform the search based on type
        if search_type == "hybrid":
            results = await hybrid_search(
                ctx=deps_ctx,
                query=query,
                match_count=match_count
            )

            # Convert hybrid search results to RetrievedChunk format
            chunks = [
                RetrievedChunk(
                    chunk_id=str(result['chunk_id']),
                    document_id=str(result['document_id']),
                    content=result['content'],
                    similarity=result['combined_score'],
                    metadata={
                        **result['metadata'],
                        'vector_similarity': result['vector_similarity'],
                        'text_similarity': result['text_similarity']
                    },
                    document_title=result['document_title'],
                    document_source=result['document_source']
                )
                for result in results
            ]
        else:
            results = await semantic_search(
                ctx=deps_ctx,
                query=query,
                match_count=match_count
            )

            # Convert SearchResult to RetrievedChunk
            chunks = [
                RetrievedChunk(
                    chunk_id=result.chunk_id,
                    document_id=result.document_id,
                    content=result.content,
                    similarity=result.similarity,
                    metadata=result.metadata,
                    document_title=result.document_title,
                    document_source=result.document_source
                )
                for result in results
            ]

        # Update state with retrieved chunks
        ctx.deps.state.retrieved_chunks = chunks

        # Clean up
        await agent_deps.cleanup()

        # Return state snapshot event
        return StateSnapshotEvent(
            type=EventType.STATE_SNAPSHOT,
            snapshot=ctx.deps.state.model_dump(),
        )

    except Exception as e:
        print(f"Search error: {e}")
        # Clear chunks on error and update status
        ctx.deps.state.retrieved_chunks = []
        ctx.deps.state.knowledge_base_status = f"error: {str(e)}"

        return StateSnapshotEvent(
            type=EventType.STATE_SNAPSHOT,
            snapshot=ctx.deps.state.model_dump(),
        )


@rag_agent.tool
async def clear_search_results(ctx: RunContext[StateDeps[RAGState]]) -> StateSnapshotEvent:
    """
    Clear the current search results from the shared state.

    Args:
        ctx: Agent runtime context with state dependencies

    Returns:
        StateSnapshotEvent with cleared chunks
    """
    ctx.deps.state.retrieved_chunks = []
    ctx.deps.state.current_query = None
    ctx.deps.state.selected_chunk_id = None

    return StateSnapshotEvent(
        type=EventType.STATE_SNAPSHOT,
        snapshot=ctx.deps.state.model_dump(),
    )


@rag_agent.tool
async def select_chunk(ctx: RunContext[StateDeps[RAGState]], chunk_id: str) -> StateSnapshotEvent:
    """
    Select/highlight a specific chunk in the UI.

    Args:
        ctx: Agent runtime context with state dependencies
        chunk_id: ID of the chunk to select

    Returns:
        StateSnapshotEvent with updated selection
    """
    ctx.deps.state.selected_chunk_id = chunk_id

    return StateSnapshotEvent(
        type=EventType.STATE_SNAPSHOT,
        snapshot=ctx.deps.state.model_dump(),
    )


@rag_agent.tool
async def get_knowledge_base_stats(ctx: RunContext[StateDeps[RAGState]]) -> StateSnapshotEvent:
    """
    Get statistics about the knowledge base.

    Args:
        ctx: Agent runtime context with state dependencies

    Returns:
        StateSnapshotEvent with knowledge base statistics
    """
    try:
        agent_deps = AgentDependencies()
        await agent_deps.initialize()

        # Get chunk count from database
        async with agent_deps.db_pool.acquire() as conn:
            count_result = await conn.fetchval("SELECT COUNT(*) FROM chunks")
            ctx.deps.state.total_chunks_in_kb = count_result or 0
            ctx.deps.state.knowledge_base_status = "ready"

        await agent_deps.cleanup()

    except Exception as e:
        ctx.deps.state.knowledge_base_status = f"error: {str(e)}"

    return StateSnapshotEvent(
        type=EventType.STATE_SNAPSHOT,
        snapshot=ctx.deps.state.model_dump(),
    )


@rag_agent.tool
async def display_search_results(ctx: RunContext[StateDeps[RAGState]]) -> CustomEvent:
    """
    Display the current search results to the user.

    This tool creates a custom event to trigger UI updates for displaying
    the retrieved chunks in a user-friendly format.

    Args:
        ctx: Agent runtime context with state dependencies

    Returns:
        CustomEvent to trigger UI display
    """
    return CustomEvent(
        type=EventType.CUSTOM,
        name="DisplaySearchResults",
        value={
            "chunks": [chunk.model_dump() for chunk in ctx.deps.state.retrieved_chunks],
            "query": ctx.deps.state.current_query.model_dump() if ctx.deps.state.current_query else None,
            "total_results": len(ctx.deps.state.retrieved_chunks)
        }
    )


@rag_agent.instructions
async def rag_instructions(ctx: RunContext[StateDeps[RAGState]]) -> str:
    """
    Dynamic instructions for the RAG agent based on current state.

    Args:
        ctx: The run context containing RAG state information.

    Returns:
        Instructions string for the RAG agent.
    """

    has_chunks = len(ctx.deps.state.retrieved_chunks) > 0
    current_query = ctx.deps.state.current_query

    base_instructions = dedent(
        f"""
        You are an intelligent RAG (Retrieval-Augmented Generation) assistant with access to a knowledge base.
        Your primary function is to help users find relevant information from the knowledge base and provide
        accurate, contextual answers based on the retrieved information.

        IMPORTANT INSTRUCTIONS:
        1. Always use the `search_knowledge_base` tool to search for relevant information before answering questions
        2. Base your answers primarily on the retrieved chunks from the knowledge base
        3. When you retrieve information, the chunks will be displayed in the UI for the user to explore
        4. Use the `display_search_results` tool after searching to ensure the UI is updated
        5. If you cannot find relevant information, be honest about it
        6. You can use the `select_chunk` tool to highlight specific chunks that are most relevant
        7. Use `clear_search_results` when starting a new topic or when asked to clear the results

        Knowledge Base Status: {ctx.deps.state.knowledge_base_status}
        Total chunks in knowledge base: {ctx.deps.state.total_chunks_in_kb}
        """
    )

    if has_chunks:
        chunks_summary = dedent(
            f"""

            CURRENT STATE:
            - You have {len(ctx.deps.state.retrieved_chunks)} chunks retrieved from the search
            - Current query: "{current_query.query if current_query else 'None'}"
            - The retrieved chunks are displayed in the UI for the user to explore

            RETRIEVED INFORMATION:
            You should base your response on the following retrieved chunks:

            """
        )

        # Add summaries of top chunks
        for i, chunk in enumerate(ctx.deps.state.retrieved_chunks[:5], 1):
            chunks_summary += f"""
            Chunk {i} (Score: {chunk.similarity:.3f}):
            Source: {chunk.document_title} ({chunk.document_source})
            Content: {chunk.content[:200]}...

            """

        return base_instructions + chunks_summary

    else:
        return base_instructions + dedent(
            """

            CURRENT STATE:
            - No chunks currently retrieved
            - Use the search_knowledge_base tool to find relevant information
            - Previous searches: {} searches in history
            """.format(len(ctx.deps.state.search_history))
        )


# Convert agent to AGUI app
app = rag_agent.to_ag_ui(deps=StateDeps(RAGState()))

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8001)