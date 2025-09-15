import React, { useState, useEffect, useRef } from 'react';
import { useAuth } from '../contexts/AuthContext';
import api from '../services/api';
import './Chat.css';

const Chat = () => {
  const { user } = useAuth();
  const [messages, setMessages] = useState([]);
  const [inputMessage, setInputMessage] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [currentThread, setCurrentThread] = useState(null);
  const [threads, setThreads] = useState([]);
  const [showThreads, setShowThreads] = useState(false);
  const [inputEnabled, setInputEnabled] = useState(false);
  const messagesEndRef = useRef(null);

  useEffect(() => {
    // Enable input immediately for better UX
    setInputEnabled(true);
    loadThreads();
    createNewThread();
  }, []);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  const loadThreads = async () => {
    try {
      const response = await api.get('/chat/threads');
      setThreads(response.data.threads || []);
    } catch (error) {
      console.error('Failed to load threads:', error);
    }
  };

  const createNewThread = async () => {
    try {
      const response = await api.post('/chat/threads', {
        title: 'New Conversation'
      });
      
      const newThread = response.data.thread;
      setCurrentThread(newThread);
      setMessages([]);
      setInputEnabled(true);
      await loadThreads();
    } catch (error) {
      console.error('Failed to create new thread:', error);
      // Fallback: create a temporary thread for immediate use
      const fallbackThread = {
        thread_id: `temp_${Date.now()}`,
        title: 'New Conversation',
        created_at: new Date(),
        updated_at: new Date(),
        message_count: 0
      };
      setCurrentThread(fallbackThread);
      setMessages([]);
      setInputEnabled(true);
    }
  };

  const loadThread = async (threadId) => {
    try {
      const response = await api.get(`/chat/threads/${threadId}/messages`);
      const threadMessages = response.data.messages || [];
      
      // Convert backend messages to frontend format
      const formattedMessages = threadMessages.map(msg => ({
        id: msg.id,
        type: msg.role === 'user' ? 'user' : 'bot',
        content: msg.content,
        timestamp: new Date(msg.created_at)
      }));
      
      setMessages(formattedMessages);
      setCurrentThread(threads.find(t => t.thread_id === threadId));
      setInputEnabled(true);
      setShowThreads(false);
    } catch (error) {
      console.error('Failed to load thread messages:', error);
    }
  };

  const sendMessage = async () => {
    if (!inputMessage.trim() || isLoading || (!currentThread && !inputEnabled)) return;

    const userMessage = {
      id: Date.now(),
      type: 'user',
      content: inputMessage.trim(),
      timestamp: new Date()
    };

    setMessages(prev => [...prev, userMessage]);
    setInputMessage('');
    setIsLoading(true);

    try {
      // If no thread exists, create one first
      let threadToUse = currentThread;
      if (!threadToUse) {
        await createNewThread();
        threadToUse = currentThread;
      }

      const response = await api.post('/chat', {
        message: inputMessage.trim(),
        config: {
          configurable: {
            thread_id: threadToUse?.thread_id || `temp_${Date.now()}`
          },
          metadata: {
            thread_id: threadToUse?.thread_id || `temp_${Date.now()}`,
            user_id: user.id.toString(),
            user_role: user.role
          },
          run_name: `chat_${Date.now()}`
        },
        user: {
          id: user.id,
          role: user.role,
          name: user.name
        }
      });

      const botMessage = {
        id: Date.now() + 1,
        type: 'bot',
        content: response.data.response || 'Sorry, I encountered an error processing your request.',
        timestamp: new Date(),
        runId: response.data.run_id,
        threadId: response.data.thread_id
      };

      setMessages(prev => [...prev, botMessage]);
      
      // Reload threads to update the thread title if it changed
      await loadThreads();
    } catch (error) {
      console.error('Chat error:', error);
      const errorMessage = {
        id: Date.now() + 1,
        type: 'bot',
        content: 'Sorry, I encountered an error. Please try again.',
        timestamp: new Date(),
        isError: true
      };
      setMessages(prev => [...prev, errorMessage]);
    } finally {
      setIsLoading(false);
    }
  };

  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  const formatTimestamp = (timestamp) => {
    return timestamp.toLocaleTimeString('en-US', { 
      hour: '2-digit', 
      minute: '2-digit' 
    });
  };

  const getUserInitials = (name) => {
    return name?.split(' ').map(n => n[0]).join('').toUpperCase() || 'U';
  };

  return (
    <div className="chat-container">
      <div className="chat-header">
        <div className="chat-title">
          <h2>HealthSecure AI Assistant</h2>
          <p className="chat-subtitle">
            Powered by advanced AI • LangSmith observability enabled
          </p>
        </div>
        <div className="chat-controls">
          <button 
            className="btn-secondary"
            onClick={() => setShowThreads(!showThreads)}
          >
            <span className="icon">💬</span>
            History ({threads.length})
          </button>
          <button 
            className="btn-primary"
            onClick={createNewThread}
          >
            <span className="icon">✨</span>
            New Chat
          </button>
        </div>
      </div>

      {showThreads && (
        <div className="threads-panel">
          <div className="threads-header">
            <h3>Conversation History</h3>
            <button 
              className="close-btn"
              onClick={() => setShowThreads(false)}
            >
              ✕
            </button>
          </div>
          <div className="threads-list">
            {threads.map(thread => (
              <div
                key={thread.thread_id}
                className={`thread-item ${currentThread?.thread_id === thread.thread_id ? 'active' : ''}`}
                onClick={() => loadThread(thread.thread_id)}
              >
                <div className="thread-title">{thread.title}</div>
                <div className="thread-meta">
                  {new Date(thread.updated_at).toLocaleDateString()} • 
                  {thread.message_count || 0} messages
                </div>
              </div>
            ))}
            {threads.length === 0 && (
              <div className="no-threads">No conversations yet</div>
            )}
          </div>
        </div>
      )}

      <div className="chat-messages">
        {messages.length === 0 && (
          <div className="welcome-message">
            <div className="welcome-content">
              <h3>Welcome to HealthSecure AI Assistant</h3>
              <p>I can help you with:</p>
              <ul>
                <li>Patient information queries</li>
                <li>Medical record analysis</li>
                <li>Healthcare protocol guidance</li>
                <li>General medical questions</li>
              </ul>
              <p className="welcome-note">
                All conversations are monitored via LangSmith for quality and security.
              </p>
            </div>
          </div>
        )}

        {messages.map(message => (
          <div key={message.id} className={`message ${message.type}`}>
            <div className="message-avatar">
              {message.type === 'user' ? (
                <div className="user-avatar">
                  {getUserInitials(user.name)}
                </div>
              ) : (
                <div className="bot-avatar">
                  🤖
                </div>
              )}
            </div>
            <div className="message-content">
              <div className="message-header">
                <span className="message-author">
                  {message.type === 'user' ? user.name : 'HealthSecure AI'}
                </span>
                <span className="message-time">
                  {formatTimestamp(message.timestamp)}
                </span>
                {message.runId && (
                  <span className="message-meta">
                    Run ID: {message.runId.slice(-8)}
                  </span>
                )}
              </div>
              <div className={`message-text ${message.isError ? 'error' : ''}`}>
                {message.content}
              </div>
            </div>
          </div>
        ))}

        {isLoading && (
          <div className="message bot">
            <div className="message-avatar">
              <div className="bot-avatar">🤖</div>
            </div>
            <div className="message-content">
              <div className="message-header">
                <span className="message-author">HealthSecure AI</span>
              </div>
              <div className="message-text">
                <div className="typing-indicator">
                  <span></span>
                  <span></span>
                  <span></span>
                </div>
                <span className="typing-text">AI is thinking...</span>
              </div>
            </div>
          </div>
        )}
        
        <div ref={messagesEndRef} />
      </div>

      <div className="chat-input">
        <div className="input-container">
          <textarea
            value={inputMessage}
            onChange={(e) => setInputMessage(e.target.value)}
            onKeyPress={handleKeyPress}
            placeholder={`Ask HealthSecure AI anything... (Thread: ${currentThread?.thread_id?.slice(-8) || 'Ready'})`}
            disabled={isLoading}
            rows="1"
            className="message-input"
          />
          <button
            onClick={sendMessage}
            disabled={!inputMessage.trim() || isLoading}
            className="send-button"
          >
            {isLoading ? (
              <span className="loading-spinner">⏳</span>
            ) : (
              <span>Send</span>
            )}
          </button>
        </div>
        <div className="input-footer">
          <span className="footer-text">
            Powered by OpenRouter & LangSmith • Role: {user.role} • 
            Thread: {currentThread?.thread_id?.slice(-8) || 'Ready'}
          </span>
        </div>
      </div>
    </div>
  );
};

export default Chat;