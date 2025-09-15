#!/bin/bash

# HealthSecure Setup Script
# This script sets up the HealthSecure application for development and production

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker and Docker Compose are installed
check_dependencies() {
    print_status "Checking dependencies..."
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    print_success "Docker and Docker Compose are installed"
}

# Create environment file
create_env_file() {
    print_status "Setting up environment configuration..."
    
    if [ ! -f "configs/.env" ]; then
        print_status "Creating .env file from template..."
        cp configs/.env.example configs/.env
        
        # Generate JWT secret
        JWT_SECRET=$(openssl rand -hex 32)
        sed -i "s/super-secure-jwt-secret-minimum-32-characters-change-in-production/$JWT_SECRET/" configs/.env
        
        print_warning "Please review and update configs/.env with your production values"
        print_warning "Especially change default passwords and secrets!"
    else
        print_status ".env file already exists"
    fi
}

# Build Docker images
build_images() {
    print_status "Building Docker images..."
    
    print_status "Building backend image..."
    docker build -t healthsecure-backend -f docker/Dockerfile.backend ./backend
    
    print_status "Building frontend image..."
    if [ -f "docker/Dockerfile.frontend" ]; then
        docker build -t healthsecure-frontend -f docker/Dockerfile.frontend ./frontend
    else
        print_warning "Frontend Dockerfile not found, skipping frontend build"
    fi
    
    print_success "Docker images built successfully"
}

# Initialize database
init_database() {
    print_status "Starting database services..."
    
    # Start MySQL and Redis
    docker-compose up -d mysql redis
    
    print_status "Waiting for database to be ready..."
    sleep 30
    
    # Check if database is healthy
    while ! docker-compose exec mysql mysqladmin ping -h localhost --silent; do
        print_status "Waiting for MySQL to be ready..."
        sleep 5
    done
    
    print_success "Database is ready"
}

# Start all services
start_services() {
    print_status "Starting all services..."
    
    docker-compose up -d
    
    print_status "Waiting for services to be healthy..."
    sleep 20
    
    # Check service health
    if curl -f http://localhost:8080/health > /dev/null 2>&1; then
        print_success "Backend service is healthy"
    else
        print_warning "Backend service health check failed"
    fi
    
    print_success "HealthSecure application is running!"
}

# Display service URLs
display_urls() {
    print_success "HealthSecure is now running at:"
    echo "  Frontend:     http://localhost:3000"
    echo "  Backend API:  http://localhost:8080"
    echo "  Health Check: http://localhost:8080/health"
    echo ""
    echo "Default Admin Credentials (CHANGE IMMEDIATELY):"
    echo "  Email: admin@healthsecure.local"
    echo "  Password: password123"
    echo ""
    print_warning "Remember to:"
    echo "  1. Change default passwords"
    echo "  2. Update JWT secrets"
    echo "  3. Configure OAuth providers"
    echo "  4. Set up SSL certificates for production"
    echo "  5. Review audit and security settings"
}

# Main setup function
main() {
    echo "======================================"
    echo "  HealthSecure Setup Script"
    echo "  HIPAA-Compliant Medical Data System"
    echo "======================================"
    echo ""
    
    check_dependencies
    create_env_file
    build_images
    init_database
    start_services
    display_urls
    
    print_success "Setup completed successfully!"
}

# Run main function
main "$@"