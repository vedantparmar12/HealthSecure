# HealthSecure AI Service

A Python microservice that provides AI chat functionality using LangChain and LangSmith integration for the HealthSecure application.

## Features

- **LangChain Integration**: Full LangChain library support with conversation memory
- **LangSmith Tracing**: Complete observability and monitoring
- **Role-Based Responses**: Different system prompts for doctors, nurses, and admins
- **Conversation History**: Persistent chat history using SQLite
- **HIPAA Compliance**: Security-focused design for healthcare environments
- **OpenRouter Integration**: Uses OpenRouter for accessing various AI models

## API Endpoints

### Health Check
```
GET /health
```

### Chat
```
POST /chat
```
Request:
```json
{
  "message": "What are the symptoms of hypertension?",
  "thread_id": "thread_123",
  "user_id": "user_456", 
  "user_role": "doctor",
  "user_name": "Dr. Smith"
}
```

Response:
```json
{
  "response": "Hypertension symptoms include...",
  "thread_id": "thread_123",
  "run_id": "run_789",
  "model_used": "openai/gpt-3.5-turbo",
  "success": true
}
```

### Chat History
```
GET /history/{thread_id}
```

## Setup

### Local Development

1. **Install Dependencies**
```bash
cd ai-service
pip install -r requirements.txt
```

2. **Environment Configuration**
The service automatically loads environment variables from `../configs/.env`:
```bash
OPENROUTER_API_KEY=your_openrouter_key
LANGCHAIN_API_KEY=your_langsmith_key
LANGCHAIN_PROJECT=healthsecure-ai
LANGCHAIN_TRACING_V2=true
```

3. **Run the Service**
```bash
python app.py
```

### Docker Deployment

1. **Build Image**
```bash
docker build -t healthsecure-ai-service .
```

2. **Run Container**
```bash
docker run -p 5000:5000 \
  -v $(pwd)/../configs/.env:/app/../configs/.env:ro \
  healthsecure-ai-service
```

## Integration with Go Backend

Update your Go backend to call this Python service instead of handling AI directly:

```go
// In langgraph_chat.go
func (h *LangGraphChatHandler) processWithLangGraph(req ChatRequest, userID, threadID string) (*ChatResponse, error) {
    // Call Python AI service
    aiServiceURL := "http://localhost:5000/chat"
    
    payload := map[string]interface{}{
        "message":   req.Message,
        "thread_id": threadID,
        "user_id":   userID,
        "user_role": req.User.Role,
        "user_name": req.User.Name,
    }
    
    // Make HTTP request to Python service
    // Handle response and return ChatResponse
}
```

## Benefits

1. **True LangChain**: Full LangChain library with all features
2. **Better Memory**: Sophisticated conversation memory management  
3. **LangSmith Integration**: Complete tracing and monitoring
4. **Easier Maintenance**: Python ecosystem for AI development
5. **Scalability**: Microservice architecture allows independent scaling
6. **Language Specialization**: Go for backend logic, Python for AI

## Monitoring

- All conversations are traced in LangSmith
- Health check endpoint for monitoring
- Detailed logging for debugging
- Error tracking and reporting

## Security

- HIPAA-compliant design
- No patient data stored permanently
- Audit logging integration
- Role-based access control