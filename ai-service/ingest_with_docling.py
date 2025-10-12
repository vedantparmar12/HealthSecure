#!/usr/bin/env python3
"""
Document Ingestion Script using Docling RAG Analyzer
Batch processes medical documents and stores them in Qdrant.
"""

import asyncio
import argparse
import logging
from pathlib import Path
from docling_rag_analyzer import DoclingRAGAnalyzer

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


async def main():
    """Main ingestion workflow."""
    parser = argparse.ArgumentParser(
        description="Ingest medical documents using Docling for RAG"
    )
    parser.add_argument(
        "--documents", "-d",
        required=True,
        help="Path to documents directory"
    )
    parser.add_argument(
        "--patterns", "-p",
        nargs="+",
        default=["*.pdf", "*.docx", "*.pptx", "*.png", "*.jpg"],
        help="File patterns to process (default: pdf, docx, pptx, png, jpg)"
    )
    parser.add_argument(
        "--test-single", "-t",
        help="Test with a single document file"
    )

    args = parser.parse_args()

    # Initialize the Docling RAG analyzer
    logger.info("Initializing Docling RAG Analyzer...")
    analyzer = DoclingRAGAnalyzer()

    # Test Qdrant connection
    if not analyzer.qdrant_client:
        logger.error("❌ Qdrant is not running!")
        logger.error("Please start Qdrant first:")
        logger.error("  Windows: run setup_qdrant.bat")
        logger.error("  Linux/Mac: docker run -p 6333:6333 qdrant/qdrant")
        return

    logger.info("✅ Qdrant connection established")

    # Test with single document if specified
    if args.test_single:
        logger.info(f"Testing with single document: {args.test_single}")
        try:
            result = await analyzer.analyze_pdf(args.test_single)
            logger.info("✅ Document analysis completed!")
            logger.info(f"Chunks created: {len(result['rag_chunks'])}")
            logger.info(f"Tables extracted: {len(result['structured_data'].get('tables', []))}")
            logger.info(f"Images found: {len(result['structured_data'].get('images', []))}")
            
            # Test search
            logger.info("\nTesting search functionality...")
            search_results = await analyzer.search_similar_chunks("patient vital signs", limit=3)
            logger.info(f"Found {len(search_results)} similar chunks")
            for i, result in enumerate(search_results, 1):
                logger.info(f"\n{i}. Similarity: {result['similarity']:.3f}")
                logger.info(f"   Content: {result['content'][:200]}...")
            
            return
        except Exception as e:
            logger.error(f"❌ Failed to process document: {e}")
            import traceback
            traceback.print_exc()
            return

    # Batch process directory
    logger.info(f"Starting batch ingestion from: {args.documents}")
    logger.info(f"File patterns: {', '.join(args.patterns)}")

    try:
        results = await analyzer.batch_process_documents(
            args.documents,
            file_patterns=args.patterns
        )

        # Print summary
        logger.info("\n" + "=" * 60)
        logger.info("INGESTION SUMMARY")
        logger.info("=" * 60)
        logger.info(f"Total files: {results['total_files']}")
        logger.info(f"Successfully processed: {len(results['processed'])}")
        logger.info(f"Failed: {len(results['failed'])}")

        if results['processed']:
            logger.info("\nProcessed files:")
            for item in results['processed']:
                logger.info(f"  ✅ {item['file']}: {item['chunks']} chunks")

        if results['failed']:
            logger.info("\nFailed files:")
            for item in results['failed']:
                logger.info(f"  ❌ {item['file']}: {item['error']}")

        # Test search with ingested documents
        logger.info("\n" + "=" * 60)
        logger.info("Testing RAG search with ingested documents...")
        logger.info("=" * 60)
        
        test_queries = [
            "patient vital signs",
            "blood pressure measurements",
            "laboratory test results"
        ]

        for query in test_queries:
            logger.info(f"\nQuery: '{query}'")
            search_results = await analyzer.search_similar_chunks(query, limit=3)
            
            if search_results:
                logger.info(f"Found {len(search_results)} relevant chunks:")
                for i, result in enumerate(search_results, 1):
                    logger.info(f"  {i}. Similarity: {result['similarity']:.3f}")
                    logger.info(f"     Source: {result['metadata'].get('source_file', 'Unknown')}")
                    logger.info(f"     Preview: {result['content'][:150]}...")
            else:
                logger.info("  No results found")

    except Exception as e:
        logger.error(f"❌ Batch processing failed: {e}")
        import traceback
        traceback.print_exc()


if __name__ == "__main__":
    print("""
    ╔═══════════════════════════════════════════════════════════╗
    ║   HealthSecure Document Ingestion with Docling & Qdrant  ║
    ║   Advanced OCR, Smart Chunking, Batch Processing         ║
    ╚═══════════════════════════════════════════════════════════╝
    """)
    
    asyncio.run(main())
