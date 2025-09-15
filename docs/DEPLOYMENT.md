# Deployment Guide

## Overview

This guide provides comprehensive instructions for deploying HealthSecure in various environments, from development to production.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Development Environment](#development-environment)
3. [Production Deployment](#production-deployment)
4. [Docker Deployment](#docker-deployment)
5. [Database Setup](#database-setup)
6. [SSL/TLS Configuration](#ssltls-configuration)
7. [Monitoring Setup](#monitoring-setup)
8. [Backup and Recovery](#backup-and-recovery)
9. [Troubleshooting](#troubleshooting)

## Prerequisites

### System Requirements

**Minimum Requirements**:
- CPU: 2 cores
- RAM: 4GB
- Storage: 50GB
- Operating System: Linux (Ubuntu 20.04+ recommended)

**Recommended for Production**:
- CPU: 4+ cores
- RAM: 8GB+
- Storage: 100GB+ SSD
- Load balancer for high availability

### Software Dependencies

```bash
# Required software versions
Go 1.21+
Node.js 18+
MySQL 8.0+
Redis 6.0+
Docker 20.10+ (optional)
Nginx 1.18+ (for reverse proxy)
```

### Network Requirements

```yaml
Ports:
  - 80: HTTP (redirects to HTTPS)
  - 443: HTTPS
  - 3306: MySQL (internal only)
  - 6379: Redis (internal only)
  - 8080: Backend API (internal only)
  - 3000: Frontend dev server (development only)
```

## Development Environment

### Quick Start

1. **Clone the repository**:
```bash
git clone https://github.com/vedantparmar12/HealthSecure.git
cd healthsecure
```

2. **Set up environment variables**:
```bash
cp configs/.env.example configs/.env
# Edit configs/.env with your development settings
```

3. **Start services with Docker Compose**:
```bash
docker-compose up -d mysql redis
```

4. **Run database migrations**:
```bash
cd backend
go run cmd/migrate/main.go
```

5. **Start backend server**:
```bash
cd backend
go run cmd/server/main.go
```

6. **Start frontend development server**:
```bash
cd frontend
npm install
npm start
```

### Development Environment Variables

```bash
# configs/.env for development
DB_HOST=localhost
DB_PORT=3306
DB_NAME=healthsecure_dev
DB_USER=healthsecure_user
DB_PASSWORD=dev_password
DB_TLS_MODE=preferred

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

JWT_SECRET=development-secret-key-change-in-production-minimum-32-chars
JWT_EXPIRES=15m
REFRESH_TOKEN_EXPIRES=7d

BCRYPT_COST=12
ENVIRONMENT=development
LOG_LEVEL=debug
SERVER_PORT=8080

ENABLE_CORS=true
CORS_ORIGINS=http://localhost:3000
```

## Production Deployment

### Server Setup

1. **Update system packages**:
```bash
sudo apt update && sudo apt upgrade -y
```

2. **Install required packages**:
```bash
# Install Go
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Install Node.js
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# Install MySQL
sudo apt install mysql-server-8.0 -y
sudo mysql_secure_installation

# Install Redis
sudo apt install redis-server -y

# Install Nginx
sudo apt install nginx -y
```

3. **Create application user**:
```bash
sudo useradd -m -s /bin/bash healthsecure
sudo mkdir -p /opt/healthsecure
sudo chown healthsecure:healthsecure /opt/healthsecure
```

4. **Deploy application**:
```bash
# Copy application files
sudo -u healthsecure cp -r . /opt/healthsecure/

# Build backend
cd /opt/healthsecure/backend
sudo -u healthsecure go build -o bin/server cmd/server/main.go
sudo -u healthsecure go build -o bin/migrate cmd/migrate/main.go

# Build frontend
cd /opt/healthsecure/frontend
sudo -u healthsecure npm install
sudo -u healthsecure npm run build
```

### Production Environment Variables

```bash
# /opt/healthsecure/configs/.env
DB_HOST=localhost
DB_PORT=3306
DB_NAME=healthsecure
DB_USER=healthsecure_user
DB_PASSWORD=STRONG_RANDOM_PASSWORD
DB_TLS_MODE=required

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=STRONG_REDIS_PASSWORD
REDIS_DB=0

JWT_SECRET=VERY_STRONG_RANDOM_JWT_SECRET_MINIMUM_32_CHARACTERS
JWT_EXPIRES=15m
REFRESH_TOKEN_EXPIRES=7d

BCRYPT_COST=12
ENVIRONMENT=production
LOG_LEVEL=info
SERVER_PORT=8080

ENABLE_CORS=true
CORS_ORIGINS=https://yourdomain.com

SSL_CERT_PATH=/etc/ssl/certs/healthsecure.crt
SSL_KEY_PATH=/etc/ssl/private/healthsecure.key

EMERGENCY_ACCESS_DURATION=1h
EMERGENCY_NOTIFICATION_EMAIL=security@yourorg.com
```

### Systemd Services

1. **Backend service**:
```ini
# /etc/systemd/system/healthsecure-backend.service
[Unit]
Description=HealthSecure Backend API
After=network.target mysql.service redis.service

[Service]
Type=simple
User=healthsecure
Group=healthsecure
WorkingDirectory=/opt/healthsecure/backend
ExecStart=/opt/healthsecure/backend/bin/server
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=healthsecure-backend
Environment=ENV_FILE=/opt/healthsecure/configs/.env

[Install]
WantedBy=multi-user.target
```

2. **Enable and start services**:
```bash
sudo systemctl daemon-reload
sudo systemctl enable healthsecure-backend
sudo systemctl start healthsecure-backend
```

## Docker Deployment

### Production Docker Compose

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  mysql:
    image: mysql:8.0
    container_name: healthsecure-mysql
    environment:
      MYSQL_ROOT_PASSWORD: ${DB_ROOT_PASSWORD}
      MYSQL_DATABASE: ${DB_NAME}
      MYSQL_USER: ${DB_USER}
      MYSQL_PASSWORD: ${DB_PASSWORD}
    volumes:
      - mysql_data:/var/lib/mysql
      - ./database/schema.sql:/docker-entrypoint-initdb.d/01-schema.sql
      - ./database/seed_data.sql:/docker-entrypoint-initdb.d/02-seed.sql
    ports:
      - "3306:3306"
    command: --default-authentication-plugin=mysql_native_password --require-secure-transport=ON
    networks:
      - healthsecure-network
    restart: unless-stopped

  redis:
    image: redis:6-alpine
    container_name: healthsecure-redis
    command: redis-server --requirepass ${REDIS_PASSWORD} --appendonly yes
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    networks:
      - healthsecure-network
    restart: unless-stopped

  backend:
    build:
      context: ./backend
      dockerfile: ../docker/Dockerfile.backend
    container_name: healthsecure-backend
    environment:
      - DB_HOST=mysql
      - REDIS_HOST=redis
    env_file:
      - ./configs/.env
    depends_on:
      - mysql
      - redis
    ports:
      - "8080:8080"
    networks:
      - healthsecure-network
    volumes:
      - ./logs:/app/logs
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  frontend:
    build:
      context: ./frontend
      dockerfile: ../docker/Dockerfile.frontend
    container_name: healthsecure-frontend
    volumes:
      - frontend_build:/app/build
    networks:
      - healthsecure-network

  nginx:
    image: nginx:alpine
    container_name: healthsecure-nginx
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./configs/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl/certs:/etc/nginx/ssl:ro
      - frontend_build:/usr/share/nginx/html
    depends_on:
      - backend
      - frontend
    networks:
      - healthsecure-network
    restart: unless-stopped

volumes:
  mysql_data:
    driver: local
  redis_data:
    driver: local
  frontend_build:
    driver: local

networks:
  healthsecure-network:
    driver: bridge
```

### Deploy with Docker

```bash
# Create production environment file
cp configs/.env.example configs/.env.prod
# Edit configs/.env.prod with production values

# Deploy
docker-compose -f docker-compose.prod.yml --env-file configs/.env.prod up -d

# Check status
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f backend
```

## Database Setup

### MySQL Configuration

1. **Create database and user**:
```sql
-- Connect as root
mysql -u root -p

-- Create database
CREATE DATABASE healthsecure CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create application user
CREATE USER 'healthsecure_user'@'localhost' IDENTIFIED BY 'STRONG_PASSWORD';
GRANT SELECT, INSERT, UPDATE, DELETE ON healthsecure.* TO 'healthsecure_user'@'localhost';
FLUSH PRIVILEGES;
```

2. **Configure MySQL for HIPAA compliance**:
```ini
# /etc/mysql/mysql.conf.d/healthsecure.cnf
[mysqld]
# Enable SSL
ssl-ca=/etc/mysql/ssl/ca.pem
ssl-cert=/etc/mysql/ssl/server-cert.pem
ssl-key=/etc/mysql/ssl/server-key.pem
require_secure_transport=ON

# Enable binary logging for point-in-time recovery
log-bin=mysql-bin
binlog-format=ROW
sync_binlog=1

# Enable InnoDB encryption
innodb_encrypt_tables=ON
innodb_encrypt_log=ON
innodb_encrypt_online_alter_logs=ON

# Performance optimizations
innodb_buffer_pool_size=2G
innodb_log_file_size=512M
max_connections=200
```

3. **Run migrations**:
```bash
cd /opt/healthsecure/backend
./bin/migrate
```

### Database Backup

```bash
#!/bin/bash
# /opt/healthsecure/scripts/backup-database.sh

BACKUP_DIR="/opt/healthsecure/backups"
DATE=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="healthsecure_backup_${DATE}.sql.gz"

# Create backup directory
mkdir -p $BACKUP_DIR

# Perform backup
mysqldump --single-transaction --routines --triggers \
  --user=healthsecure_user --password=$DB_PASSWORD \
  --ssl-mode=REQUIRED \
  healthsecure | gzip > $BACKUP_DIR/$BACKUP_FILE

# Keep only last 30 days of backups
find $BACKUP_DIR -name "healthsecure_backup_*.sql.gz" -mtime +30 -delete

echo "Database backup completed: $BACKUP_FILE"
```

## SSL/TLS Configuration

### Obtain SSL Certificate

1. **Using Let's Encrypt**:
```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx -y

# Obtain certificate
sudo certbot certonly --nginx -d yourdomain.com

# Set up auto-renewal
sudo crontab -e
# Add: 0 12 * * * /usr/bin/certbot renew --quiet
```

2. **Using custom certificate**:
```bash
# Copy certificates
sudo mkdir -p /etc/ssl/certs /etc/ssl/private
sudo cp your-certificate.crt /etc/ssl/certs/healthsecure.crt
sudo cp your-private-key.key /etc/ssl/private/healthsecure.key
sudo chmod 600 /etc/ssl/private/healthsecure.key
```

### Nginx Configuration

```nginx
# /etc/nginx/sites-available/healthsecure
server {
    listen 80;
    server_name yourdomain.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com;

    # SSL Configuration
    ssl_certificate /etc/ssl/certs/healthsecure.crt;
    ssl_certificate_key /etc/ssl/private/healthsecure.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers on;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;

    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-Frame-Options "DENY" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Frontend
    location / {
        root /opt/healthsecure/frontend/build;
        index index.html;
        try_files $uri $uri/ /index.html;
        
        # Cache static assets
        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }
    }

    # API Proxy
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
        proxy_read_timeout 300s;
        proxy_connect_timeout 75s;
    }

    # Health check
    location /health {
        proxy_pass http://localhost:8080/health;
        access_log off;
    }
}
```

## Monitoring Setup

### Application Monitoring

1. **Prometheus configuration**:
```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'healthsecure'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 30s
```

2. **Grafana dashboard**:
```json
{
  "dashboard": {
    "title": "HealthSecure Monitoring",
    "panels": [
      {
        "title": "API Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))"
          }
        ]
      },
      {
        "title": "Authentication Failures",
        "type": "stat",
        "targets": [
          {
            "expr": "increase(healthsecure_auth_failures_total[1h])"
          }
        ]
      }
    ]
  }
}
```

### Log Management

```bash
# /etc/rsyslog.d/healthsecure.conf
# HealthSecure application logs
:programname, isequal, "healthsecure-backend" /var/log/healthsecure/backend.log
& stop

# Log rotation
# /etc/logrotate.d/healthsecure
/var/log/healthsecure/*.log {
    daily
    rotate 365
    missingok
    compress
    delaycompress
    notifempty
    copytruncate
    postrotate
        systemctl reload rsyslog > /dev/null 2>&1 || true
    endscript
}
```

## Backup and Recovery

### Automated Backup Script

```bash
#!/bin/bash
# /opt/healthsecure/scripts/backup-all.sh

set -e

BACKUP_DIR="/opt/healthsecure/backups"
DATE=$(date +"%Y%m%d_%H%M%S")
RETENTION_DAYS=30

echo "Starting backup process..."

# Database backup
echo "Backing up database..."
mysqldump --single-transaction --routines --triggers \
  --user=$DB_USER --password=$DB_PASSWORD \
  --ssl-mode=REQUIRED \
  $DB_NAME | gzip > $BACKUP_DIR/database_${DATE}.sql.gz

# Configuration backup
echo "Backing up configuration..."
tar -czf $BACKUP_DIR/config_${DATE}.tar.gz /opt/healthsecure/configs/

# Application logs backup
echo "Backing up logs..."
tar -czf $BACKUP_DIR/logs_${DATE}.tar.gz /var/log/healthsecure/

# Clean old backups
echo "Cleaning old backups..."
find $BACKUP_DIR -name "*_*.sql.gz" -mtime +$RETENTION_DAYS -delete
find $BACKUP_DIR -name "*_*.tar.gz" -mtime +$RETENTION_DAYS -delete

echo "Backup process completed successfully"

# Upload to remote storage (optional)
if [ ! -z "$BACKUP_S3_BUCKET" ]; then
    echo "Uploading to S3..."
    aws s3 cp $BACKUP_DIR/database_${DATE}.sql.gz s3://$BACKUP_S3_BUCKET/healthsecure/
    aws s3 cp $BACKUP_DIR/config_${DATE}.tar.gz s3://$BACKUP_S3_BUCKET/healthsecure/
fi
```

### Recovery Procedure

```bash
#!/bin/bash
# /opt/healthsecure/scripts/restore-database.sh

BACKUP_FILE=$1

if [ -z "$BACKUP_FILE" ]; then
    echo "Usage: $0 <backup_file.sql.gz>"
    exit 1
fi

echo "Restoring database from $BACKUP_FILE"

# Stop application
systemctl stop healthsecure-backend

# Restore database
zcat $BACKUP_FILE | mysql --user=$DB_USER --password=$DB_PASSWORD $DB_NAME

# Start application
systemctl start healthsecure-backend

echo "Database restore completed"
```

## Troubleshooting

### Common Issues

1. **Database connection failed**:
```bash
# Check MySQL status
systemctl status mysql

# Test connection
mysql -u healthsecure_user -p -h localhost -e "SELECT 1"

# Check logs
journalctl -u mysql -f
```

2. **Redis connection failed**:
```bash
# Check Redis status
systemctl status redis

# Test connection
redis-cli ping

# Check configuration
grep -v '^#' /etc/redis/redis.conf
```

3. **SSL certificate issues**:
```bash
# Check certificate validity
openssl x509 -in /etc/ssl/certs/healthsecure.crt -text -noout

# Test SSL connection
openssl s_client -connect yourdomain.com:443 -servername yourdomain.com
```

4. **Application not starting**:
```bash
# Check application logs
journalctl -u healthsecure-backend -f

# Verify environment variables
sudo -u healthsecure printenv | grep -E "(DB|REDIS|JWT)"

# Test configuration
cd /opt/healthsecure/backend
sudo -u healthsecure ./bin/server --check-config
```

### Performance Tuning

1. **Database optimization**:
```sql
-- Check slow queries
SELECT * FROM mysql.slow_log ORDER BY start_time DESC LIMIT 10;

-- Optimize tables
OPTIMIZE TABLE patients, medical_records, audit_logs;

-- Update statistics
ANALYZE TABLE patients, medical_records, audit_logs;
```

2. **Application tuning**:
```bash
# Increase file descriptor limits
echo "healthsecure soft nofile 65535" >> /etc/security/limits.conf
echo "healthsecure hard nofile 65535" >> /etc/security/limits.conf

# Tune kernel parameters
echo "net.core.somaxconn = 65535" >> /etc/sysctl.conf
echo "net.ipv4.tcp_max_syn_backlog = 65535" >> /etc/sysctl.conf
sysctl -p
```

### Health Checks

```bash
#!/bin/bash
# /opt/healthsecure/scripts/health-check.sh

# Check application health
if ! curl -f http://localhost:8080/health > /dev/null 2>&1; then
    echo "ERROR: Application health check failed"
    exit 1
fi

# Check database connectivity
if ! mysql -u $DB_USER -p$DB_PASSWORD -e "SELECT 1" > /dev/null 2>&1; then
    echo "ERROR: Database connectivity check failed"
    exit 1
fi

# Check Redis connectivity
if ! redis-cli -a $REDIS_PASSWORD ping > /dev/null 2>&1; then
    echo "ERROR: Redis connectivity check failed"
    exit 1
fi

# Check disk space
DISK_USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
if [ $DISK_USAGE -gt 85 ]; then
    echo "WARNING: Disk usage is at ${DISK_USAGE}%"
fi

echo "All health checks passed"
```

---

**Note**: This deployment guide assumes a Linux-based production environment. Adjust paths, commands, and configurations based on your specific environment and requirements.
