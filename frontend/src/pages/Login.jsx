import React, { useState } from 'react';
import { useAuth } from '../contexts/AuthContext';

const Login = () => {
  const { login, loading, error, clearError } = useAuth();
  const [formData, setFormData] = useState({
    email: '',
    password: ''
  });

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
    
    // Clear error when user starts typing
    if (error) {
      clearError();
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!formData.email || !formData.password) {
      return;
    }

    await login(formData);
  };

  // Demo credentials helper
  const useDemoCredentials = (role) => {
    const demoCredentials = {
      admin: { email: 'admin@healthsecure.local', password: 'password123' },
      doctor: { email: 'dr.smith@hospital.local', password: 'password123' },
      nurse: { email: 'nurse.wilson@hospital.local', password: 'password123' }
    };

    setFormData(demoCredentials[role]);
  };

  return (
    <div className="login-page">
      <div className="login-container">
        <div className="login-header">
          <h1 className="login-title">HealthSecure</h1>
          <p className="login-subtitle">HIPAA-Compliant Medical Data System</p>
        </div>

        <form onSubmit={handleSubmit} className="login-form">
          {error && (
            <div className="alert alert-error">
              ⚠️ {error}
            </div>
          )}

          <div className="form-group">
            <label htmlFor="email" className="form-label">
              Email Address
            </label>
            <input
              type="email"
              id="email"
              name="email"
              value={formData.email}
              onChange={handleChange}
              className={`form-input ${error ? 'error' : ''}`}
              placeholder="Enter your email"
              required
              autoComplete="email"
            />
          </div>

          <div className="form-group">
            <label htmlFor="password" className="form-label">
              Password
            </label>
            <input
              type="password"
              id="password"
              name="password"
              value={formData.password}
              onChange={handleChange}
              className={`form-input ${error ? 'error' : ''}`}
              placeholder="Enter your password"
              required
              autoComplete="current-password"
            />
          </div>

          <button
            type="submit"
            className={`btn btn-primary btn-full-width ${loading ? 'btn-loading' : ''}`}
            disabled={loading}
          >
            {loading ? (
              <>
                <div className="spinner" style={{ width: '16px', height: '16px', marginRight: '8px' }}></div>
                Signing In...
              </>
            ) : (
              '🔐 Sign In'
            )}
          </button>
        </form>

        {/* Demo Credentials Helper */}
        <div style={{ 
          marginTop: '30px', 
          padding: '20px', 
          background: '#f8f9fa', 
          borderRadius: '8px',
          border: '1px solid #dee2e6'
        }}>
          <h3 style={{ 
            margin: '0 0 15px', 
            fontSize: '14px', 
            fontWeight: '600',
            color: '#333'
          }}>
            Demo Accounts
          </h3>
          <div style={{ display: 'flex', gap: '8px', flexWrap: 'wrap' }}>
            <button
              type="button"
              onClick={() => useDemoCredentials('admin')}
              className="btn btn-secondary"
              style={{ fontSize: '12px', padding: '8px 12px' }}
            >
              👨‍💼 Admin
            </button>
            <button
              type="button"
              onClick={() => useDemoCredentials('doctor')}
              className="btn btn-secondary"
              style={{ fontSize: '12px', padding: '8px 12px' }}
            >
              👨‍⚕️ Doctor
            </button>
            <button
              type="button"
              onClick={() => useDemoCredentials('nurse')}
              className="btn btn-secondary"
              style={{ fontSize: '12px', padding: '8px 12px' }}
            >
              👩‍⚕️ Nurse
            </button>
          </div>
          <p style={{ 
            margin: '12px 0 0', 
            fontSize: '11px', 
            color: '#666',
            fontStyle: 'italic'
          }}>
            Click any role above to auto-fill demo credentials
          </p>
        </div>

        {/* Security Notice */}
        <div style={{ 
          marginTop: '20px', 
          padding: '16px', 
          background: '#e3f2fd', 
          borderRadius: '8px',
          border: '1px solid #bbdefb'
        }}>
          <div style={{ fontSize: '12px', color: '#1976d2', textAlign: 'center' }}>
            🔒 This system is HIPAA compliant and all access is audited
          </div>
        </div>
      </div>
    </div>
  );
};

export default Login;