#!/usr/bin/env python3
"""
Ollama Web Search Tools for HealthSecure AI Service
Uses Ollama's native web search and web fetch capabilities
"""

import os
import logging
import json
import requests
from typing import List, Dict, Any, Optional
from langchain.agents import tool

logger = logging.getLogger(__name__)

# Get Ollama API key from environment
OLLAMA_API_KEY = os.getenv('OLLAMA_API_KEY')
OLLAMA_API_BASE = "https://ollama.com/api"

@tool
def ollama_web_search(query: str, max_results: int = 5) -> str:
    """
    Search the web using Ollama's native web search API.
    This provides access to current information from the internet.

    Args:
        query: The search query string
        max_results: Maximum number of results to return (default: 5)

    Returns:
        JSON string containing search results with titles, URLs, and content
    """
    try:
        logger.info(f"Performing Ollama web search for: {query}")

        if not OLLAMA_API_KEY:
            return json.dumps({
                "error": "Ollama API key not configured. Please set OLLAMA_API_KEY environment variable.",
                "query": query,
                "results": []
            })

        # Make direct API call to Ollama web search
        headers = {
            "Authorization": f"Bearer {OLLAMA_API_KEY}",
            "Content-Type": "application/json"
        }

        data = {
            "query": query,
            "max_results": max_results
        }

        response = requests.post(
            f"{OLLAMA_API_BASE}/web_search",
            headers=headers,
            json=data,
            timeout=30
        )
        response.raise_for_status()

        result = response.json()

        # Format the response for consistency
        formatted_response = {
            "query": query,
            "results": result.get('results', []),
            "count": len(result.get('results', [])),
            "source": "ollama"
        }

        return json.dumps(formatted_response, indent=2)

    except Exception as e:
        logger.error(f"Ollama web search error: {e}")
        return json.dumps({
            "error": f"Ollama web search failed: {e}",
            "query": query,
            "results": []
        })

@tool
def ollama_web_fetch(url: str) -> str:
    """
    Fetch and extract content from a specific web page using Ollama's web fetch API.

    Args:
        url: The URL to fetch content from

    Returns:
        JSON string containing the page title, content, and links
    """
    try:
        logger.info(f"Fetching web page via Ollama: {url}")

        if not OLLAMA_API_KEY:
            return json.dumps({
                "error": "Ollama API key not configured. Please set OLLAMA_API_KEY environment variable.",
                "url": url,
                "title": "Error",
                "content": "",
                "links": []
            })

        # Make direct API call to Ollama web fetch
        headers = {
            "Authorization": f"Bearer {OLLAMA_API_KEY}",
            "Content-Type": "application/json"
        }

        data = {
            "url": url
        }

        response = requests.post(
            f"{OLLAMA_API_BASE}/web_fetch",
            headers=headers,
            json=data,
            timeout=30
        )
        response.raise_for_status()

        result = response.json()

        # Format the response
        formatted_response = {
            "url": url,
            "title": result.get('title', 'No title'),
            "content": result.get('content', ''),
            "links": result.get('links', []),
            "source": "ollama"
        }

        return json.dumps(formatted_response, indent=2)

    except Exception as e:
        logger.error(f"Ollama web fetch error: {e}")
        return json.dumps({
            "error": f"Failed to fetch page via Ollama: {e}",
            "url": url,
            "title": "Error",
            "content": "",
            "links": []
        })

@tool
def search_medical_research(query: str, max_results: int = 3) -> str:
    """
    Search for current medical research and guidelines using Ollama's web search.
    Optimized for finding reliable medical information from authoritative sources.

    Args:
        query: Medical research search query
        max_results: Maximum number of results (default: 3)

    Returns:
        JSON string with medical research results
    """
    try:
        # Enhance the query for medical research
        medical_query = f"{query} medical research evidence clinical guidelines latest studies"

        logger.info(f"Searching medical research for: {query}")

        if not OLLAMA_API_KEY:
            return json.dumps({
                "error": "Ollama API key not configured for medical research search.",
                "query": query,
                "results": []
            })

        # Make direct API call to Ollama web search with medical focus
        headers = {
            "Authorization": f"Bearer {OLLAMA_API_KEY}",
            "Content-Type": "application/json"
        }

        data = {
            "query": medical_query,
            "max_results": max_results
        }

        response = requests.post(
            f"{OLLAMA_API_BASE}/web_search",
            headers=headers,
            json=data,
            timeout=30
        )
        response.raise_for_status()

        result = response.json()

        # Add medical context to results
        results = result.get('results', [])
        for search_result in results:
            search_result['type'] = 'medical_research'
            search_result['enhanced_query'] = medical_query

        formatted_response = {
            "query": query,
            "enhanced_query": medical_query,
            "results": results,
            "count": len(results),
            "type": "medical_research",
            "source": "ollama"
        }

        return json.dumps(formatted_response, indent=2)

    except Exception as e:
        logger.error(f"Medical research search error: {e}")
        return json.dumps({
            "error": f"Medical research search failed: {e}",
            "query": query,
            "results": []
        })

# Export the Ollama-based tools
ollama_web_search_tools = [ollama_web_search, ollama_web_fetch, search_medical_research]

def test_ollama_api_key():
    """Test if Ollama API key is configured and working"""
    if not OLLAMA_API_KEY:
        logger.warning("OLLAMA_API_KEY not set in environment variables")
        return False

    try:
        # Test with a simple search using direct API call
        headers = {
            "Authorization": f"Bearer {OLLAMA_API_KEY}",
            "Content-Type": "application/json"
        }

        data = {
            "query": "test query",
            "max_results": 1
        }

        response = requests.post(
            f"{OLLAMA_API_BASE}/web_search",
            headers=headers,
            json=data,
            timeout=10
        )
        response.raise_for_status()
        logger.info("Ollama API key test successful")
        return True
    except Exception as e:
        logger.error(f"Ollama API key test failed: {e}")
        return False