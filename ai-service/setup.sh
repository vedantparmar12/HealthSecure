#!/bin/bash

# HealthSecure AI Service Setup Script

echo "🚀 Setting up HealthSecure AI Service..."

# Check if Python 3.11+ is installed
if ! command -v python3 &> /dev/null; then
    echo "❌ Python 3 is not installed. Please install Python 3.11 or higher."
    exit 1
fi

PYTHON_VERSION=$(python3 --version | cut -d' ' -f2)
echo "✅ Python version: $PYTHON_VERSION"

# Create virtual environment
echo "📦 Creating virtual environment..."
python3 -m venv venv

# Activate virtual environment
source venv/bin/activate || source venv/Scripts/activate

# Upgrade pip
echo "⬆️ Upgrading pip..."
pip install --upgrade pip

# Install requirements
echo "📋 Installing Python dependencies..."
pip install -r requirements.txt

# Check if .env file exists
if [ ! -f "../configs/.env" ]; then
    echo "❌ .env file not found at ../configs/.env"
    echo "Please ensure your Go backend .env file exists with:"
    echo "  - OPENROUTER_API_KEY"
    echo "  - LANGCHAIN_API_KEY" 
    echo "  - LANGCHAIN_PROJECT"
    exit 1
fi

echo "✅ Environment file found"

# Test the service
echo "🧪 Testing the AI service..."
python -c "
import os
import sys
sys.path.append('.')
from app import Config

if not Config.OPENROUTER_API_KEY:
    print('❌ OPENROUTER_API_KEY not found in environment')
    sys.exit(1)
    
print('✅ OpenRouter API key configured')

if Config.LANGCHAIN_TRACING_V2:
    print('✅ LangSmith tracing enabled')
else:
    print('⚠️ LangSmith tracing disabled')

print('🎯 Configuration validated successfully!')
"

if [ $? -eq 0 ]; then
    echo ""
    echo "🎉 HealthSecure AI Service setup complete!"
    echo ""
    echo "To start the service:"
    echo "  1. Activate virtual environment: source venv/bin/activate"
    echo "  2. Run the service: python app.py"
    echo ""
    echo "The service will be available at: http://localhost:5000"
    echo "Health check: curl http://localhost:5000/health"
else
    echo "❌ Setup failed. Please check the error messages above."
    exit 1
fi