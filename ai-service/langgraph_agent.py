#!/usr/bin/env python3
"""
HealthSecure LangGraph Agent with Query Router
Intelligent routing between RAG search and patient database queries.
"""

import os
import json
import logging
from typing import Dict, List, Any, Optional, TypedDict, Annotated
from datetime import datetime
import asyncio
import aiomysql
import uuid

from langchain_community.chat_models import ChatOllama
from langchain_community.embeddings import OllamaEmbeddings
from langchain_core.messages import HumanMessage, AIMessage, SystemMessage
from langchain_core.prompts import ChatPromptTemplate
from langgraph.graph import StateGraph, END
from langgraph.graph.message import add_messages
from langgraph.checkpoint.memory import MemorySaver
import requests

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# State definition for LangGraph
class HealthSecureState(TypedDict):
    messages: Annotated[List[HumanMessage | AIMessage], add_messages]
    query_type: str  # "rag", "database", "combined", "general"
    query_intent: str  # "patient_lookup", "medical_records", "clinical_guidelines", etc.
    context: Dict[str, Any]  # Additional context data
    retrieved_chunks: List[Dict[str, Any]]  # RAG results
    database_results: List[Dict[str, Any]]  # Database query results
    user_info: Dict[str, Any]  # User context
    response_data: Dict[str, Any]  # Final response data for AG-UI

class HealthSecureLangGraphAgent:
    """LangGraph-based agent for HealthSecure with intelligent query routing."""

    def __init__(self,
                 ollama_base_url: str = "http://localhost:11434",
                 model_name: str = "gpt-oss:20b-cloud",
                 embedding_model: str = "mxbai-embed-large",
                 db_connection_string: str = None):

        self.ollama_base_url = ollama_base_url
        self.model_name = model_name
        self.embedding_model = embedding_model
        self.db_connection_string = db_connection_string or os.getenv(
            "DATABASE_URL",
            "mysql+aiomysql://healthsecure_user:password@localhost:3306/healthsecure"
        )

        # Initialize LLM and embeddings
        self.llm = ChatOllama(
            model=model_name,
            base_url=ollama_base_url,
            temperature=0.7
        )

        self.embeddings = OllamaEmbeddings(
            model=embedding_model,
            base_url=ollama_base_url
        )

        # Database connection pool
        self.db_pool = None

        # Build the graph
        self.graph = self._build_graph()

    async def initialize(self):
        """Initialize database connections."""
        try:
            # The connection string is for SQLAlchemy, but aiomysql needs parameters.
            # A bit of parsing is needed. This is a simplified example.
            # For production, use a more robust parsing library.
            conn_parts = self.db_connection_string.split('@')
            user_pass_db = conn_parts[0].split('//')[1]
            host_port_db = conn_parts[1]
            user_pass, db_name = user_pass_db.split('/')[0], host_port_db.split('/')[1]
            user, password = user_pass.split(':')
            host_port = host_port_db.split('/')[0]
            host, port = host_port.split(':')

            self.db_pool = await aiomysql.create_pool(
                host=host, port=int(port),
                user=user, password=password,
                db=db_name,
                autocommit=True
            )
            logger.info("Database connection pool initialized")
        except Exception as e:
            logger.error(f"Failed to initialize database: {e}")
            self.db_pool = None

    def _build_graph(self) -> StateGraph:
        """Build the LangGraph workflow."""

        # Create the state graph
        workflow = StateGraph(HealthSecureState)

        # Add nodes
        workflow.add_node("query_classifier", self._classify_query)
        workflow.add_node("rag_search", self._rag_search)
        workflow.add_node("database_query", self._database_query)
        workflow.add_node("combined_search", self._combined_search)
        workflow.add_node("response_generator", self._generate_response)

        # Add edges
        workflow.set_entry_point("query_classifier")

        # Conditional routing based on query type
        workflow.add_conditional_edges(
            "query_classifier",
            self._route_query,
            {
                "rag": "rag_search",
                "database": "database_query",
                "combined": "combined_search",
                "general": "response_generator"
            }
        )

        # All search nodes lead to response generation
        workflow.add_edge("rag_search", "response_generator")
        workflow.add_edge("database_query", "response_generator")
        workflow.add_edge("combined_search", "response_generator")
        workflow.add_edge("response_generator", END)

        # Compile with memory
        memory = MemorySaver()
        return workflow.compile(checkpointer=memory)

    async def _classify_query(self, state: HealthSecureState) -> Dict[str, Any]:
        """Classify the user query to determine routing."""

        last_message = state["messages"][-1]
        query = last_message.content if hasattr(last_message, 'content') else str(last_message)

        classification_prompt = ChatPromptTemplate.from_messages([
            ("system", """You are a query classifier for a healthcare system. Classify the user's query into one of these categories:

CATEGORIES:
1. "database" - Questions about specific patients, medical records, appointments, or operational data
   Examples: "Show me patient John Doe's records", "List all patients with diabetes", "What appointments are scheduled today?"

2. "rag" - Questions about medical knowledge, clinical guidelines, protocols, or general medical information
   Examples: "What are the symptoms of hypertension?", "How to treat pneumonia?", "Latest COVID protocols?"

3. "combined" - Questions that need both patient data AND medical knowledge
   Examples: "Based on patient X's symptoms, what condition might they have?", "Compare this patient's labs to normal ranges"

4. "general" - Greetings, general questions, or unclear queries
   Examples: "Hello", "Help me", "What can you do?"

Also determine the INTENT:
- patient_lookup, medical_records, appointments, lab_results (for database)
- clinical_guidelines, symptoms, treatments, medications (for rag)
- diagnosis_support, treatment_recommendation (for combined)
- greeting, help, unclear (for general)

Respond with JSON only: {{"category": "...", "intent": "...", "reasoning": "..."}}"""),
            ("human", "{query}")
        ])

        try:
            chain = classification_prompt | self.llm
            response = await chain.ainvoke({"query": query})

            # Parse JSON response
            classification = json.loads(response.content)

            return {
                "query_type": classification["category"],
                "query_intent": classification["intent"],
                "context": {"classification_reasoning": classification.get("reasoning", "")}
            }

        except Exception as e:
            logger.error(f"Query classification failed: {e}")
            # Default fallback
            return {
                "query_type": "general",
                "query_intent": "unclear",
                "context": {"error": str(e)}
            }

    def _route_query(self, state: HealthSecureState) -> str:
        """Route query based on classification."""
        return state["query_type"]

    async def _rag_search(self, state: HealthSecureState) -> Dict[str, Any]:
        """Perform RAG search on medical knowledge base."""

        last_message = state["messages"][-1]
        query = last_message.content if hasattr(last_message, 'content') else str(last_message)

        try:
            if not self.db_pool:
                return {"retrieved_chunks": [], "context": {"error": "Database not available"}}

            # Generate query embedding
            query_embedding = await asyncio.to_thread(
                self.embeddings.embed_query, query
            )

            # Search vector database
            async with self.db_pool.acquire() as conn:
                async with conn.cursor(aiomysql.DictCursor) as cur:
                    # Simplified RAG search without vector similarity for now
                    # Using basic text search until vector database is properly set up
                    await cur.execute("""
                        SELECT
                            rc.id as chunk_id,
                            rc.content,
                            rc.metadata,
                            rd.title as document_title,
                            rd.source_path,
                            0.5 as similarity
                        FROM rag_chunks rc
                        JOIN rag_documents rd ON rc.document_id = rd.id
                        WHERE rc.content LIKE %s
                        ORDER BY rc.id DESC
                        LIMIT 5
                    """, (f"%{query}%",))
                    search_results = await cur.fetchall()

                chunks = []
                for row in search_results:
                    chunks.append({
                        "chunk_id": str(row["chunk_id"]),
                        "content": row["content"],
                        "similarity": float(row["similarity"]),
                        "document_title": row["document_title"],
                        "source": row["source_path"],
                        "metadata": json.loads(row["metadata"]) if row["metadata"] else {}
                    })

                logger.info(f"RAG search returned {len(chunks)} chunks")
                return {"retrieved_chunks": chunks}

        except Exception as e:
            logger.error(f"RAG search failed: {e}")
            return {"retrieved_chunks": [], "context": {"error": str(e)}}

    async def _database_query(self, state: HealthSecureState) -> Dict[str, Any]:
        """Query patient database based on intent."""

        last_message = state["messages"][-1]
        query = last_message.content if hasattr(last_message, 'content') else str(last_message)
        intent = state.get("query_intent", "")

        try:
            if not self.db_pool:
                return {"database_results": [], "context": {"error": "Database not available"}}

            results = []

            async with self.db_pool.acquire() as conn:
                async with conn.cursor(aiomysql.DictCursor) as cur:
                    if intent == "patient_lookup":
                        # Extract patient name or ID from query using LLM
                        await cur.execute("""
                            SELECT id, first_name, last_name, date_of_birth, email, phone
                            FROM patients
                            WHERE is_active = 1
                            ORDER BY created_at DESC
                            LIMIT 10
                        """)
                        patient_results = await cur.fetchall()
                        results = [dict(row) for row in patient_results]

                    elif intent == "medical_records":
                        # Get recent medical records
                        await cur.execute("""
                            SELECT mr.id, mr.record_type, mr.content, mr.created_at,
                                   p.first_name, p.last_name
                            FROM medical_records mr
                            JOIN patients p ON mr.patient_id = p.id
                            ORDER BY mr.created_at DESC
                            LIMIT 10
                        """)
                        records_results = await cur.fetchall()
                        results = [dict(row) for row in records_results]

                    else:
                        # Generic patient statistics
                        await cur.execute("""
                            SELECT
                                COUNT(*) as total_patients,
                                COUNT(CASE WHEN created_at > NOW() - INTERVAL 30 DAY THEN 1 END) as new_patients_30d
                            FROM patients
                            WHERE is_active = 1
                        """)
                        stats = await cur.fetchone()
                        results = [dict(stats)] if stats else []

            logger.info(f"Database query returned {len(results)} results")
            return {"database_results": results}

        except Exception as e:
            logger.error(f"Database query failed: {e}")
            return {"database_results": [], "context": {"error": str(e)}}

    async def _combined_search(self, state: HealthSecureState) -> Dict[str, Any]:
        """Perform both RAG and database search."""

        # Run both searches in parallel
        rag_task = self._rag_search(state)
        db_task = self._database_query(state)

        rag_result, db_result = await asyncio.gather(rag_task, db_task)

        return {
            "retrieved_chunks": rag_result.get("retrieved_chunks", []),
            "database_results": db_result.get("database_results", []),
            "context": {
                **rag_result.get("context", {}),
                **db_result.get("context", {})
            }
        }

    async def _generate_response(self, state: HealthSecureState) -> Dict[str, Any]:
        """Generate final response based on gathered information."""

        last_message = state["messages"][-1]
        query = last_message.content if hasattr(last_message, 'content') else str(last_message)

        # Build context for response
        context_parts = []

        # Add RAG context
        if state.get("retrieved_chunks"):
            context_parts.append("RETRIEVED MEDICAL KNOWLEDGE:")
            for i, chunk in enumerate(state["retrieved_chunks"][:3], 1):
                context_parts.append(f"{i}. {chunk['content'][:300]}...")

        # Add database context
        if state.get("database_results"):
            context_parts.append("\nPATIENT/SYSTEM DATA:")
            for i, result in enumerate(state["database_results"][:5], 1):
                context_parts.append(f"{i}. {json.dumps(result, default=str)}")

        context_text = "\n".join(context_parts)

        # Generate response
        response_prompt = ChatPromptTemplate.from_messages([
            ("system", """You are HealthSecure AI Assistant, a medical AI helping healthcare professionals.

USER ROLE: {user_role}
QUERY TYPE: {query_type}
QUERY INTENT: {query_intent}

INSTRUCTIONS:
- Provide accurate, helpful responses based on the available context
- For patient data queries, present information clearly and respect privacy
- For medical knowledge queries, provide evidence-based information
- Always maintain HIPAA compliance
- If information is insufficient, say so clearly
- Include relevant source information when available

AVAILABLE CONTEXT:
{context}

Remember: You are assisting healthcare professionals with their clinical and administrative tasks."""),
            ("human", "{query}")
        ])

        try:
            chain = response_prompt | self.llm
            response = await chain.ainvoke({
                "query": query,
                "user_role": state.get("user_info", {}).get("role", "user"),
                "query_type": state.get("query_type", "general"),
                "query_intent": state.get("query_intent", "unclear"),
                "context": context_text
            })

            # Prepare response data for AG-UI
            response_data = {
                "response": response.content,
                "query_type": state.get("query_type"),
                "query_intent": state.get("query_intent"),
                "sources": {
                    "rag_chunks": len(state.get("retrieved_chunks", [])),
                    "database_results": len(state.get("database_results", []))
                },
                "retrieved_chunks": state.get("retrieved_chunks", []),
                "database_results": state.get("database_results", [])
            }

            # Add AI message to conversation
            ai_message = AIMessage(content=response.content)

            return {
                "messages": [ai_message],
                "response_data": response_data
            }

        except Exception as e:
            logger.error(f"Response generation failed: {e}")
            error_message = AIMessage(content=f"I encountered an error processing your request: {e}")

            return {
                "messages": [error_message],
                "response_data": {
                    "response": error_message.content,
                    "error": str(e)
                }
            }

    async def process_query(self,
                          query: str,
                          user_info: Dict[str, Any] = None,
                          thread_id: str = None) -> Dict[str, Any]:
        """Process a user query through the LangGraph workflow."""

        thread_id = thread_id or f"thread_{uuid.uuid4()}"
        user_info = user_info or {}

        # Initial state
        initial_state = {
            "messages": [HumanMessage(content=query)],
            "user_info": user_info,
            "context": {},
            "retrieved_chunks": [],
            "database_results": []
        }

        try:
            # Run the graph
            config = {"configurable": {"thread_id": thread_id}}
            result = await self.graph.ainvoke(initial_state, config)

            return {
                "success": True,
                "response": result["response_data"]["response"],
                "thread_id": thread_id,
                "query_type": result.get("query_type"),
                "query_intent": result.get("query_intent"),
                "retrieved_chunks": result.get("retrieved_chunks", []),
                "database_results": result.get("database_results", []),
                "model_used": self.model_name
            }

        except Exception as e:
            logger.error(f"Query processing failed: {e}")
            return {
                "success": False,
                "error": str(e),
                "response": f"I encountered an error: {e}",
                "thread_id": thread_id
            }

    async def cleanup(self):
        """Clean up resources."""
        if self.db_pool:
            self.db_pool.close()
            await self.db_pool.wait_closed()

# Test the agent
async def test_agent():
    """Test function for the LangGraph agent."""
    agent = HealthSecureLangGraphAgent()
    await agent.initialize()

    test_queries = [
        "Hello, how can you help me?",
        "Show me all patients",
        "What are the symptoms of diabetes?",
        "Based on recent lab results, what conditions should I consider?"
    ]

    for query in test_queries:
        print(f"\n=== Query: {query} ===")
        result = await agent.process_query(
            query=query,
            user_info={"role": "doctor", "name": "Dr. Test"}
        )

        print(f"Success: {result['success']}")
        print(f"Query Type: {result.get('query_type')}")
        print(f"Response: {result['response'][:200]}...")

    await agent.cleanup()

if __name__ == "__main__":
    asyncio.run(test_agent())