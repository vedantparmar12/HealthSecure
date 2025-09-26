#!/usr/bin/env python3
"""
True AG-UI Agent for HealthSecure
Implements real AG-UI functionality where the agent can dynamically modify frontend code,
create components, and manage UI state through tools and LangGraph routing.
"""

import os
import json
import asyncio
import logging
from typing import Dict, List, Any, Optional, TypedDict, Annotated
from datetime import datetime
import uuid
import re

from pydantic import BaseModel, Field
from langchain_community.chat_models import ChatOllama
from langchain_core.messages import HumanMessage, AIMessage, SystemMessage
from langchain_core.prompts import ChatPromptTemplate
from langgraph.graph import StateGraph, END
from langgraph.graph.message import add_messages
from langgraph.checkpoint.memory import MemorySaver

from pdf_rag_analyzer import PDFRAGAnalyzer, PDFAnalysisResult

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# AG-UI State Management
class UIComponent(BaseModel):
    """Represents a dynamic UI component."""
    id: str = Field(description="Unique component ID")
    type: str = Field(description="Component type (e.g., 'chart', 'table', 'form')")
    props: Dict[str, Any] = Field(default_factory=dict, description="Component properties")
    position: str = Field(default="main", description="Where to render ('main', 'sidebar', 'modal')")
    temporary: bool = Field(default=False, description="If true, component will auto-remove")
    created_at: str = Field(default_factory=lambda: datetime.now().isoformat())

class AGUIState(BaseModel):
    """Complete AG-UI state for dynamic frontend control."""
    # UI State
    ui_components: List[UIComponent] = Field(default_factory=list, description="Active UI components")
    layout_config: Dict[str, Any] = Field(default_factory=dict, description="Layout configuration")
    theme_config: Dict[str, Any] = Field(default_factory=dict, description="Theme settings")

    # Data State
    patient_data: Optional[Dict[str, Any]] = Field(default=None, description="Current patient context")
    report_analysis: Optional[Dict[str, Any]] = Field(default=None, description="PDF report analysis")
    search_results: List[Dict[str, Any]] = Field(default_factory=list, description="RAG search results")

    # Activity State
    current_activity: Optional[str] = Field(default=None, description="Current agent activity")
    activity_progress: float = Field(default=0.0, description="Progress percentage (0-100)")
    activity_steps: List[Dict[str, Any]] = Field(default_factory=list, description="Activity step history")

class TrueAGUIAgentState(TypedDict):
    """LangGraph state combining messages with AG-UI state."""
    messages: Annotated[List[HumanMessage | AIMessage], add_messages]
    agui_state: AGUIState
    route_decision: str  # "chat", "agui", "database", "rag_analysis"
    user_info: Dict[str, Any]
    context: Dict[str, Any]

class TrueAGUIAgent:
    """Complete AG-UI agent with dynamic frontend modification capabilities."""

    def __init__(self, ollama_base_url: str = "http://localhost:11434", model_name: str = "gpt-oss:20b-cloud"):
        self.ollama_base_url = ollama_base_url
        self.model_name = model_name
        self.project_root = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))

        # Initialize LLM
        self.llm = ChatOllama(
            model=model_name,
            base_url=ollama_base_url,
            temperature=0.7
        )

        # Initialize PDF RAG analyzer
        self.pdf_analyzer = PDFRAGAnalyzer(ollama_base_url=ollama_base_url)

        # Build the routing graph
        self.graph = self._build_routing_graph()

    def _build_routing_graph(self) -> StateGraph:
        """Build LangGraph with intelligent routing."""
        workflow = StateGraph(TrueAGUIAgentState)

        # Add nodes
        workflow.add_node("router", self._route_query)
        workflow.add_node("normal_chat", self._handle_normal_chat)
        workflow.add_node("agui_modification", self._handle_agui_modification)
        workflow.add_node("database_query", self._handle_database_query)
        workflow.add_node("rag_analysis", self._handle_rag_analysis)

        # Set entry point
        workflow.set_entry_point("router")

        # Add conditional routing
        workflow.add_conditional_edges(
            "router",
            self._decide_route,
            {
                "chat": "normal_chat",
                "agui": "agui_modification",
                "database": "database_query",
                "rag": "rag_analysis"
            }
        )

        # All paths lead to END
        workflow.add_edge("normal_chat", END)
        workflow.add_edge("agui_modification", END)
        workflow.add_edge("database_query", END)
        workflow.add_edge("rag_analysis", END)

        # Compile with memory
        memory = MemorySaver()
        return workflow.compile(checkpointer=memory)

    async def _route_query(self, state: TrueAGUIAgentState) -> Dict[str, Any]:
        """Intelligent routing based on query analysis with aggressive UI detection."""
        last_message = state["messages"][-1]
        query = last_message.content if hasattr(last_message, 'content') else str(last_message)

        # AGGRESSIVE UI DETECTION - Check for UI keywords first
        query_lower = query.lower()
        ui_keywords = [
            "change", "color", "blue", "green", "red", "theme", "ui", "interface",
            "chart", "pie", "bar", "graph", "component", "add", "create",
            "modify", "layout", "sidebar", "background", "style", "design"
        ]

        # If any UI keywords are found, force route to AG-UI
        has_ui_keywords = any(keyword in query_lower for keyword in ui_keywords)

        if has_ui_keywords:
            logger.info(f"🎨 UI KEYWORDS DETECTED - Forcing AG-UI route for: {query}")
            return {
                "route_decision": "agui",
                "context": {
                    "routing_confidence": 1.0,
                    "routing_reasoning": f"UI keywords detected: {[k for k in ui_keywords if k in query_lower]}"
                }
            }

        # Check for database keywords
        db_keywords = ["patient", "show", "find", "list", "search", "database", "records"]
        has_db_keywords = any(keyword in query_lower for keyword in db_keywords)

        if has_db_keywords:
            logger.info(f"🗄️ DATABASE KEYWORDS DETECTED - Routing to database for: {query}")
            return {
                "route_decision": "database",
                "context": {
                    "routing_confidence": 0.9,
                    "routing_reasoning": f"Database keywords detected: {[k for k in db_keywords if k in query_lower]}"
                }
            }

        # Check for RAG/PDF keywords
        rag_keywords = ["analyze", "pdf", "report", "document", "file"]
        has_rag_keywords = any(keyword in query_lower for keyword in rag_keywords)

        if has_rag_keywords:
            logger.info(f"📄 RAG KEYWORDS DETECTED - Routing to RAG for: {query}")
            return {
                "route_decision": "rag",
                "context": {
                    "routing_confidence": 0.9,
                    "routing_reasoning": f"RAG keywords detected: {[k for k in rag_keywords if k in query_lower]}"
                }
            }

        # Default to chat for everything else
        logger.info(f"💬 NO SPECIAL KEYWORDS - Routing to chat for: {query}")
        return {
            "route_decision": "chat",
            "context": {
                "routing_confidence": 0.8,
                "routing_reasoning": "No specific keywords detected, defaulting to chat"
            }
        }

    def _decide_route(self, state: TrueAGUIAgentState) -> str:
        """Return the routing decision."""
        return state.get("route_decision", "chat")

    async def _handle_normal_chat(self, state: TrueAGUIAgentState) -> Dict[str, Any]:
        """Handle normal medical chat queries."""
        last_message = state["messages"][-1]
        query = last_message.content if hasattr(last_message, 'content') else str(last_message)

        chat_prompt = ChatPromptTemplate.from_messages([
            ("system", """You are HealthSecure AI Assistant, a medical AI helping healthcare professionals.

Provide helpful, accurate medical information while maintaining HIPAA compliance.
Keep responses concise and professional. If you need to access patient data or modify the UI,
suggest that the user ask for those specific actions."""),
            ("human", "{query}")
        ])

        try:
            chain = chat_prompt | self.llm
            response = await chain.ainvoke({"query": query})

            ai_message = AIMessage(content=response.content)
            return {"messages": [ai_message]}

        except Exception as e:
            logger.error(f"Chat handling failed: {e}")
            error_message = AIMessage(content=f"I encountered an error: {e}")
            return {"messages": [error_message]}

    async def _handle_agui_modification(self, state: TrueAGUIAgentState) -> Dict[str, Any]:
        """Handle AG-UI dynamic frontend modifications."""
        last_message = state["messages"][-1]
        query = last_message.content if hasattr(last_message, 'content') else str(last_message)

        agui_state = state.get("agui_state", AGUIState())

        # Update activity
        agui_state.current_activity = "ui_modification"
        agui_state.activity_progress = 10.0
        agui_state.activity_steps.append({
            "step": "analyzing_ui_request",
            "message": "Analyzing UI modification request...",
            "timestamp": datetime.now().isoformat()
        })

        # Determine what UI changes are needed
        modification_prompt = ChatPromptTemplate.from_messages([
            ("system", """You are an AG-UI modification agent. Analyze the user's request and determine what UI changes to make.

AVAILABLE ACTIONS:
1. modify_theme - Change colors, fonts, layout (e.g., "blue theme", "change color", "dark mode")
2. add_component - Add charts, forms, tables, widgets (e.g., "pie chart", "patient chart", "data table")
3. modify_layout - Change sidebar, main area, positioning
4. create_modal - Create popup dialogs or forms
5. modify_existing - Update existing components

SPECIFIC COMPONENT TYPES:
- pie_chart: For age groups, diagnosis distribution, patient statistics
- bar_chart: For comparing values across categories
- line_chart: For trends over time
- patient_table: For displaying patient data
- summary_card: For key information display
- data_grid: For detailed tabular data

For each action, provide:
- action_type: One of the actions above
- target: What to modify/create
- parameters: Specific changes to make
- code_changes: React/CSS code modifications needed

Respond with JSON only: {{"actions": [...]}}"""),
            ("human", "{query}")
        ])

        try:
            agui_state.activity_progress = 30.0
            agui_state.activity_steps.append({
                "step": "planning_modifications",
                "message": "Planning UI modifications...",
                "timestamp": datetime.now().isoformat()
            })

            chain = modification_prompt | self.llm
            response = await chain.ainvoke({"query": query})

            try:
                modifications = json.loads(response.content)
            except:
                # If JSON parsing fails, analyze the query directly for common patterns
                query_lower = query.lower()
                if "pie chart" in query_lower or "chart" in query_lower:
                    modifications = {
                        "actions": [{
                            "action_type": "add_component",
                            "target": "pie_chart",
                            "parameters": {"title": "Patient Age Groups", "category": "age_groups"},
                            "position": "main"
                        }]
                    }
                elif "blue" in query_lower and ("color" in query_lower or "theme" in query_lower):
                    modifications = {
                        "actions": [{
                            "action_type": "modify_theme",
                            "target": "blue_theme",
                            "parameters": {"color": "blue"}
                        }]
                    }
                elif "green" in query_lower and ("color" in query_lower or "theme" in query_lower):
                    modifications = {
                        "actions": [{
                            "action_type": "modify_theme",
                            "target": "green_theme",
                            "parameters": {"color": "green"}
                        }]
                    }
                else:
                    modifications = {
                        "actions": [{
                            "action_type": "add_component",
                            "target": "notification",
                            "parameters": {"message": "UI modification requested", "type": "info"}
                        }]
                    }

            agui_state.activity_progress = 60.0
            agui_state.activity_steps.append({
                "step": "executing_modifications",
                "message": "Executing UI modifications...",
                "timestamp": datetime.now().isoformat()
            })

            # Execute the modifications
            for action in modifications.get("actions", []):
                await self._execute_ui_action(action, agui_state)

            agui_state.activity_progress = 100.0
            agui_state.activity_steps.append({
                "step": "modifications_complete",
                "message": "UI modifications completed successfully!",
                "timestamp": datetime.now().isoformat()
            })

            agui_state.current_activity = None

            ai_message = AIMessage(content=f"I've successfully applied the UI modifications you requested. The changes are now active in your interface.")

            return {
                "messages": [ai_message],
                "agui_state": agui_state
            }

        except Exception as e:
            logger.error(f"AG-UI modification failed: {e}")
            error_message = AIMessage(content=f"I encountered an error while modifying the UI: {e}")
            return {"messages": [error_message], "agui_state": agui_state}

    async def _execute_ui_action(self, action: Dict[str, Any], agui_state: AGUIState):
        """Execute a specific UI action."""
        action_type = action.get("action_type")

        if action_type == "modify_theme":
            # Modify theme configuration
            theme_changes = action.get("parameters", {})

            # Handle specific theme requests
            if "blue" in str(action.get("target", "")).lower() or "blue" in str(theme_changes).lower():
                theme_changes = {
                    "primary_color": "#2196f3",
                    "secondary_color": "#1976d2",
                    "background_color": "#e3f2fd",
                    "text_color": "#0d47a1",
                    "theme_name": "blue"
                }
            elif "green" in str(action.get("target", "")).lower():
                theme_changes = {
                    "primary_color": "#4caf50",
                    "secondary_color": "#388e3c",
                    "background_color": "#e8f5e8",
                    "text_color": "#1b5e20",
                    "theme_name": "green"
                }

            agui_state.theme_config.update(theme_changes)

        elif action_type == "add_component":
            # Add a new UI component
            target = action.get("target", "generic")
            parameters = action.get("parameters", {})

            # Handle specific component types
            if "pie_chart" in target.lower() or "pie chart" in target.lower():
                # Create sample age group data for pie chart
                sample_data = [
                    {"name": "0-18 years", "value": 25, "color": "#8884d8"},
                    {"name": "19-35 years", "value": 35, "color": "#82ca9d"},
                    {"name": "36-50 years", "value": 22, "color": "#ffc658"},
                    {"name": "51-65 years", "value": 12, "color": "#ff7c7c"},
                    {"name": "65+ years", "value": 6, "color": "#8dd1e1"}
                ]

                component = UIComponent(
                    id=f"pie_chart_{uuid.uuid4().hex[:8]}",
                    type="pie_chart",
                    props={
                        "title": "Patient Age Groups Distribution",
                        "data": sample_data,
                        "width": 400,
                        "height": 300,
                        "showLegend": True,
                        "showTooltip": True
                    },
                    position="main"
                )
            else:
                component = UIComponent(
                    id=f"comp_{uuid.uuid4().hex[:8]}",
                    type=target,
                    props=parameters,
                    position=action.get("position", "main")
                )

            agui_state.ui_components.append(component)

        elif action_type == "modify_layout":
            # Modify layout configuration
            layout_changes = action.get("parameters", {})
            agui_state.layout_config.update(layout_changes)

        elif action_type == "create_modal":
            # Create a modal component
            modal = UIComponent(
                id=f"modal_{uuid.uuid4().hex[:8]}",
                type="modal",
                props=action.get("parameters", {}),
                position="overlay",
                temporary=True
            )
            agui_state.ui_components.append(modal)

        elif action_type == "modify_existing":
            # Modify existing component
            target_id = action.get("target")
            for component in agui_state.ui_components:
                if component.id == target_id or component.type == target_id:
                    component.props.update(action.get("parameters", {}))
                    break

        # Also perform actual file modifications if requested
        if action.get("code_changes"):
            await self._modify_frontend_code(action["code_changes"])

    async def _modify_frontend_code(self, code_changes: Dict[str, Any]):
        """Actually modify the frontend React code files."""
        try:
            file_path = code_changes.get("file_path")
            if not file_path:
                return

            full_path = os.path.join(self.project_root, file_path)

            if os.path.exists(full_path):
                with open(full_path, 'r', encoding='utf-8') as f:
                    content = f.read()

                # Apply changes based on type
                change_type = code_changes.get("type")

                if change_type == "replace":
                    old_code = code_changes.get("old")
                    new_code = code_changes.get("new")
                    content = content.replace(old_code, new_code)

                elif change_type == "insert":
                    position = code_changes.get("position", "end")
                    new_code = code_changes.get("code")

                    if position == "end":
                        content += f"\n{new_code}"
                    elif position == "start":
                        content = f"{new_code}\n{content}"

                # Write back the modified content
                with open(full_path, 'w', encoding='utf-8') as f:
                    f.write(content)

                logger.info(f"Modified frontend file: {file_path}")

        except Exception as e:
            logger.error(f"Failed to modify frontend code: {e}")

    async def _handle_database_query(self, state: TrueAGUIAgentState) -> Dict[str, Any]:
        """Handle patient database queries."""
        # This would integrate with the existing database query logic
        ai_message = AIMessage(content="Database query handling - would connect to patient database")
        return {"messages": [ai_message]}

    async def _handle_rag_analysis(self, state: TrueAGUIAgentState) -> Dict[str, Any]:
        """Handle RAG analysis of medical documents/PDFs."""
        last_message = state["messages"][-1]
        query = last_message.content if hasattr(last_message, 'content') else str(last_message)

        agui_state = state.get("agui_state", AGUIState())

        # Update activity
        agui_state.current_activity = "rag_analysis"
        agui_state.activity_progress = 10.0
        agui_state.activity_steps.append({
            "step": "initializing_analysis",
            "message": "Initializing document analysis...",
            "timestamp": datetime.now().isoformat()
        })

        try:
            # Check if query contains file path or if we should analyze uploaded files
            file_path = self._extract_file_path_from_query(query)

            if file_path and os.path.exists(file_path):
                agui_state.activity_progress = 30.0
                agui_state.activity_steps.append({
                    "step": "processing_pdf",
                    "message": f"Processing PDF: {os.path.basename(file_path)}",
                    "timestamp": datetime.now().isoformat()
                })

                # Analyze the PDF
                patient_context = agui_state.patient_data
                analysis_result = await self.pdf_analyzer.analyze_pdf(file_path, patient_context)

                agui_state.activity_progress = 70.0
                agui_state.activity_steps.append({
                    "step": "extracting_insights",
                    "message": "Extracting medical insights from document...",
                    "timestamp": datetime.now().isoformat()
                })

                # Update AG-UI state with analysis results
                agui_state.report_analysis = analysis_result

                # Create UI components to display the analysis
                await self._create_analysis_ui_components(analysis_result, agui_state)

                agui_state.activity_progress = 100.0
                agui_state.activity_steps.append({
                    "step": "analysis_complete",
                    "message": "Document analysis completed successfully!",
                    "timestamp": datetime.now().isoformat()
                })

                # Generate response based on analysis
                response_prompt = ChatPromptTemplate.from_messages([
                    ("system", """You are analyzing a medical document. Based on the extracted information, provide a comprehensive summary that includes:

1. Key findings from the report
2. Important patient data
3. Any notable lab results or vital signs
4. Recommendations for follow-up

Present the information in a clear, professional manner suitable for healthcare professionals.

Analysis Data: {analysis_data}"""),
                    ("human", "{query}")
                ])

                chain = response_prompt | self.llm
                response = await chain.ainvoke({
                    "query": query,
                    "analysis_data": analysis_result.analysis_summary
                })

                ai_message = AIMessage(content=response.content)

            else:
                # No file found, provide instructions
                ai_message = AIMessage(content="To analyze a medical document, please upload a PDF file or provide the file path. I can extract patient information, lab results, diagnoses, and create visualizations of the data.")

            agui_state.current_activity = None

            return {
                "messages": [ai_message],
                "agui_state": agui_state
            }

        except Exception as e:
            logger.error(f"RAG analysis failed: {e}")
            error_message = AIMessage(content=f"I encountered an error while analyzing the document: {e}")
            return {"messages": [error_message], "agui_state": agui_state}

    def _extract_file_path_from_query(self, query: str) -> Optional[str]:
        """Extract file path from user query."""
        import re

        # Look for common patterns indicating file paths
        patterns = [
            r'analyze\s+(.+\.pdf)',
            r'process\s+(.+\.pdf)',
            r'file:\s*(.+\.pdf)',
            r'"([^"]+\.pdf)"',
            r"'([^']+\.pdf)'"
        ]

        for pattern in patterns:
            match = re.search(pattern, query, re.IGNORECASE)
            if match:
                file_path = match.group(1).strip()
                # Convert relative paths to absolute
                if not os.path.isabs(file_path):
                    file_path = os.path.join(self.project_root, file_path)
                return file_path

        return None

    async def _create_analysis_ui_components(self, analysis_result: PDFAnalysisResult, agui_state: AGUIState):
        """Create UI components to display PDF analysis results."""

        # Create a summary card component
        summary_component = UIComponent(
            id=f"pdf_summary_{uuid.uuid4().hex[:8]}",
            type="summary_card",
            props={
                "title": "Document Analysis Summary",
                "content": analysis_result.analysis_summary,
                "file_name": os.path.basename(analysis_result.file_path),
                "pages": analysis_result.total_pages,
                "sections": len(analysis_result.sections)
            },
            position="main"
        )
        agui_state.ui_components.append(summary_component)

        # Create structured data component if available
        if analysis_result.extracted_data:
            data_component = UIComponent(
                id=f"pdf_data_{uuid.uuid4().hex[:8]}",
                type="structured_data",
                props={
                    "title": "Extracted Patient Data",
                    "data": analysis_result.extracted_data,
                    "expandable": True
                },
                position="sidebar"
            )
            agui_state.ui_components.append(data_component)

        # Create sections component
        sections_component = UIComponent(
            id=f"pdf_sections_{uuid.uuid4().hex[:8]}",
            type="document_sections",
            props={
                "title": "Document Sections",
                "sections": [
                    {
                        "type": section.section_type,
                        "content": section.content[:200] + "..." if len(section.content) > 200 else section.content,
                        "confidence": section.confidence,
                        "page": section.page_number
                    }
                    for section in analysis_result.sections
                ]
            },
            position="sidebar"
        )
        agui_state.ui_components.append(sections_component)

    async def process_query(self, query: str, user_info: Dict[str, Any] = None, thread_id: str = None) -> Dict[str, Any]:
        """Process a query through the AG-UI routing system."""
        thread_id = thread_id or f"thread_{uuid.uuid4()}"
        user_info = user_info or {}

        # Initial state
        initial_state = {
            "messages": [HumanMessage(content=query)],
            "agui_state": AGUIState(),
            "route_decision": "",
            "user_info": user_info,
            "context": {}
        }

        try:
            # Run the graph
            config = {"configurable": {"thread_id": thread_id}}
            result = await self.graph.ainvoke(initial_state, config)

            return {
                "success": True,
                "response": result["messages"][-1].content if result["messages"] else "No response generated",
                "route": result.get("route_decision", "unknown"),
                "agui_state": result.get("agui_state", AGUIState()).model_dump(),
                "thread_id": thread_id
            }

        except Exception as e:
            logger.error(f"Query processing failed: {e}")
            return {
                "success": False,
                "error": str(e),
                "response": f"I encountered an error: {e}",
                "thread_id": thread_id
            }

# Test function
async def test_true_agui():
    """Test the True AG-UI agent."""
    agent = TrueAGUIAgent()

    test_queries = [
        "Hello, how can you help me?",
        "Change the UI color to dark green",
        "Add a patient chart to the dashboard",
        "Show me all diabetic patients",
        "Analyze this medical report PDF"
    ]

    for query in test_queries:
        print(f"\n=== Query: {query} ===")
        result = await agent.process_query(query)
        print(f"Route: {result.get('route')}")
        print(f"Response: {result['response'][:200]}...")

        if result.get("agui_state"):
            agui_state = result["agui_state"]
            if agui_state.get("ui_components"):
                print(f"UI Components: {len(agui_state['ui_components'])}")

if __name__ == "__main__":
    asyncio.run(test_true_agui())