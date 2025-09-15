import React, { useState } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { authAPI, handleAPIError } from '../services/api';

const Profile = () => {
  const { user, updateUser } = useAuth();
  const [activeTab, setActiveTab] = useState('info');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [success, setSuccess] = useState(null);

  // Profile update form
  const [profileForm, setProfileForm] = useState({
    name: user?.name || ''
  });

  // Password change form
  const [passwordForm, setPasswordForm] = useState({
    current_password: '',
    new_password: '',
    confirm_password: ''
  });

  const handleProfileChange = (e) => {
    const { name, value } = e.target;
    setProfileForm(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handlePasswordChange = (e) => {
    const { name, value } = e.target;
    setPasswordForm(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleProfileSubmit = async (e) => {
    e.preventDefault();
    
    if (!profileForm.name.trim()) {
      setError('Name is required');
      return;
    }

    try {
      setLoading(true);
      setError(null);

      // In a real implementation, this would call a profile update API
      // For now, we'll just update the context
      updateUser({
        ...user,
        name: profileForm.name
      });

      setSuccess('Profile updated successfully');
    } catch (err) {
      setError(handleAPIError(err));
    } finally {
      setLoading(false);
    }
  };

  const handlePasswordSubmit = async (e) => {
    e.preventDefault();

    if (!passwordForm.current_password || !passwordForm.new_password) {
      setError('All password fields are required');
      return;
    }

    if (passwordForm.new_password !== passwordForm.confirm_password) {
      setError('New passwords do not match');
      return;
    }

    if (passwordForm.new_password.length < 8) {
      setError('New password must be at least 8 characters long');
      return;
    }

    try {
      setLoading(true);
      setError(null);

      await authAPI.changePassword({
        current_password: passwordForm.current_password,
        new_password: passwordForm.new_password
      });

      setPasswordForm({
        current_password: '',
        new_password: '',
        confirm_password: ''
      });

      setSuccess('Password changed successfully');
    } catch (err) {
      setError(handleAPIError(err));
    } finally {
      setLoading(false);
    }
  };

  const clearMessages = () => {
    setError(null);
    setSuccess(null);
  };

  const getRoleBadgeStyle = (role) => {
    const styles = {
      admin: { background: '#e3f2fd', color: '#1976d2' },
      doctor: { background: '#e8f5e8', color: '#2e7d32' },
      nurse: { background: '#fff3e0', color: '#f57c00' }
    };
    return styles[role] || { background: '#f5f5f5', color: '#666' };
  };

  return (
    <div className="profile-page">
      <div className="mb-4">
        <h1 className="mb-1">👤 Profile Settings</h1>
        <p className="text-muted mb-0">
          Manage your account settings and security preferences
        </p>
      </div>

      {/* Messages */}
      {error && (
        <div className="alert alert-error mb-4">
          ⚠️ {error}
          <button 
            onClick={clearMessages}
            style={{ 
              float: 'right', 
              background: 'none', 
              border: 'none', 
              color: 'inherit',
              cursor: 'pointer'
            }}
          >
            ×
          </button>
        </div>
      )}
      
      {success && (
        <div className="alert alert-success mb-4">
          ✅ {success}
          <button 
            onClick={clearMessages}
            style={{ 
              float: 'right', 
              background: 'none', 
              border: 'none', 
              color: 'inherit',
              cursor: 'pointer'
            }}
          >
            ×
          </button>
        </div>
      )}

      <div className="row">
        {/* User Info Card */}
        <div className="col-md-4">
          <div className="card mb-4">
            <div className="card-body text-center">
              <div className="user-avatar" style={{ 
                width: '80px', 
                height: '80px', 
                fontSize: '32px',
                margin: '0 auto 20px'
              }}>
                {user?.name?.split(' ').map(n => n[0]).join('').toUpperCase() || 'U'}
              </div>
              
              <h3 className="mb-2">{user?.name}</h3>
              <div 
                className="badge mb-3"
                style={{
                  ...getRoleBadgeStyle(user?.role),
                  padding: '8px 16px',
                  borderRadius: '20px',
                  fontSize: '14px',
                  fontWeight: '600',
                  textTransform: 'capitalize'
                }}
              >
                {user?.role}
              </div>
              
              <div style={{ fontSize: '14px', color: '#666' }}>
                <div className="mb-2">
                  <strong>Email:</strong> {user?.email}
                </div>
                <div className="mb-2">
                  <strong>Status:</strong> 
                  <span className="status-indicator status-active ml-1">
                    <span className="status-dot"></span>
                    Active
                  </span>
                </div>
                <div>
                  <strong>Last Login:</strong><br />
                  {user?.last_login ? new Date(user.last_login).toLocaleString() : 'N/A'}
                </div>
              </div>
            </div>
          </div>

          {/* Role Permissions */}
          <div className="card">
            <div className="card-header">
              <h3 className="card-title">🔐 Your Permissions</h3>
            </div>
            <div className="card-body">
              <ul style={{ listStyle: 'none', padding: 0, margin: 0 }}>
                {user?.role === 'admin' && (
                  <>
                    <li className="mb-2">✅ User Management</li>
                    <li className="mb-2">✅ System Configuration</li>
                    <li className="mb-2">✅ Audit Log Access</li>
                    <li className="mb-2">✅ Security Monitoring</li>
                    <li className="mb-2">❌ Patient Data Access</li>
                  </>
                )}
                {user?.role === 'doctor' && (
                  <>
                    <li className="mb-2">✅ Full Patient Access</li>
                    <li className="mb-2">✅ Medical Records</li>
                    <li className="mb-2">✅ Sensitive Data (SSN)</li>
                    <li className="mb-2">✅ Create/Update Records</li>
                    <li className="mb-2">✅ Emergency Access</li>
                  </>
                )}
                {user?.role === 'nurse' && (
                  <>
                    <li className="mb-2">✅ Limited Patient Access</li>
                    <li className="mb-2">✅ Basic Medical Records</li>
                    <li className="mb-2">❌ Sensitive Data</li>
                    <li className="mb-2">✅ Update Care Notes</li>
                    <li className="mb-2">✅ Emergency Access</li>
                  </>
                )}
              </ul>
            </div>
          </div>
        </div>

        {/* Settings Panel */}
        <div className="col-md-8">
          {/* Tab Navigation */}
          <div className="card mb-4">
            <div className="card-header" style={{ padding: '0' }}>
              <div className="d-flex" style={{ borderBottom: '1px solid #dee2e6' }}>
                {[
                  { id: 'info', label: 'Profile Information', icon: '👤' },
                  { id: 'security', label: 'Security', icon: '🔒' },
                  { id: 'sessions', label: 'Sessions', icon: '📱' }
                ].map(tab => (
                  <button
                    key={tab.id}
                    onClick={() => {
                      setActiveTab(tab.id);
                      clearMessages();
                    }}
                    className={`btn ${activeTab === tab.id ? 'btn-primary' : 'btn-secondary'}`}
                    style={{
                      border: 'none',
                      borderRadius: '0',
                      borderBottom: activeTab === tab.id ? '2px solid #2c5aa0' : 'none',
                      background: activeTab === tab.id ? '#f8f9fa' : 'transparent',
                      color: activeTab === tab.id ? '#2c5aa0' : '#666'
                    }}
                  >
                    {tab.icon} {tab.label}
                  </button>
                ))}
              </div>
            </div>
          </div>

          {/* Profile Information Tab */}
          {activeTab === 'info' && (
            <div className="card">
              <div className="card-header">
                <h3 className="card-title">Profile Information</h3>
              </div>
              <div className="card-body">
                <form onSubmit={handleProfileSubmit}>
                  <div className="form-group mb-3">
                    <label htmlFor="name" className="form-label">
                      Full Name *
                    </label>
                    <input
                      type="text"
                      id="name"
                      name="name"
                      value={profileForm.name}
                      onChange={handleProfileChange}
                      className="form-input"
                      placeholder="Enter your full name"
                      required
                    />
                  </div>

                  <div className="form-group mb-3">
                    <label className="form-label">Email Address</label>
                    <input
                      type="email"
                      value={user?.email || ''}
                      className="form-input"
                      disabled
                      style={{ background: '#f8f9fa', cursor: 'not-allowed' }}
                    />
                    <div style={{ fontSize: '12px', color: '#666', marginTop: '4px' }}>
                      Email cannot be changed. Contact your administrator if needed.
                    </div>
                  </div>

                  <div className="form-group mb-4">
                    <label className="form-label">Role</label>
                    <input
                      type="text"
                      value={user?.role || ''}
                      className="form-input"
                      disabled
                      style={{ 
                        background: '#f8f9fa', 
                        cursor: 'not-allowed',
                        textTransform: 'capitalize'
                      }}
                    />
                    <div style={{ fontSize: '12px', color: '#666', marginTop: '4px' }}>
                      Role is assigned by administrators and cannot be self-modified.
                    </div>
                  </div>

                  <button
                    type="submit"
                    disabled={loading}
                    className={`btn btn-primary ${loading ? 'btn-loading' : ''}`}
                  >
                    {loading ? (
                      <>
                        <div className="spinner" style={{ width: '16px', height: '16px', marginRight: '8px' }}></div>
                        Updating...
                      </>
                    ) : (
                      '💾 Save Changes'
                    )}
                  </button>
                </form>
              </div>
            </div>
          )}

          {/* Security Tab */}
          {activeTab === 'security' && (
            <div className="card">
              <div className="card-header">
                <h3 className="card-title">Change Password</h3>
              </div>
              <div className="card-body">
                <form onSubmit={handlePasswordSubmit}>
                  <div className="form-group mb-3">
                    <label htmlFor="current_password" className="form-label">
                      Current Password *
                    </label>
                    <input
                      type="password"
                      id="current_password"
                      name="current_password"
                      value={passwordForm.current_password}
                      onChange={handlePasswordChange}
                      className="form-input"
                      placeholder="Enter your current password"
                      required
                      autoComplete="current-password"
                    />
                  </div>

                  <div className="form-group mb-3">
                    <label htmlFor="new_password" className="form-label">
                      New Password *
                    </label>
                    <input
                      type="password"
                      id="new_password"
                      name="new_password"
                      value={passwordForm.new_password}
                      onChange={handlePasswordChange}
                      className="form-input"
                      placeholder="Enter your new password"
                      required
                      minLength="8"
                      autoComplete="new-password"
                    />
                    <div style={{ fontSize: '12px', color: '#666', marginTop: '4px' }}>
                      Password must be at least 8 characters long
                    </div>
                  </div>

                  <div className="form-group mb-4">
                    <label htmlFor="confirm_password" className="form-label">
                      Confirm New Password *
                    </label>
                    <input
                      type="password"
                      id="confirm_password"
                      name="confirm_password"
                      value={passwordForm.confirm_password}
                      onChange={handlePasswordChange}
                      className="form-input"
                      placeholder="Confirm your new password"
                      required
                      autoComplete="new-password"
                    />
                  </div>

                  <button
                    type="submit"
                    disabled={loading}
                    className={`btn btn-primary ${loading ? 'btn-loading' : ''}`}
                  >
                    {loading ? (
                      <>
                        <div className="spinner" style={{ width: '16px', height: '16px', marginRight: '8px' }}></div>
                        Changing Password...
                      </>
                    ) : (
                      '🔐 Change Password'
                    )}
                  </button>
                </form>

                {/* Security Tips */}
                <div className="mt-4 p-3" style={{ background: '#e3f2fd', borderRadius: '8px', border: '1px solid #bbdefb' }}>
                  <h4 style={{ color: '#1976d2', fontSize: '16px', marginBottom: '12px' }}>
                    🔒 Password Security Tips
                  </h4>
                  <ul style={{ margin: '0', paddingLeft: '20px', color: '#1976d2', fontSize: '14px' }}>
                    <li>Use a strong, unique password for this system</li>
                    <li>Include uppercase, lowercase, numbers, and symbols</li>
                    <li>Avoid using personal information</li>
                    <li>Change your password regularly</li>
                    <li>Never share your credentials</li>
                  </ul>
                </div>
              </div>
            </div>
          )}

          {/* Sessions Tab */}
          {activeTab === 'sessions' && (
            <div className="card">
              <div className="card-header">
                <h3 className="card-title">Active Sessions</h3>
              </div>
              <div className="card-body">
                <div style={{ padding: '40px', textAlign: 'center', color: '#666' }}>
                  <div style={{ fontSize: '48px', marginBottom: '16px' }}>📱</div>
                  <h3>Session Management</h3>
                  <p>Active session information would be displayed here</p>
                  <div style={{ 
                    background: '#f8f9fa', 
                    padding: '20px', 
                    borderRadius: '8px',
                    marginTop: '20px'
                  }}>
                    <div style={{ fontWeight: '600', marginBottom: '8px' }}>Current Session</div>
                    <div style={{ fontSize: '14px' }}>
                      <div>Browser: {navigator.userAgent.split(' ')[0]}</div>
                      <div>Started: {new Date().toLocaleString()}</div>
                      <div>Status: Active</div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* CSS for responsive layout */}
      <style jsx>{`
        .col-md-4, .col-md-8 {
          flex: 1;
        }
        .col-md-4 {
          max-width: 300px;
          margin-right: 20px;
        }
        .col-md-8 {
          flex: 2;
        }
        .row {
          display: flex;
          gap: 0;
        }
        @media (max-width: 768px) {
          .row {
            flex-direction: column;
          }
          .col-md-4 {
            max-width: 100%;
            margin-right: 0;
            margin-bottom: 20px;
          }
        }
      `}</style>
    </div>
  );
};

export default Profile;