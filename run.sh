#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# --- HealthSecure Unified Runner ---
# This script installs dependencies and runs the backend, frontend, and AI services.

# Function to activate virtual environment based on OS
activate_venv() {
  if [ -f ".venv/bin/activate" ]; then
    # Unix-like systems
    source .venv/bin/activate
  elif [ -f ".venv/Scripts/activate" ]; then
    # Windows
    source .venv/Scripts/activate
  else
    echo "Error: Could not find activation script in .venv"
    exit 1
  fi
}

# 1. Install Backend (Go) Dependencies
echo "--- Installing Backend Dependencies ---"
cd backend
go mod tidy
cd ..

# 2. Install AI Service (Python) Dependencies
echo "--- Installing AI Service Dependencies ---"
cd ai-service
# Creating a virtual environment is recommended
if [ ! -d ".venv" ]; then
  python -m venv .venv
fi
# Activate virtual environment and install requirements
activate_venv
pip install -r requirements.txt
deactivate
cd ..

# 3. Install Frontend (React) Dependencies
echo "--- Installing Frontend Dependencies ---"
cd frontend
npm install
cd ..

# 4. Run Services
echo "--- Starting All Services ---"

# Run AI Service in the background
echo "Starting AI Service..."
cd ai-service
activate_venv
python app.py &
AI_PID=$!
cd ..

# Run Backend Service in the background
echo "Starting Backend Service..."
cd backend
go run cmd/server/main.go &
BACKEND_PID=$!
cd ..

# Run Frontend Service in the foreground
echo "Starting Frontend Service..."
cd frontend
npm start

# Clean up background processes on exit
kill $AI_PID
kill $BACKEND_PID