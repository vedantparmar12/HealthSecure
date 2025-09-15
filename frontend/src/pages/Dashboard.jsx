import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { adminAPI, handleAPIError } from '../services/api';

const Dashboard = () => {
  const { user, isAdmin, isMedicalStaff } = useAuth();
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    loadDashboardData();
  }, []);

  const loadDashboardData = async () => {
    try {
      setLoading(true);
      if (isAdmin()) {
        const response = await adminAPI.getDashboardStats();
        setStats(response.data.dashboard_stats);
      } else {
        // For medical staff, show basic stats
        setStats({
          user_role: user.role,
          access_level: user.role === 'doctor' ? 'Full Access' : 'Limited Access',
          system_status: 'healthy'
        });
      }
    } catch (err) {
      setError(handleAPIError(err));
    } finally {
      setLoading(false);
    }
  };

  const getGreeting = () => {
    const hour = new Date().getHours();
    if (hour < 12) return 'Good morning';
    if (hour < 18) return 'Good afternoon';
    return 'Good evening';
  };

  const quickActions = [
    ...(isMedicalStaff() ? [
      { title: 'View Patients', description: 'Access patient records', link: '/patients', icon: '👥' },
      { title: 'Emergency Access', description: 'Request emergency access', link: '/emergency', icon: '🚨' },
    ] : []),
    { title: 'Audit Logs', description: 'View system audit trail', link: '/audit', icon: '📋' },
    { title: 'Profile Settings', description: 'Update your profile', link: '/profile', icon: '👤' },
  ];

  if (loading) {
    return (
      <div className="loading">
        <div className="spinner"></div>
      </div>
    );
  }

  return (
    <div className="dashboard">
      <div className="dashboard-header mb-4">
        <h1 className="mb-1">
          {getGreeting()}, {user?.name}! 👋
        </h1>
        <p className="text-muted mb-0">
          Welcome to the HealthSecure medical data management system
        </p>
      </div>

      {error && (
        <div className="alert alert-error mb-3">
          ⚠️ {error}
        </div>
      )}

      {/* Stats Cards */}
      {stats && (
        <div className="stats-grid mb-4" style={{ 
          display: 'grid', 
          gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', 
          gap: '20px',
          marginBottom: '30px'
        }}>
          {isAdmin() ? (
            <>
              <div className="card">
                <div className="card-body text-center">
                  <div style={{ fontSize: '32px', color: '#2c5aa0', marginBottom: '8px' }}>
                    {stats.total_users}
                  </div>
                  <div style={{ fontWeight: '600', marginBottom: '4px' }}>Total Users</div>
                  <div style={{ fontSize: '12px', color: '#666' }}>
                    {stats.active_users} active
                  </div>
                </div>
              </div>
              <div className="card">
                <div className="card-body text-center">
                  <div style={{ fontSize: '32px', color: '#28a745', marginBottom: '8px' }}>
                    {stats.total_patients}
                  </div>
                  <div style={{ fontWeight: '600', marginBottom: '4px' }}>Total Patients</div>
                  <div style={{ fontSize: '12px', color: '#666' }}>
                    In system
                  </div>
                </div>
              </div>
              <div className="card">
                <div className="card-body text-center">
                  <div style={{ fontSize: '32px', color: '#17a2b8', marginBottom: '8px' }}>
                    {stats.recent_logins}
                  </div>
                  <div style={{ fontWeight: '600', marginBottom: '4px' }}>Recent Logins</div>
                  <div style={{ fontSize: '12px', color: '#666' }}>
                    Last 24 hours
                  </div>
                </div>
              </div>
              <div className="card">
                <div className="card-body text-center">
                  <div style={{ fontSize: '32px', color: stats.security_alerts > 0 ? '#dc3545' : '#28a745', marginBottom: '8px' }}>
                    {stats.security_alerts}
                  </div>
                  <div style={{ fontWeight: '600', marginBottom: '4px' }}>Security Alerts</div>
                  <div style={{ fontSize: '12px', color: '#666' }}>
                    Requires attention
                  </div>
                </div>
              </div>
            </>
          ) : (
            <>
              <div className="card">
                <div className="card-body text-center">
                  <div style={{ fontSize: '24px', marginBottom: '8px' }}>👤</div>
                  <div style={{ fontWeight: '600', marginBottom: '4px' }}>Your Role</div>
                  <div className={`role-badge ${user.role}`}>
                    {user.role}
                  </div>
                </div>
              </div>
              <div className="card">
                <div className="card-body text-center">
                  <div style={{ fontSize: '24px', marginBottom: '8px' }}>🔐</div>
                  <div style={{ fontWeight: '600', marginBottom: '4px' }}>Access Level</div>
                  <div style={{ fontSize: '14px', color: '#666' }}>
                    {stats.access_level}
                  </div>
                </div>
              </div>
              <div className="card">
                <div className="card-body text-center">
                  <div style={{ fontSize: '24px', marginBottom: '8px' }}>
                    {stats.system_status === 'healthy' ? '✅' : '⚠️'}
                  </div>
                  <div style={{ fontWeight: '600', marginBottom: '4px' }}>System Status</div>
                  <div style={{ fontSize: '14px', color: stats.system_status === 'healthy' ? '#28a745' : '#dc3545' }}>
                    {stats.system_status}
                  </div>
                </div>
              </div>
            </>
          )}
        </div>
      )}

      {/* Quick Actions */}
      <div className="card">
        <div className="card-header">
          <h3 className="card-title">Quick Actions</h3>
        </div>
        <div className="card-body">
          <div style={{ 
            display: 'grid', 
            gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))', 
            gap: '20px' 
          }}>
            {quickActions.map((action, index) => (
              <Link
                key={index}
                to={action.link}
                className="card"
                style={{ 
                  textDecoration: 'none',
                  transition: 'transform 0.2s ease',
                  cursor: 'pointer'
                }}
                onMouseOver={(e) => e.currentTarget.style.transform = 'translateY(-2px)'}
                onMouseOut={(e) => e.currentTarget.style.transform = 'translateY(0)'}
              >
                <div className="card-body">
                  <div style={{ fontSize: '32px', marginBottom: '12px' }}>
                    {action.icon}
                  </div>
                  <h4 style={{ margin: '0 0 8px', color: '#333' }}>
                    {action.title}
                  </h4>
                  <p style={{ margin: '0', color: '#666', fontSize: '14px' }}>
                    {action.description}
                  </p>
                </div>
              </Link>
            ))}
          </div>
        </div>
      </div>

      {/* HIPAA Compliance Notice */}
      <div className="card mt-4" style={{ background: '#e3f2fd', border: '1px solid #bbdefb' }}>
        <div className="card-body">
          <div className="d-flex align-items-center gap-2">
            <span style={{ fontSize: '20px' }}>🔒</span>
            <div>
              <h4 style={{ margin: '0 0 4px', color: '#1976d2' }}>HIPAA Compliance</h4>
              <p style={{ margin: '0', fontSize: '14px', color: '#1976d2' }}>
                This system maintains full HIPAA compliance. All access is logged and audited. 
                Unauthorized access or misuse of patient data may result in legal action.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;