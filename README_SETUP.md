# HealthSecure Complete Setup Guide

## 🏥 Overview
HealthSecure is a complete HIPAA-compliant medical AI system with:
- **Go Backend**: Microservices architecture with JWT authentication
- **AI Service**: LangChain-powered AI with AG-UI RAG agent integration
- **Next.js Frontend**: Modern React UI with CopilotKit integration
- **Railway MySQL**: Cloud database for production reliability

## 🚀 Quick Start

### Prerequisites
- **Go 1.21+**: Download from https://golang.org/
- **Python 3.10+**: Download from https://python.org/
- **Node.js 18+**: Download from https://nodejs.org/
- **Git**: For version control

### 1. Clone and Setup
```bash
git clone <your-repo>
cd HealthSecure
```

### 2. Start All Services
```bash
# Option 1: Use the automated start script
start_local.bat

# Option 2: Manual start (recommended for development)
# Terminal 1 - Go Backend
cd backend
go mod tidy
go run cmd/server/main.go

# Terminal 2 - AI Service
cd ai-service
pip install -r requirements.txt
python app.py

# Terminal 3 - Frontend
cd frontend
npm install
npm run dev
```

### 3. Test the System
```bash
# Run integration tests
test_system.bat

# Or check services manually:
# - Backend: http://localhost:8080/health
# - AI Service: http://localhost:5000/health
# - Frontend: http://localhost:3000
```

## 🎯 System Architecture

### Backend Services
- **Port 8080**: Go backend with REST API
- **Port 5000**: Python AI service with AG-UI agent
- **Port 3000**: Next.js frontend
- **Railway MySQL**: Remote database connection

### Database Configuration
The system uses Railway MySQL with automatic fallback to SQLite for local development.

Database connection configured in: `configs/.env`
```env
DB_HOST=maglev.proxy.rlwy.net
DB_PORT=12371
DB_NAME=railway
DB_USER=root
DB_PASSWORD=seYqYtFMYUVhAexDvNyyQStGtxwKpkEf
```

## 🔐 Demo Credentials

| Role | Email | Password |
|------|-------|----------|
| Doctor | dr.smith@hospital.local | Doctor123 |
| Nurse | nurse.jane@hospital.local | Nurse123 |
| Admin | admin@hospital.local | admin123 |

## 🤖 AG-UI Features

The system includes AG-UI (Agent UI) capabilities that allow the AI to dynamically modify the frontend:

### Supported Commands
- **Theme Changes**: "change color to blue", "make it green"
- **Add Components**: "add a pie chart", "create a patient table"
- **UI Modifications**: "change the layout", "add a sidebar"

### AG-UI Endpoints
- `/agui/chat` - Enhanced chat with UI modification
- `/agui/state` - Current UI state
- CopilotKit integration for seamless AI interaction

## 📁 Project Structure

```
HealthSecure/
├── backend/                 # Go microservices
│   ├── cmd/server/         # Main server entry point
│   ├── internal/           # Internal packages
│   └── configs/           # Configuration
├── ai-service/             # Python AI service
│   ├── app.py             # Main Flask application
│   ├── true_agui_agent.py # AG-UI agent implementation
│   └── requirements.txt   # Python dependencies
├── frontend/               # Next.js frontend
│   ├── src/app/           # App router pages
│   ├── components/        # React components
│   └── package.json       # Node dependencies
├── configs/               # Shared configuration
└── *.bat                 # Windows run scripts
```

## 🔧 Development Workflow

### 1. Start Development Environment
```bash
start_local.bat
```

### 2. Access Services
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080/api
- **AI Service**: http://localhost:5000
- **AG-UI State**: http://localhost:5000/agui/state

### 3. Test AG-UI Features
1. Login with demo credentials
2. Go to "AI Assistant" tab
3. Try commands:
   - "Change the theme to blue"
   - "Add a pie chart showing patient age groups"
   - "Create a summary card"

### 4. Monitor Logs
Each service runs in its own terminal window with real-time logs.

## 🛠️ Troubleshooting

### Common Issues

**Backend won't start:**
- Check if Railway MySQL is accessible
- Verify `.env` file in `configs/` folder
- System will fallback to SQLite if MySQL fails

**AI Service errors:**
- Install Python dependencies: `pip install -r ai-service/requirements.txt`
- Check if all required Python packages are installed
- Verify AG-UI agent initialization in logs

**Frontend issues:**
- Install Node dependencies: `npm install` in frontend folder
- Check if ports 3000, 8080, 5000 are available
- Verify CopilotKit integration in browser console

**AG-UI not working:**
- Ensure AI service is running on port 5000
- Check `/agui/chat` endpoint availability
- Verify AG-UI agent is properly initialized

### Health Checks
```bash
# Backend
curl http://localhost:8080/health

# AI Service
curl http://localhost:5000/health

# AG-UI State
curl http://localhost:5000/agui/state
```

## 🔒 Security Features

- **JWT Authentication**: Secure token-based auth
- **HIPAA Compliance**: Full audit logging
- **Role-based Access**: Doctor, Nurse, Admin roles
- **Emergency Access**: Break-glass functionality
- **Secure Headers**: CORS, CSP, security headers

## 📊 Monitoring

- **Audit Logs**: Complete user activity tracking
- **Health Endpoints**: Service availability monitoring
- **Performance Metrics**: Request/response tracking
- **Security Events**: Threat detection and logging

## 🎨 AG-UI Capabilities

### Dynamic UI Generation
- Real-time component creation
- Theme customization
- Layout modifications
- Data visualization

### RAG Integration
- PDF document analysis
- Knowledge base search
- Context-aware responses
- Medical information retrieval

### Frontend State Management
- UI component state tracking
- Theme persistence
- User interaction logging
- Real-time updates

---

## 💡 Usage Examples

### Login and Authentication
1. Open http://localhost:3000
2. Use demo credentials to login
3. System authenticates via Go backend
4. JWT token stored for subsequent requests

### AI Assistant with AG-UI
1. Navigate to "AI Assistant" tab
2. Chat interface loads with CopilotKit
3. Try AG-UI commands:
   - "Show me patient statistics in a pie chart"
   - "Change the interface color to green"
   - "Add a dashboard widget for emergency alerts"

### Medical Data Access
1. Use Doctor/Nurse credentials
2. Access "Patients" tab
3. View HIPAA-compliant patient records
4. All access is logged for compliance

## 🆘 Support

For issues or questions:
1. Check logs in terminal windows
2. Run `test_system.bat` for diagnostics
3. Verify all services are running
4. Check Railway MySQL connection status

---

**🎉 You're ready to use HealthSecure with full AG-UI capabilities!**