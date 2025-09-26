'use client';

import { useCoAgent, useCopilotAction } from "@copilotkit/react-core";
import { CopilotKitCSSProperties, CopilotSidebar } from "@copilotkit/react-ui";
import { useState } from "react";
import ClientOnly from "./client-only";

// RAG Agent State Type - must match the backend RAGState
type RAGState = {
  retrieved_chunks: RetrievedChunk[];
  current_query: SearchQuery | null;
  search_history: SearchQuery[];
  selected_chunk_id: string | null;
  total_chunks_in_kb: number;
  knowledge_base_status: string;
};

type RetrievedChunk = {
  chunk_id: string;
  document_id: string;
  content: string;
  similarity: number;
  metadata: Record<string, any>;
  document_title: string;
  document_source: string;
  highlight?: string;
};

type SearchQuery = {
  query: string;
  timestamp: string;
  match_count: number;
  search_type: string;
};

interface UserInfo {
  email: string;
  role: string;
}

interface RAGCopilotProps {
  user: UserInfo;
  onLogout: () => void;
}

export default function RAGCopilot({ user, onLogout }: RAGCopilotProps) {
  const [themeColor, setThemeColor] = useState("#4f46e5");
  const [expandedChunks, setExpandedChunks] = useState<Set<string>>(new Set());
  const [searchFilter, setSearchFilter] = useState("");

  // Connect to the RAG agent's shared state
  const { state, setState } = useCoAgent<RAGState>({
    name: "rag_agent",
    initialState: {
      retrieved_chunks: [],
      current_query: null,
      search_history: [],
      selected_chunk_id: null,
      total_chunks_in_kb: 0,
      knowledge_base_status: "initializing",
    },
  });

  // Frontend action for highlighting specific chunks
  useCopilotAction({
    name: "highlightChunk",
    description: "Highlight a specific chunk in the UI",
    parameters: [{
      name: "chunkId",
      type: "string",
      description: "The ID of the chunk to highlight",
      required: true,
    }],
    handler({ chunkId }) {
      setState({
        ...state,
        selected_chunk_id: chunkId,
      });
      // Auto-expand the highlighted chunk
      setExpandedChunks(prev => new Set(prev).add(chunkId));
    },
  });

  // Frontend action for theme customization
  useCopilotAction({
    name: "setThemeColor",
    parameters: [{
      name: "themeColor",
      description: "The theme color to set for the UI",
      required: true,
    }],
    handler({ themeColor }) {
      setThemeColor(themeColor);
    },
  });

  // Filter chunks based on search (with null checks)
  const filteredChunks = (state?.retrieved_chunks || []).filter(chunk =>
    searchFilter === "" ||
    chunk?.content?.toLowerCase().includes(searchFilter.toLowerCase()) ||
    chunk?.document_title?.toLowerCase().includes(searchFilter.toLowerCase())
  );

  return (
    <main style={{ "--copilot-kit-primary-color": themeColor } as CopilotKitCSSProperties}>
      <div className="flex h-screen">
        <RAGKnowledgeBaseView
          state={state}
          user={user}
          themeColor={themeColor}
          expandedChunks={expandedChunks}
          setExpandedChunks={setExpandedChunks}
          searchFilter={searchFilter}
          setSearchFilter={setSearchFilter}
          filteredChunks={filteredChunks}
          onLogout={onLogout}
        />
        <ClientOnly fallback={<div className="w-80 bg-gray-100 animate-pulse" />}>
          <CopilotSidebar
            clickOutsideToClose={false}
            defaultOpen={true}
            labels={{
              title: "HealthSecure AI Assistant",
              initial: `👋 Hi ${user.email}! I'm your HealthSecure AI assistant with access to a knowledge base containing ${state?.total_chunks_in_kb || "many"} chunks of medical information.

As a ${user.role}, you can ask me about:
- Medical procedures and protocols
- Patient care guidelines
- Clinical decision support
- Emergency procedures
- Healthcare compliance information

Knowledge Base Status: ${state?.knowledge_base_status || "initializing"}`,
            }}
          />
        </ClientOnly>
      </div>
    </main>
  );
}

function RAGKnowledgeBaseView({
  state,
  user,
  themeColor,
  expandedChunks,
  setExpandedChunks,
  searchFilter,
  setSearchFilter,
  filteredChunks,
  onLogout,
}: {
  state: RAGState;
  user: UserInfo;
  themeColor: string;
  expandedChunks: Set<string>;
  setExpandedChunks: React.Dispatch<React.SetStateAction<Set<string>>>;
  searchFilter: string;
  setSearchFilter: React.Dispatch<React.SetStateAction<string>>;
  filteredChunks: RetrievedChunk[];
  onLogout: () => void;
}) {
  const toggleChunkExpansion = (chunkId: string) => {
    setExpandedChunks(prev => {
      const newSet = new Set(prev);
      if (newSet.has(chunkId)) {
        newSet.delete(chunkId);
      } else {
        newSet.add(chunkId);
      }
      return newSet;
    });
  };

  return (
    <div className="flex-1 flex flex-col bg-gray-50 overflow-hidden">
      {/* Header */}
      <div
        style={{ backgroundColor: themeColor }}
        className="p-6 text-white shadow-lg"
      >
        <div className="flex justify-between items-start mb-4">
          <div>
            <h1 className="text-2xl font-bold mb-2">Knowledge Base Explorer</h1>
            <div className="flex items-center gap-4">
              <p className="text-sm opacity-90">
                {state?.retrieved_chunks?.length > 0
                  ? `${state.retrieved_chunks.length} chunks retrieved`
                  : "No chunks retrieved yet"}
              </p>
              <div className="text-sm bg-white/20 px-3 py-1 rounded-full">
                Status: {state?.knowledge_base_status || "initializing"}
              </div>
            </div>
          </div>

          <div className="flex items-center gap-4">
            <div className="text-right">
              <p className="text-sm font-medium">{user.email}</p>
              <p className="text-xs opacity-75 capitalize">{user.role}</p>
            </div>
            <button
              onClick={onLogout}
              className="px-3 py-1 text-sm bg-white/20 hover:bg-white/30 rounded-md transition-colors"
            >
              Logout
            </button>
          </div>
        </div>

        {/* Current Query Display */}
        {state?.current_query && (
          <div className="mt-4 bg-white/10 p-3 rounded-lg">
            <p className="text-xs uppercase tracking-wide opacity-75 mb-1">Current Query</p>
            <p className="font-medium">{state.current_query.query}</p>
            <p className="text-xs opacity-75 mt-1">
              {state.current_query.search_type} search • {state.current_query.match_count} results
            </p>
          </div>
        )}
      </div>

      {/* Search Filter */}
      {state?.retrieved_chunks?.length > 0 && (
        <div className="p-4 bg-white border-b">
          <input
            type="text"
            placeholder="Filter retrieved chunks..."
            value={searchFilter}
            onChange={(e) => setSearchFilter(e.target.value)}
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2"
            style={{ '--tw-ring-color': themeColor } as any}
          />
        </div>
      )}

      {/* Chunks Display */}
      <div className="flex-1 overflow-y-auto p-4">
        {filteredChunks.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-gray-500">
            {(state?.retrieved_chunks?.length || 0) === 0 ? (
              <>
                <SearchIcon className="w-16 h-16 mb-4 opacity-25" />
                <p className="text-lg font-medium mb-2">No chunks retrieved yet</p>
                <p className="text-sm text-center max-w-md">
                  Ask the assistant to search for information and the retrieved chunks will appear here.
                </p>
              </>
            ) : (
              <>
                <p className="text-lg font-medium mb-2">No matching chunks</p>
                <p className="text-sm">Try adjusting your filter criteria</p>
              </>
            )}
          </div>
        ) : (
          <div className="space-y-4">
            {filteredChunks.map((chunk, index) => (
              <ChunkCard
                key={chunk.chunk_id}
                chunk={chunk}
                index={index}
                isExpanded={expandedChunks.has(chunk.chunk_id)}
                isSelected={state?.selected_chunk_id === chunk.chunk_id}
                onToggle={() => toggleChunkExpansion(chunk.chunk_id)}
                themeColor={themeColor}
              />
            ))}
          </div>
        )}
      </div>

      {/* Search History Footer */}
      {state?.search_history?.length > 0 && (
        <div className="p-4 bg-white border-t">
          <p className="text-xs uppercase tracking-wide text-gray-500 mb-2">Recent Searches</p>
          <div className="flex flex-wrap gap-2">
            {state.search_history.slice(-5).map((query, idx) => (
              <span
                key={idx}
                className="text-xs bg-gray-100 px-2 py-1 rounded-full"
                title={new Date(query.timestamp).toLocaleString()}
              >
                {query.query}
              </span>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function ChunkCard({
  chunk,
  index,
  isExpanded,
  isSelected,
  onToggle,
  themeColor,
}: {
  chunk: RetrievedChunk;
  index: number;
  isExpanded: boolean;
  isSelected: boolean;
  onToggle: () => void;
  themeColor: string;
}) {
  const relevanceColor = chunk.similarity > 0.8 ? "bg-green-100" :
                         chunk.similarity > 0.6 ? "bg-yellow-100" : "bg-gray-100";

  return (
    <div
      className={`
        bg-white rounded-lg shadow-md overflow-hidden transition-all duration-200
        ${isSelected ? 'ring-2 ring-offset-2' : ''}
        ${isExpanded ? 'shadow-lg' : 'hover:shadow-lg'}
      `}
      style={{
        borderColor: isSelected ? themeColor : undefined,
        '--tw-ring-color': isSelected ? themeColor : undefined,
      } as any}
    >
      <div
        className="p-4 cursor-pointer select-none"
        onClick={onToggle}
      >
        <div className="flex items-start justify-between mb-2">
          <div className="flex-1">
            <div className="flex items-center gap-2 mb-1">
              <span className="text-xs font-medium text-gray-500">#{index + 1}</span>
              <span className={`text-xs px-2 py-1 rounded-full ${relevanceColor}`}>
                {(chunk.similarity * 100).toFixed(1)}% match
              </span>
              {isSelected && (
                <span
                  className="text-xs px-2 py-1 rounded-full text-white"
                  style={{ backgroundColor: themeColor }}
                >
                  Selected
                </span>
              )}
            </div>
            <h3 className="font-medium text-gray-900">{chunk.document_title}</h3>
            <p className="text-xs text-gray-500 mt-1">{chunk.document_source}</p>
          </div>
          <ChevronIcon expanded={isExpanded} />
        </div>

        {/* Preview of content */}
        <p className="text-sm text-gray-700 line-clamp-2">
          {chunk.highlight || chunk.content.substring(0, 150)}
          {chunk.content.length > 150 && !isExpanded && "..."}
        </p>
      </div>

      {/* Expanded Content */}
      {isExpanded && (
        <div className="px-4 pb-4 border-t">
          <div className="mt-4">
            <h4 className="text-xs font-medium text-gray-500 uppercase tracking-wide mb-2">
              Full Content
            </h4>
            <div className="bg-gray-50 p-3 rounded-lg">
              <p className="text-sm text-gray-700 whitespace-pre-wrap">{chunk.content}</p>
            </div>
          </div>

          {/* Metadata */}
          {Object.keys(chunk.metadata).length > 0 && (
            <div className="mt-4">
              <h4 className="text-xs font-medium text-gray-500 uppercase tracking-wide mb-2">
                Metadata
              </h4>
              <div className="bg-gray-50 p-3 rounded-lg">
                {Object.entries(chunk.metadata).map(([key, value]) => (
                  <div key={key} className="flex justify-between text-xs mb-1">
                    <span className="font-medium text-gray-600">{key}:</span>
                    <span className="text-gray-700">
                      {typeof value === 'object' ? JSON.stringify(value) : String(value)}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Chunk Details */}
          <div className="mt-4 flex gap-4 text-xs text-gray-500">
            <span>Chunk ID: {chunk.chunk_id}</span>
            <span>Document ID: {chunk.document_id}</span>
          </div>
        </div>
      )}
    </div>
  );
}

// Utility Icons
function SearchIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
      />
    </svg>
  );
}

function ChevronIcon({ expanded }: { expanded: boolean }) {
  return (
    <svg
      className={`w-5 h-5 text-gray-400 transform transition-transform ${
        expanded ? "rotate-180" : ""
      }`}
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M19 9l-7 7-7-7"
      />
    </svg>
  );
}