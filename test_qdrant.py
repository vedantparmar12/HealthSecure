#!/usr/bin/env python3
"""
Test script for Qdrant integration with HealthSecure
"""

import asyncio
import os
import sys

# Add ai-service to path
sys.path.append(os.path.join(os.path.dirname(__file__), 'ai-service'))

from pdf_rag_analyzer import PDFRAGAnalyzer

async def test_qdrant_integration():
    """Test Qdrant vector database integration."""
    print("🧪 Testing Qdrant Integration with HealthSecure")
    print("=" * 50)

    try:
        # Initialize the PDF RAG analyzer with Qdrant
        analyzer = PDFRAGAnalyzer()

        print("✅ PDFRAGAnalyzer initialized")

        if analyzer.qdrant_client:
            print("✅ Qdrant client connected successfully")
            print(f"📍 Qdrant URL: {analyzer.qdrant_url}")
            print(f"📦 Collection: {analyzer.collection_name}")
            print(f"📏 Vector size: {analyzer.vector_size}")

            # Test collection info
            try:
                collections = analyzer.qdrant_client.get_collections()
                print(f"📋 Available collections: {[col.name for col in collections.collections]}")

                # Get collection info
                if analyzer.collection_name in [col.name for col in collections.collections]:
                    collection_info = analyzer.qdrant_client.get_collection(analyzer.collection_name)
                    print(f"📊 Collection '{analyzer.collection_name}' status: {collection_info.status}")
                    print(f"🔢 Points count: {collection_info.points_count}")
                else:
                    print(f"❌ Collection '{analyzer.collection_name}' not found")

            except Exception as e:
                print(f"⚠️ Could not get collection info: {e}")

            # Test embedding generation
            try:
                print("\n🔍 Testing embedding generation...")
                test_query = "patient vital signs blood pressure"
                search_results = await analyzer.search_similar_chunks(test_query, limit=3)

                if search_results:
                    print(f"✅ Found {len(search_results)} similar chunks for query: '{test_query}'")
                    for i, result in enumerate(search_results, 1):
                        print(f"   {i}. Similarity: {result['similarity']:.3f}")
                        print(f"      Content preview: {result['content'][:100]}...")
                else:
                    print("ℹ️ No existing chunks found (empty database)")

            except Exception as e:
                print(f"⚠️ Search test failed: {e}")

        else:
            print("❌ Qdrant client not available")
            print("   Make sure Qdrant is running on http://localhost:6333")

    except Exception as e:
        print(f"❌ Test failed: {e}")
        import traceback
        traceback.print_exc()

    print("\n" + "=" * 50)
    print("Test completed!")

if __name__ == "__main__":
    asyncio.run(test_qdrant_integration())