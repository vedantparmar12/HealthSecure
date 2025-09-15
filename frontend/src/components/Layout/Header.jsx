import React, { useState } from 'react';
import { useAuth } from '../../contexts/AuthContext';

const Header = () => {
  const { user, logout } = useAuth();
  const [showUserMenu, setShowUserMenu] = useState(false);

  const handleLogout = async () => {
    await logout();
    setShowUserMenu(false);
  };

  const getInitials = (name) => {
    return name
      .split(' ')
      .map(n => n[0])
      .join('')
      .toUpperCase();
  };

  const getRoleBadgeClass = (role) => {
    const baseClass = 'role-badge';
    return `${baseClass} ${role}`;
  };

  return (
    <header className="header">
      <div className="header-content">
        <h1 className="header-title">Medical Data Management</h1>
        
        <div className="user-info">
          <div className="user-details">
            <div className="user-name">{user?.name}</div>
            <div className={getRoleBadgeClass(user?.role)}>
              {user?.role}
            </div>
          </div>
          
          <div className="user-menu-container" style={{ position: 'relative' }}>
            <button
              className="user-avatar"
              onClick={() => setShowUserMenu(!showUserMenu)}
              style={{ border: 'none', cursor: 'pointer' }}
              title="User Menu"
            >
              {getInitials(user?.name || 'User')}
            </button>
            
            {showUserMenu && (
              <div 
                className="user-menu"
                style={{
                  position: 'absolute',
                  top: '100%',
                  right: 0,
                  background: 'white',
                  border: '1px solid #dee2e6',
                  borderRadius: '8px',
                  boxShadow: '0 4px 20px rgba(0,0,0,0.1)',
                  zIndex: 1000,
                  minWidth: '180px',
                  marginTop: '8px'
                }}
              >
                <div style={{ padding: '16px', borderBottom: '1px solid #dee2e6' }}>
                  <div style={{ fontWeight: '600', fontSize: '14px' }}>{user?.name}</div>
                  <div style={{ fontSize: '12px', color: '#666' }}>{user?.email}</div>
                </div>
                
                <div style={{ padding: '8px 0' }}>
                  <button
                    onClick={handleLogout}
                    style={{
                      width: '100%',
                      padding: '12px 16px',
                      border: 'none',
                      background: 'none',
                      textAlign: 'left',
                      cursor: 'pointer',
                      color: '#dc3545',
                      fontSize: '14px'
                    }}
                    onMouseOver={(e) => e.target.style.background = '#f8f9fa'}
                    onMouseOut={(e) => e.target.style.background = 'none'}
                  >
                    🔓 Logout
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
      
      {/* Click outside to close menu */}
      {showUserMenu && (
        <div
          style={{
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            zIndex: 999
          }}
          onClick={() => setShowUserMenu(false)}
        />
      )}
    </header>
  );
};

export default Header;