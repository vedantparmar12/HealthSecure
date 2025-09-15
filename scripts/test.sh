#!/bin/bash

# HealthSecure - Comprehensive Test Runner
# This script runs all tests for the HealthSecure project

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKEND_DIR="$PROJECT_ROOT/backend"
FRONTEND_DIR="$PROJECT_ROOT/frontend"
TESTS_DIR="$PROJECT_ROOT/tests"

# Test results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Helper functions
print_section() {
    echo -e "\n${BLUE}===================================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}===================================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

run_test() {
    local test_name="$1"
    local test_command="$2"
    local test_dir="${3:-$PROJECT_ROOT}"
    
    echo -n "Running $test_name... "
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if (cd "$test_dir" && eval "$test_command") >/dev/null 2>&1; then
        print_success "PASSED"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        print_error "FAILED"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

run_test_verbose() {
    local test_name="$1"
    local test_command="$2"
    local test_dir="${3:-$PROJECT_ROOT}"
    
    echo -e "\n${YELLOW}Running $test_name...${NC}"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if (cd "$test_dir" && eval "$test_command"); then
        print_success "$test_name PASSED"
        PASSED_TESTS=$((PASSED_TESTS + 1))
        return 0
    else
        print_error "$test_name FAILED"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        return 1
    fi
}

check_prerequisites() {
    print_section "Checking Prerequisites"
    
    local missing_deps=()
    
    # Check Go
    if ! command -v go &> /dev/null; then
        missing_deps+=("Go")
    else
        print_success "Go $(go version | cut -d' ' -f3)"
    fi
    
    # Check Node.js
    if ! command -v node &> /dev/null; then
        missing_deps+=("Node.js")
    else
        print_success "Node.js $(node --version)"
    fi
    
    # Check npm
    if ! command -v npm &> /dev/null; then
        missing_deps+=("npm")
    else
        print_success "npm $(npm --version)"
    fi
    
    # Check Docker (optional)
    if command -v docker &> /dev/null; then
        print_success "Docker $(docker --version | cut -d' ' -f3 | cut -d',' -f1)"
    else
        print_warning "Docker not found (optional for integration tests)"
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        print_error "Missing dependencies: ${missing_deps[*]}"
        echo "Please install missing dependencies before running tests."
        exit 1
    fi
}

setup_test_environment() {
    print_section "Setting Up Test Environment"
    
    # Create test database configuration
    if [ ! -f "$PROJECT_ROOT/configs/.env.test" ]; then
        print_info "Creating test environment configuration..."
        cat > "$PROJECT_ROOT/configs/.env.test" << EOF
# Test Environment Configuration
DB_HOST=localhost
DB_PORT=3306
DB_NAME=healthsecure_test
DB_USER=root
DB_PASSWORD=
DB_TLS_MODE=preferred

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=1

JWT_SECRET=test-jwt-secret-minimum-32-characters-for-testing-purposes
JWT_EXPIRES=15m
REFRESH_TOKEN_EXPIRES=7d

BCRYPT_COST=4
ENVIRONMENT=test
LOG_LEVEL=error
SERVER_PORT=8081

ENABLE_CORS=true
CORS_ORIGINS=http://localhost:3001
EOF
        print_success "Test environment configuration created"
    fi
    
    # Install backend dependencies
    if [ -f "$BACKEND_DIR/go.mod" ]; then
        print_info "Installing backend dependencies..."
        (cd "$BACKEND_DIR" && go mod download) || {
            print_error "Failed to install backend dependencies"
            exit 1
        }
        print_success "Backend dependencies installed"
    fi
    
    # Install frontend dependencies
    if [ -f "$FRONTEND_DIR/package.json" ]; then
        print_info "Installing frontend dependencies..."
        (cd "$FRONTEND_DIR" && npm ci --silent) || {
            print_error "Failed to install frontend dependencies"
            exit 1
        }
        print_success "Frontend dependencies installed"
    fi
}

run_backend_tests() {
    print_section "Running Backend Tests"
    
    if [ ! -d "$BACKEND_DIR" ]; then
        print_warning "Backend directory not found, skipping backend tests"
        return 0
    fi
    
    # Unit tests
    print_info "Running Go unit tests..."
    run_test_verbose "Backend Unit Tests" "go test -v ./..." "$BACKEND_DIR"
    
    # Integration tests
    if [ -d "$TESTS_DIR/backend" ]; then
        print_info "Running backend integration tests..."
        run_test_verbose "Backend Integration Tests" "go test -v ./tests/backend/..." "$PROJECT_ROOT"
    fi
    
    # Security tests
    print_info "Running security tests..."
    if command -v gosec &> /dev/null; then
        run_test "Security Scan (gosec)" "gosec -quiet ./..." "$BACKEND_DIR"
    else
        print_warning "gosec not found, install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"
    fi
    
    # Code quality tests
    if command -v golangci-lint &> /dev/null; then
        run_test "Code Quality (golangci-lint)" "golangci-lint run" "$BACKEND_DIR"
    else
        print_warning "golangci-lint not found, skipping code quality checks"
    fi
    
    # Dependency vulnerability scan
    if command -v nancy &> /dev/null; then
        run_test "Dependency Security Scan" "go list -json -m all | nancy sleuth" "$BACKEND_DIR"
    else
        print_warning "nancy not found, skipping dependency security scan"
    fi
}

run_frontend_tests() {
    print_section "Running Frontend Tests"
    
    if [ ! -d "$FRONTEND_DIR" ]; then
        print_warning "Frontend directory not found, skipping frontend tests"
        return 0
    fi
    
    # Unit tests
    print_info "Running React unit tests..."
    run_test_verbose "Frontend Unit Tests" "npm test -- --coverage --watchAll=false" "$FRONTEND_DIR"
    
    # Linting
    print_info "Running ESLint..."
    if [ -f "$FRONTEND_DIR/.eslintrc.js" ] || [ -f "$FRONTEND_DIR/.eslintrc.json" ]; then
        run_test "Frontend Linting" "npm run lint" "$FRONTEND_DIR"
    else
        print_warning "ESLint configuration not found, skipping linting"
    fi
    
    # Security audit
    print_info "Running npm security audit..."
    run_test "NPM Security Audit" "npm audit --audit-level=high" "$FRONTEND_DIR"
    
    # Build test
    print_info "Testing production build..."
    run_test "Production Build" "npm run build" "$FRONTEND_DIR"
}

run_database_tests() {
    print_section "Running Database Tests"
    
    # Check if MySQL is available
    if ! command -v mysql &> /dev/null; then
        print_warning "MySQL not found, skipping database tests"
        return 0
    fi
    
    # Test database connection
    print_info "Testing database connection..."
    run_test "Database Connection" "mysql -e 'SELECT 1' >/dev/null 2>&1"
    
    # Test schema validation
    if [ -f "$PROJECT_ROOT/database/schema.sql" ]; then
        print_info "Validating database schema..."
        run_test "Schema Validation" "mysql < database/schema.sql" "$PROJECT_ROOT"
    fi
}

run_api_tests() {
    print_section "Running API Tests"
    
    # Check if server is running or start it for tests
    if ! curl -f http://localhost:8080/health >/dev/null 2>&1; then
        print_info "Starting test server..."
        (cd "$BACKEND_DIR" && go run cmd/server/main.go &)
        SERVER_PID=$!
        sleep 5  # Wait for server to start
        
        # Register cleanup function
        cleanup_server() {
            if [ ! -z "$SERVER_PID" ]; then
                kill $SERVER_PID 2>/dev/null || true
            fi
        }
        trap cleanup_server EXIT
    fi
    
    # Health check test
    run_test "API Health Check" "curl -f http://localhost:8080/health"
    
    # Authentication tests
    run_test "Authentication Endpoint" "curl -f -X POST http://localhost:8080/api/auth/login -H 'Content-Type: application/json' -d '{\"email\":\"test@example.com\",\"password\":\"password\"}' | grep -q 'error\\|token'"
    
    # CORS headers test
    run_test "CORS Headers" "curl -H 'Origin: http://localhost:3000' -I http://localhost:8080/api/auth/login | grep -q 'Access-Control-Allow-Origin'"
}

run_security_tests() {
    print_section "Running Security Tests"
    
    # SSL/TLS configuration test (if certificates exist)
    if [ -f "/etc/ssl/certs/healthsecure.crt" ]; then
        run_test "SSL Certificate Validity" "openssl x509 -in /etc/ssl/certs/healthsecure.crt -checkend 86400 -noout"
    fi
    
    # Environment variable security
    print_info "Checking environment variable security..."
    if grep -r "password.*=" "$PROJECT_ROOT/configs/" | grep -v ".example" | grep -v ".template" >/dev/null 2>&1; then
        print_error "Plain text passwords found in configuration files"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    else
        print_success "No plain text passwords in configuration"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    fi
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    # JWT secret strength
    if [ -f "$PROJECT_ROOT/configs/.env" ]; then
        JWT_SECRET=$(grep "JWT_SECRET" "$PROJECT_ROOT/configs/.env" | cut -d'=' -f2)
        if [ ${#JWT_SECRET} -lt 32 ]; then
            print_error "JWT secret is too short (minimum 32 characters)"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        else
            print_success "JWT secret meets minimum length requirement"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        fi
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
    fi
}

run_docker_tests() {
    print_section "Running Docker Tests"
    
    if ! command -v docker &> /dev/null; then
        print_warning "Docker not found, skipping Docker tests"
        return 0
    fi
    
    # Docker build tests
    if [ -f "$PROJECT_ROOT/docker/Dockerfile.backend" ]; then
        run_test "Backend Docker Build" "docker build -f docker/Dockerfile.backend -t healthsecure-backend-test ."
        
        # Cleanup test image
        docker rmi healthsecure-backend-test >/dev/null 2>&1 || true
    fi
    
    if [ -f "$PROJECT_ROOT/docker/Dockerfile.frontend" ]; then
        run_test "Frontend Docker Build" "docker build -f docker/Dockerfile.frontend -t healthsecure-frontend-test ."
        
        # Cleanup test image
        docker rmi healthsecure-frontend-test >/dev/null 2>&1 || true
    fi
    
    # Docker Compose validation
    if [ -f "$PROJECT_ROOT/docker-compose.yml" ]; then
        run_test "Docker Compose Validation" "docker-compose config -q"
    fi
}

run_performance_tests() {
    print_section "Running Performance Tests"
    
    # Check if Apache Bench is available
    if command -v ab &> /dev/null; then
        print_info "Running basic load test..."
        run_test "API Load Test (100 requests)" "ab -n 100 -c 10 -q http://localhost:8080/health"
    else
        print_warning "Apache Bench (ab) not found, skipping load tests"
    fi
    
    # Frontend bundle size check
    if [ -f "$FRONTEND_DIR/build/static/js/*.js" ]; then
        BUNDLE_SIZE=$(du -sb "$FRONTEND_DIR/build/static/js/"*.js | awk '{sum += $1} END {print sum}')
        MAX_BUNDLE_SIZE=$((2 * 1024 * 1024))  # 2MB
        
        if [ "$BUNDLE_SIZE" -gt "$MAX_BUNDLE_SIZE" ]; then
            print_error "Frontend bundle size exceeds 2MB ($BUNDLE_SIZE bytes)"
            FAILED_TESTS=$((FAILED_TESTS + 1))
        else
            print_success "Frontend bundle size is acceptable ($BUNDLE_SIZE bytes)"
            PASSED_TESTS=$((PASSED_TESTS + 1))
        fi
        TOTAL_TESTS=$((TOTAL_TESTS + 1))
    fi
}

generate_test_report() {
    print_section "Test Results Summary"
    
    echo -e "Total Tests: $TOTAL_TESTS"
    echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
    echo -e "${RED}Failed: $FAILED_TESTS${NC}"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "\n${GREEN}🎉 All tests passed! HealthSecure is ready for deployment.${NC}"
        return 0
    else
        echo -e "\n${RED}❌ Some tests failed. Please review the output above and fix the issues.${NC}"
        return 1
    fi
}

# Main execution
main() {
    echo -e "${BLUE}HealthSecure Test Runner${NC}"
    echo -e "${BLUE}========================${NC}\n"
    
    # Parse command line arguments
    local run_all=true
    local run_backend=false
    local run_frontend=false
    local run_security=false
    local run_docker=false
    local run_api=false
    local verbose=false
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --backend)
                run_all=false
                run_backend=true
                shift
                ;;
            --frontend)
                run_all=false
                run_frontend=true
                shift
                ;;
            --security)
                run_all=false
                run_security=true
                shift
                ;;
            --docker)
                run_all=false
                run_docker=true
                shift
                ;;
            --api)
                run_all=false
                run_api=true
                shift
                ;;
            --verbose|-v)
                verbose=true
                shift
                ;;
            --help|-h)
                echo "Usage: $0 [OPTIONS]"
                echo ""
                echo "Options:"
                echo "  --backend     Run only backend tests"
                echo "  --frontend    Run only frontend tests"
                echo "  --security    Run only security tests"
                echo "  --docker      Run only Docker tests"
                echo "  --api         Run only API tests"
                echo "  --verbose,-v  Verbose output"
                echo "  --help,-h     Show this help message"
                echo ""
                echo "If no specific test type is specified, all tests will be run."
                exit 0
                ;;
            *)
                echo "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    # Run tests based on arguments
    check_prerequisites
    setup_test_environment
    
    if [ "$run_all" = true ] || [ "$run_backend" = true ]; then
        run_backend_tests
    fi
    
    if [ "$run_all" = true ] || [ "$run_frontend" = true ]; then
        run_frontend_tests
    fi
    
    if [ "$run_all" = true ]; then
        run_database_tests
    fi
    
    if [ "$run_all" = true ] || [ "$run_api" = true ]; then
        run_api_tests
    fi
    
    if [ "$run_all" = true ] || [ "$run_security" = true ]; then
        run_security_tests
    fi
    
    if [ "$run_all" = true ] || [ "$run_docker" = true ]; then
        run_docker_tests
    fi
    
    if [ "$run_all" = true ]; then
        run_performance_tests
    fi
    
    # Generate final report
    generate_test_report
}

# Run main function with all arguments
main "$@"
