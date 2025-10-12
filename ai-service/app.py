#!/usr/bin/env python3
"""
HealthSecure AI Service - Python LangChain Microservice
Handles AI chat functionality with full LangSmith integration
"""

import os
import json
import uuid
import logging
from datetime import datetime
from typing import List, Dict, Any, Optional
from dataclasses import dataclass

from flask import Flask, request, jsonify
from flask_cors import CORS
import requests

# LangChain imports
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage, SystemMessage, AIMessage
from langchain_core.prompts import ChatPromptTemplate, MessagesPlaceholder
from langchain.agents import tool, create_openai_tools_agent, AgentExecutor

# LangSmith integration
from langsmith import Client as LangSmithClient
import langsmith

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Ollama web search tools integration
try:
    from ollama_web_search_tools import ollama_web_search_tools, test_ollama_api_key
    WEB_SEARCH_AVAILABLE = test_ollama_api_key()
    if WEB_SEARCH_AVAILABLE:
        logger.info("Ollama web search tools loaded and API key verified")
    else:
        logger.warning("Ollama web search tools loaded but API key not configured")
except ImportError as e:
    logger.warning(f"Ollama web search tools not available: {e}")
    ollama_web_search_tools = []
    WEB_SEARCH_AVAILABLE = False

# AG-UI RAG Agent integration
try:
    from true_agui_agent import TrueAGUIAgent
    AGUI_AVAILABLE = True
    logger.info("AG-UI RAG Agent integration enabled")
except ImportError as e:
    logger.warning(f"AG-UI RAG Agent not available: {e}")
    TrueAGUIAgent = None
    AGUI_AVAILABLE = False

# Docling RAG integration
try:
    from docling_rag_analyzer import DoclingRAGAnalyzer
    DOCLING_RAG_AVAILABLE = True
    logger.info("Docling RAG Analyzer integration enabled")
except ImportError as e:
    logger.warning(f"Docling RAG Analyzer not available: {e}")
    DoclingRAGAnalyzer = None
    DOCLING_RAG_AVAILABLE = False

# Initialize Flask app
app = Flask(__name__)
CORS(app)

# Load environment variables
def load_env_from_file(file_path: str):
    """Load environment variables from .env file"""
    env_vars = {}
    try:
        with open(file_path, 'r') as f:
            for line in f:
                line = line.strip()
                if line and not line.startswith('#') and '=' in line:
                    key, value = line.split('=', 1)
                    key = key.strip()
                    value = value.strip().strip('"').strip("'")
                    env_vars[key] = value
                    os.environ[key] = value
        logger.info(f"Loaded {len(env_vars)} environment variables from {file_path}")
    except Exception as e:
        logger.error(f"Failed to load env file {file_path}: {e}")
    return env_vars

# Load environment from Go backend configs
env_file_paths = [
    "../configs/.env",    # Development path
    "/app/.env",          # Docker path
    ".env"                # Local fallback
]

for env_file_path in env_file_paths:
    if os.path.exists(env_file_path):
        load_env_from_file(env_file_path)
        break

# Configuration
class Config:
    OPENROUTER_API_KEY = os.getenv("OPENROUTER_API_KEY", "your-openrouter-api-key")
    LANGCHAIN_API_KEY = os.getenv("LANGCHAIN_API_KEY")
    LANGCHAIN_PROJECT = os.getenv("LANGCHAIN_PROJECT", "healthsecure-ai")
    LANGCHAIN_TRACING_V2 = os.getenv("LANGCHAIN_TRACING_V2", "true").lower() == "true"
    MODEL_NAME = os.getenv("MODEL_NAME", "gpt-oss:20b-cloud")  # Use gpt-oss:20b-cloud by default
    MAX_TOKENS = 1000
    TEMPERATURE = 0.7

# Initialize LangSmith
if Config.LANGCHAIN_TRACING_V2 and Config.LANGCHAIN_API_KEY:
    os.environ["LANGCHAIN_TRACING_V2"] = "true"
    os.environ["LANGCHAIN_API_KEY"] = Config.LANGCHAIN_API_KEY
    os.environ["LANGCHAIN_PROJECT"] = Config.LANGCHAIN_PROJECT
    langsmith_client = LangSmithClient()
    logger.info("LangSmith tracing enabled")
else:
    langsmith_client = None
    logger.warning("LangSmith tracing disabled - missing API key or disabled")

# Data models
@dataclass
class ChatMessage:
    role: str
    content: str
    timestamp: str = None

@dataclass
class ChatRequest:
    message: str
    thread_id: str
    user_id: str
    user_role: str
    user_name: str
    history: List[Dict[str, Any]]
    max_tokens: Optional[int] = None
    temperature: Optional[float] = None

@dataclass
class ChatResponse:
    response: str
    thread_id: str
    run_id: str
    model_used: str
    new_title: Optional[str] = None
    tokens_used: Optional[int] = None

# Tools
@tool
def get_all_patients():
    """
    Retrieves a list of all patients in the system.
    Returns a list of patients with their ID, name, age, and other non-sensitive information.
    """
    try:
        backend_url = "http://localhost:8080"
        response = requests.get(f"{backend_url}/api/ai/patients")
        response.raise_for_status()
        return response.json()
    except Exception as e:
        logger.error(f"Error getting patients: {e}")
        return f"Error getting patients: {e}"

# Combine all available tools
tools = [get_all_patients] + ollama_web_search_tools

# Initialize OpenAI via OpenRouter
def create_llm():
    """Create LangChain LLM instance using OpenRouter"""
    return ChatOpenAI(
        model=Config.MODEL_NAME,
        api_key=Config.OPENROUTER_API_KEY,
        base_url="https://openrouter.ai/api/v1",
        max_tokens=Config.MAX_TOKENS,
        temperature=Config.TEMPERATURE,
        extra_headers={
            "HTTP-Referer": "https://healthsecure.ai",
            "X-Title": "HealthSecure AI Assistant"
        }
    )

# System prompts based on user roles
def create_system_prompt(user_role: str, user_name: str) -> str:
    """Create role-based system prompt"""
    base_prompt = f"""You are HealthSecure AI Assistant, a specialized medical AI designed to help healthcare professionals in a HIPAA-compliant environment.

Current User: {user_name} (Role: {user_role})
Current Time: {datetime.now().strftime("%Y-%m-%d %H:%M:%S")}

CAPABILITIES:
- Patient information queries and analysis
- Medical record interpretation and insights
- Healthcare protocol guidance and recommendations
- General medical knowledge and best practices
- Emergency procedure assistance
- Ollama web search for latest medical research and guidelines
- Real-time information retrieval from the internet
- Evidence-based medicine with current research access

IMPORTANT GUIDELINES:
- Always maintain HIPAA compliance and patient confidentiality
- Provide evidence-based medical information
- Acknowledge when additional medical expertise is needed
- Never diagnose patients - provide information to support clinical decision-making
- All interactions are logged for quality assurance and audit purposes

"""

    role_specific = {
        "doctor": """As a Doctor, you have full access to patient data and can:
- Review and analyze patient medical histories
- Get assistance with differential diagnosis considerations
- Access clinical decision support
- Review medication interactions and contraindications
- Get emergency protocol guidance""",
        
        "nurse": """As a Nurse, you can:
- Access patient care plans and medication schedules
- Get assistance with nursing protocols and procedures
- Review patient monitoring guidelines
- Access emergency response procedures
- Get medication administration guidance""",
        
        "admin": """As an Administrator, you can:
- Access system usage statistics and reports
- Get assistance with compliance and audit requirements
- Review system capabilities and configurations
- Access general medical information for training purposes"""
    }
    
    return base_prompt + role_specific.get(user_role, role_specific["admin"])

# Initialize the conversation chain
def create_agent():
    """Create the main conversation agent"""
    llm = create_llm()

    prompt = ChatPromptTemplate.from_messages([
        ("system", "{system_prompt}"),
        MessagesPlaceholder(variable_name="chat_history"),
        ("human", "{input}"),
        MessagesPlaceholder(variable_name="agent_scratchpad"),
    ])

    agent = create_openai_tools_agent(llm, tools, prompt)
    agent_executor = AgentExecutor(agent=agent, tools=tools, verbose=True)

    return agent_executor

def generate_thread_title(history: List[Dict[str, Any]], new_message: str) -> str:
    """Generate a title for the conversation history."""
    if not history and not new_message:
        return "New Conversation"

    # Create a prompt for the LLM to generate a title
    prompt = f"""
    Based on the following conversation, please generate a short, descriptive title (5 words or less).

    Conversation:
    """
    for msg in history:
        prompt += f"{msg.get('role')}: {msg.get('content')}\n"

    prompt += f"user: {new_message}\n"

    prompt += "\nTitle:"

    # Create a simple chain to generate the title
    llm = create_llm()
    chain = ChatPromptTemplate.from_template(prompt) | llm
    title = chain.invoke({}).content

    return title.strip().strip('"')

# Initialize the conversation chain globally
agent_executor = create_agent()

# Initialize AG-UI Agent if available
agui_agent = None
if AGUI_AVAILABLE:
    try:
        agui_agent = TrueAGUIAgent()
        logger.info("AG-UI RAG Agent initialized successfully")
    except Exception as e:
        logger.error(f"Failed to initialize AG-UI agent: {e}")
        agui_agent = None

# Initialize Docling RAG Analyzer
docling_rag = None
if DOCLING_RAG_AVAILABLE:
    try:
        docling_rag = DoclingRAGAnalyzer()
        logger.info("Docling RAG Analyzer initialized successfully")
        if docling_rag.qdrant_client:
            logger.info(f"✅ Qdrant connected: {docling_rag.qdrant_url}")
        else:
            logger.warning("⚠️ Qdrant not available - RAG search disabled")
    except Exception as e:
        logger.error(f"Failed to initialize Docling RAG: {e}")
        docling_rag = None

@app.route('/health', methods=['GET'])
def health_check():
    """Health check endpoint"""
    rag_status = "disabled"
    if docling_rag:
        if docling_rag.qdrant_client:
            rag_status = "enabled"
        else:
            rag_status = "no_qdrant"
    
    return jsonify({
        "status": "healthy",
        "service": "HealthSecure AI Service",
        "langsmith_enabled": Config.LANGCHAIN_TRACING_V2,
        "ollama_web_search_enabled": WEB_SEARCH_AVAILABLE,
        "agui_enabled": AGUI_AVAILABLE,
        "docling_rag_enabled": DOCLING_RAG_AVAILABLE,
        "rag_status": rag_status,
        "qdrant_url": docling_rag.qdrant_url if docling_rag else None,
        "model": Config.MODEL_NAME,
        "tools_available": len(tools),
        "search_provider": "ollama" if WEB_SEARCH_AVAILABLE else "none",
        "timestamp": datetime.now().isoformat()
    })

@app.route('/chat', methods=['POST'])
@langsmith.traceable(name="healthsecure_chat")
def chat():
    """Main chat endpoint"""
    try:
        # Parse request
        data = request.get_json()
        if not data:
            return jsonify({"error": "No JSON data provided"}), 400
            
        chat_req = ChatRequest(
            message=data.get('message', ''),
            thread_id=data.get('thread_id', str(uuid.uuid4())),
            user_id=data.get('user_id', ''),
            user_role=data.get('user_role', 'admin'),
            user_name=data.get('user_name', 'Unknown User'),
            history=data.get('history', []),
            max_tokens=data.get('max_tokens', Config.MAX_TOKENS),
            temperature=data.get('temperature', Config.TEMPERATURE)
        )
        
        if not chat_req.message.strip():
            return jsonify({"error": "Message cannot be empty"}), 400
            
        # Generate run ID for tracking
        run_id = str(uuid.uuid4())
        
        # Create system prompt
        system_prompt = create_system_prompt(chat_req.user_role, chat_req.user_name)

        # Format history
        chat_history = []
        for msg in chat_req.history:
            if msg.get('role') == 'user':
                chat_history.append(HumanMessage(content=msg.get('content')))
            elif msg.get('role') == 'assistant':
                chat_history.append(AIMessage(content=msg.get('content')))
        
        # Log the interaction
        logger.info(f"Chat request - Thread: {chat_req.thread_id}, User: {chat_req.user_id}, Role: {chat_req.user_role}")
        
        # RAG: Search for relevant context from documents
        rag_context = ""
        rag_sources = []
        if docling_rag and docling_rag.qdrant_client:
            try:
                import asyncio
                # Search for relevant chunks
                search_results = asyncio.run(
                    docling_rag.search_similar_chunks(chat_req.message, limit=3)
                )
                
                if search_results:
                    logger.info(f"RAG: Found {len(search_results)} relevant document chunks")
                    rag_context = "\n\n📚 Relevant Information from Medical Documents:\n"
                    for i, result in enumerate(search_results, 1):
                        rag_context += f"\n[Source {i}] (Relevance: {result['similarity']:.2%})\n"
                        rag_context += f"{result['content'][:500]}...\n"
                        rag_sources.append({
                            "chunk_id": result['chunk_id'],
                            "similarity": result['similarity'],
                            "source_file": result['metadata'].get('source_file', 'Unknown')
                        })
                    
                    # Add RAG context to system prompt
                    system_prompt += f"\n\n{rag_context}\n"
                    system_prompt += "\nPlease use the above document excerpts to provide accurate, evidence-based responses."
            except Exception as e:
                logger.warning(f"RAG search failed: {e}")
        
        # Process with LangChain
        response = agent_executor.invoke(
            {
                "system_prompt": system_prompt,
                "input": chat_req.message,
                "chat_history": chat_history
            }
        )
        
        # Extract response content
        ai_response = response.get('output')

        # Generate a title for the conversation
        new_title = generate_thread_title(chat_req.history, chat_req.message)
        
        # Create response object
        chat_response = ChatResponse(
            response=ai_response,
            thread_id=chat_req.thread_id,
            run_id=run_id,
            model_used=Config.MODEL_NAME,
            new_title=new_title,
            tokens_used=None  # OpenRouter doesn't always provide this
        )
        
        logger.info(f"Chat response generated - Run ID: {run_id}")
        
        return jsonify({
            "response": chat_response.response,
            "thread_id": chat_response.thread_id,
            "run_id": chat_response.run_id,
            "model_used": chat_response.model_used,
            "new_title": chat_response.new_title,
            "success": True
        })
        
    except Exception as e:
        logger.error(f"Chat error: {str(e)}", exc_info=True)
        return jsonify({
            "error": "Failed to process chat message",
            "details": str(e),
            "success": False
        }), 500

@app.route('/history/<thread_id>', methods=['GET'])
def get_chat_history(thread_id: str):
    return jsonify({"error": "History is now managed by the Go backend"}), 404


@app.route('/threads', methods=['GET'])
def list_threads():
    """List available chat threads (basic implementation)"""
    # This would typically query the database for thread metadata
    return jsonify({
        "threads": [],
        "message": "Thread listing requires database integration"
    })

@app.route('/agui/chat', methods=['POST'])
def agui_chat():
    """AG-UI enhanced chat endpoint with dynamic frontend modification capabilities"""
    if not agui_agent:
        return jsonify({
            "error": "AG-UI functionality not available",
            "fallback_to": "/chat"
        }), 503

    try:
        data = request.get_json()
        if not data:
            return jsonify({"error": "No JSON data provided"}), 400

        query = data.get('message', '')
        user_info = {
            "user_id": data.get('user_id', ''),
            "user_name": data.get('user_name', 'Unknown User'),
            "user_role": data.get('user_role', 'admin')
        }
        thread_id = data.get('thread_id', f"agui_{uuid.uuid4().hex[:8]}")

        if not query.strip():
            return jsonify({"error": "Message cannot be empty"}), 400

        # Process with AG-UI agent using asyncio.run to handle async call
        logger.info(f"AG-UI processing: {query} for user {user_info['user_name']}")
        import asyncio
        try:
            result = asyncio.run(agui_agent.process_query(query, user_info, thread_id))
        except RuntimeError:
            # If there's already an event loop running, create a new thread
            import concurrent.futures
            with concurrent.futures.ThreadPoolExecutor() as executor:
                future = executor.submit(lambda: asyncio.run(agui_agent.process_query(query, user_info, thread_id)))
                result = future.result()

        if result["success"]:
            return jsonify({
                "response": result["response"],
                "thread_id": result["thread_id"],
                "route": result.get("route", "unknown"),
                "agui_state": result.get("agui_state", {}),
                "model_used": "AG-UI Enhanced",
                "success": True
            })
        else:
            return jsonify({
                "error": "AG-UI processing failed",
                "details": result.get("error", "Unknown error"),
                "fallback_response": result.get("response", ""),
                "success": False
            }), 500

    except Exception as e:
        logger.error(f"AG-UI chat error: {str(e)}", exc_info=True)
        return jsonify({
            "error": "Failed to process AG-UI chat message",
            "details": str(e),
            "success": False
        }), 500

@app.route('/agui/state', methods=['GET'])
def get_agui_state():
    """Get current AG-UI state for frontend synchronization"""
    if not agui_agent:
        return jsonify({"error": "AG-UI functionality not available"}), 503

    # In a real implementation, this would fetch the current state from storage
    return jsonify({
        "ui_components": [],
        "layout_config": {},
        "theme_config": {},
        "current_activity": None,
        "activity_progress": 0.0
    })

@app.route('/rag/upload', methods=['POST'])
def upload_document():
    """Upload and ingest a medical document for RAG"""
    if not docling_rag:
        return jsonify({"error": "Docling RAG not available"}), 503
    
    if not docling_rag.qdrant_client:
        return jsonify({"error": "Qdrant not connected"}), 503
    
    try:
        # Check if file is present
        if 'file' not in request.files:
            return jsonify({"error": "No file provided"}), 400
        
        file = request.files['file']
        if file.filename == '':
            return jsonify({"error": "Empty filename"}), 400
        
        # Save file temporarily
        import tempfile
        import os
        from pathlib import Path
        
        temp_dir = tempfile.gettempdir()
        file_path = os.path.join(temp_dir, file.filename)
        file.save(file_path)
        
        logger.info(f"Processing uploaded file: {file.filename}")
        
        # Process with Docling RAG
        import asyncio
        result = asyncio.run(docling_rag.analyze_pdf(file_path))
        
        # Clean up temp file
        os.remove(file_path)
        
        return jsonify({
            "success": True,
            "filename": file.filename,
            "chunks_created": len(result.get('rag_chunks', [])),
            "tables_extracted": len(result.get('structured_data', {}).get('tables', [])),
            "pages": result.get('total_pages', 0),
            "file_hash": result.get('file_hash', ''),
            "message": "Document ingested successfully"
        })
        
    except Exception as e:
        logger.error(f"Document upload error: {str(e)}", exc_info=True)
        return jsonify({
            "error": "Failed to process document",
            "details": str(e)
        }), 500

@app.route('/rag/search', methods=['POST'])
def search_documents():
    """Search for similar content in ingested documents"""
    if not docling_rag:
        return jsonify({"error": "Docling RAG not available"}), 503
    
    if not docling_rag.qdrant_client:
        return jsonify({"error": "Qdrant not connected"}), 503
    
    try:
        data = request.get_json()
        query = data.get('query', '')
        limit = data.get('limit', 5)
        
        if not query:
            return jsonify({"error": "Query cannot be empty"}), 400
        
        logger.info(f"RAG search: {query}")
        
        # Search with Docling RAG
        import asyncio
        results = asyncio.run(docling_rag.search_similar_chunks(query, limit=limit))
        
        return jsonify({
            "success": True,
            "query": query,
            "results_count": len(results),
            "results": results
        })
        
    except Exception as e:
        logger.error(f"RAG search error: {str(e)}", exc_info=True)
        return jsonify({
            "error": "Failed to search documents",
            "details": str(e)
        }), 500

@app.route('/rag/status', methods=['GET'])
def rag_status():
    """Get RAG system status and statistics"""
    if not docling_rag:
        return jsonify({
            "rag_available": False,
            "message": "Docling RAG not initialized"
        })
    
    status_data = {
        "rag_available": True,
        "qdrant_connected": docling_rag.qdrant_client is not None,
        "qdrant_url": docling_rag.qdrant_url,
        "collection_name": docling_rag.collection_name,
        "vector_size": docling_rag.vector_size,
        "embedding_model": "mxbai-embed-large"
    }
    
    # Get collection stats if Qdrant is connected
    if docling_rag.qdrant_client:
        try:
            collection_info = docling_rag.qdrant_client.get_collection(docling_rag.collection_name)
            status_data["documents_count"] = collection_info.points_count
            status_data["collection_status"] = collection_info.status
        except Exception as e:
            status_data["collection_error"] = str(e)
    
    return jsonify(status_data)

@app.errorhandler(404)
def not_found(error):
    return jsonify({"error": "Endpoint not found"}), 404

@app.errorhandler(500)
def internal_error(error):
    return jsonify({"error": "Internal server error"}), 500

if __name__ == '__main__':
    # Ensure API key is available
    if not Config.OPENROUTER_API_KEY:
        logger.error("OPENROUTER_API_KEY is not set!")
        exit(1)
    
    logger.info("Starting HealthSecure AI Service...")
    logger.info(f"Model: {Config.MODEL_NAME}")
    logger.info(f"LangSmith Tracing: {'Enabled' if Config.LANGCHAIN_TRACING_V2 else 'Disabled'}")
    
    # Run the Flask app
    try:
        # Try port 5000 first
        port = int(os.getenv('AI_SERVICE_PORT', 5000))
        # Use localhost for Windows compatibility
        host = '127.0.0.1' if os.name == 'nt' else '0.0.0.0'
        app.run(
            host=host,
            port=port,
            debug=os.getenv('ENVIRONMENT', 'production') == 'development'
        )
    except OSError as e:
        if "permission" in str(e).lower() or "access" in str(e).lower():
            # Try alternative port
            logger.warning(f"Port {port} access denied, trying port 5001...")
            app.run(
                host=host,
                port=5001,
                debug=os.getenv('ENVIRONMENT', 'production') == 'development'
            )
        else:
            raise e
