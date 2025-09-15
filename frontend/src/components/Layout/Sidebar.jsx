import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';

const Sidebar = () => {
  const { user, isAdmin, isMedicalStaff } = useAuth();
  const location = useLocation();

  const navItems = [
    {
      path: '/dashboard',
      label: 'Dashboard',
      icon: '📊',
      roles: ['admin', 'doctor', 'nurse']
    },
    {
      path: '/patients',
      label: 'Patients',
      icon: '👥',
      roles: ['doctor', 'nurse']
    },
    {
      path: '/chat',
      label: 'AI Assistant',
      icon: '🤖',
      roles: ['admin', 'doctor', 'nurse']
    },
    {
      path: '/emergency',
      label: 'Emergency Access',
      icon: '🚨',
      roles: ['doctor', 'nurse']
    },
    {
      path: '/audit',
      label: 'Audit Logs',
      icon: '📋',
      roles: ['admin', 'doctor', 'nurse']
    },
    {
      path: '/profile',
      label: 'Profile',
      icon: '👤',
      roles: ['admin', 'doctor', 'nurse']
    }
  ];

  const filteredNavItems = navItems.filter(item => 
    item.roles.includes(user?.role)
  );

  return (
    <aside className="sidebar">
      <div className="sidebar-header">
        <h1>HealthSecure</h1>
        <div className="subtitle">HIPAA-Compliant System</div>
      </div>
      
      <nav>
        <ul className="nav-menu">
          {filteredNavItems.map((item) => (
            <li key={item.path} className="nav-item">
              <Link 
                to={item.path} 
                className={`nav-link ${location.pathname === item.path ? 'active' : ''}`}
              >
                <span className="nav-icon">{item.icon}</span>
                {item.label}
              </Link>
            </li>
          ))}
        </ul>
      </nav>

      <div className="sidebar-footer" style={{ 
        padding: '20px', 
        borderTop: '1px solid rgba(255,255,255,0.2)', 
        marginTop: 'auto' 
      }}>
        <div style={{ fontSize: '12px', color: 'rgba(255,255,255,0.7)' }}>
          <div>Version 1.0.0</div>
          <div>HIPAA Compliant</div>
        </div>
      </div>
    </aside>
  );
};

export default Sidebar;