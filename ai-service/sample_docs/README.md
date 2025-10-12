# Sample Medical Documents

Place your medical PDF documents here for testing the RAG system.

## Supported Formats

- PDF documents (with or without OCR)
- DOCX files
- PPTX presentations
- Images (PNG, JPG) - will be processed with OCR

## Example Documents to Test

1. **Patient Reports**: Lab results, vital signs, medical history
2. **Clinical Notes**: Doctor's notes, nursing notes, consultation reports
3. **Medical Images**: X-rays, scans, charts (as images)
4. **Research Papers**: Medical research, clinical guidelines

## How to Ingest Documents

### Single Document Test
```bash
python ingest_with_docling.py --test-single "sample_docs/patient_report.pdf"
```

### Batch Processing
```bash
python ingest_with_docling.py --documents sample_docs
```

### Custom File Patterns
```bash
python ingest_with_docling.py --documents sample_docs --patterns "*.pdf" "*.docx"
```

## What Docling Extracts

- **Text Content**: With OCR for scanned documents
- **Tables**: Structured table data with cell relationships
- **Images**: Charts, diagrams, medical images
- **Metadata**: Title, authors, dates
- **Document Structure**: Sections, headings, lists

## Features

✅ **OCR Support**: Reads scanned PDFs automatically
✅ **Smart Chunking**: HybridChunker for optimal chunk sizes
✅ **Table Extraction**: Preserves complex medical tables
✅ **Batch Processing**: Process multiple documents efficiently
✅ **Vector Search**: Semantic similarity search via Qdrant
