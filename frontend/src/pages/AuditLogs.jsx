import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { auditAPI, handleAPIError } from '../services/api';

const AuditLogs = () => {
  const { user, isAdmin } = useAuth();
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [filters, setFilters] = useState({
    page: 1,
    limit: 50,
    action: '',
    success: '',
    emergency: '',
    start_time: '',
    end_time: ''
  });
  const [pagination, setPagination] = useState({
    current_page: 1,
    total_pages: 1,
    total: 0,
    limit: 50
  });

  useEffect(() => {
    loadAuditLogs();
  }, [filters]);

  const loadAuditLogs = async () => {
    try {
      setLoading(true);
      setError(null);

      // Clean up filters - remove empty values
      const cleanFilters = Object.entries(filters).reduce((acc, [key, value]) => {
        if (value !== '' && value !== null && value !== undefined) {
          acc[key] = value;
        }
        return acc;
      }, {});

      const response = await auditAPI.getLogs(cleanFilters);
      setLogs(response.data.audit_logs || []);
      setPagination(response.data.pagination || pagination);

    } catch (err) {
      setError(handleAPIError(err));
      setLogs([]);
    } finally {
      setLoading(false);
    }
  };

  const handleFilterChange = (key, value) => {
    setFilters(prev => ({
      ...prev,
      [key]: value,
      page: 1 // Reset to first page when filters change
    }));
  };

  const handlePageChange = (newPage) => {
    setFilters(prev => ({ ...prev, page: newPage }));
  };

  const resetFilters = () => {
    setFilters({
      page: 1,
      limit: 50,
      action: '',
      success: '',
      emergency: '',
      start_time: '',
      end_time: ''
    });
  };

  const formatDateTime = (dateString) => {
    return new Date(dateString).toLocaleString();
  };

  const getActionIcon = (action) => {
    const icons = {
      LOGIN: '🔑',
      LOGOUT: '🔓',
      VIEW: '👁️',
      CREATE: '➕',
      UPDATE: '✏️',
      DELETE: '🗑️',
      EMERGENCY_REQUEST: '🚨',
      EMERGENCY_ACCESS: '⚡',
      UNAUTHORIZED_ACCESS: '⚠️'
    };
    return icons[action] || '📝';
  };

  const getActionColor = (action, success) => {
    if (!success) return '#dc3545';
    
    const colors = {
      LOGIN: '#28a745',
      LOGOUT: '#6c757d',
      VIEW: '#17a2b8',
      CREATE: '#28a745',
      UPDATE: '#ffc107',
      DELETE: '#dc3545',
      EMERGENCY_REQUEST: '#fd7e14',
      EMERGENCY_ACCESS: '#dc3545',
      UNAUTHORIZED_ACCESS: '#dc3545'
    };
    return colors[action] || '#6c757d';
  };

  const getSuccessIcon = (success) => {
    return success ? '✅' : '❌';
  };

  const getCurrentWeekRange = () => {
    const now = new Date();
    const weekStart = new Date(now);
    weekStart.setDate(now.getDate() - 7);
    
    return {
      start: weekStart.toISOString().split('T')[0],
      end: now.toISOString().split('T')[0]
    };
  };

  return (
    <div className="audit-logs-page">
      <div className="mb-4">
        <h1 className="mb-1">📋 Audit Logs</h1>
        <p className="text-muted mb-0">
          System activity and access audit trail for HIPAA compliance
        </p>
      </div>

      {/* Filters */}
      <div className="card mb-4">
        <div className="card-header">
          <h3 className="card-title">🔍 Filter Logs</h3>
        </div>
        <div className="card-body">
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: '16px' }}>
            <div className="form-group">
              <label className="form-label">Action</label>
              <select
                value={filters.action}
                onChange={(e) => handleFilterChange('action', e.target.value)}
                className="form-input"
              >
                <option value="">All Actions</option>
                <option value="LOGIN">Login</option>
                <option value="LOGOUT">Logout</option>
                <option value="VIEW">View</option>
                <option value="CREATE">Create</option>
                <option value="UPDATE">Update</option>
                <option value="DELETE">Delete</option>
                <option value="EMERGENCY_REQUEST">Emergency Request</option>
                <option value="EMERGENCY_ACCESS">Emergency Access</option>
                <option value="UNAUTHORIZED_ACCESS">Unauthorized</option>
              </select>
            </div>

            <div className="form-group">
              <label className="form-label">Status</label>
              <select
                value={filters.success}
                onChange={(e) => handleFilterChange('success', e.target.value)}
                className="form-input"
              >
                <option value="">All</option>
                <option value="true">Success</option>
                <option value="false">Failed</option>
              </select>
            </div>

            <div className="form-group">
              <label className="form-label">Emergency Access</label>
              <select
                value={filters.emergency}
                onChange={(e) => handleFilterChange('emergency', e.target.value)}
                className="form-input"
              >
                <option value="">All</option>
                <option value="true">Emergency Only</option>
                <option value="false">Regular Access</option>
              </select>
            </div>

            <div className="form-group">
              <label className="form-label">Start Date</label>
              <input
                type="date"
                value={filters.start_time}
                onChange={(e) => handleFilterChange('start_time', e.target.value)}
                className="form-input"
              />
            </div>

            <div className="form-group">
              <label className="form-label">End Date</label>
              <input
                type="date"
                value={filters.end_time}
                onChange={(e) => handleFilterChange('end_time', e.target.value)}
                className="form-input"
              />
            </div>
          </div>

          <div className="d-flex gap-2 mt-3">
            <button
              onClick={() => {
                const range = getCurrentWeekRange();
                setFilters(prev => ({
                  ...prev,
                  start_time: range.start,
                  end_time: range.end,
                  page: 1
                }));
              }}
              className="btn btn-secondary"
              style={{ fontSize: '12px', padding: '8px 12px' }}
            >
              📅 Last 7 Days
            </button>
            <button
              onClick={resetFilters}
              className="btn btn-secondary"
              style={{ fontSize: '12px', padding: '8px 12px' }}
            >
              🔄 Reset Filters
            </button>
          </div>
        </div>
      </div>

      {/* Error Message */}
      {error && (
        <div className="alert alert-error mb-4">
          ⚠️ {error}
        </div>
      )}

      {/* Audit Logs Table */}
      <div className="card">
        <div className="card-header">
          <h3 className="card-title">
            Audit Trail 
            {pagination.total > 0 && (
              <span style={{ fontWeight: 'normal', fontSize: '14px', color: '#666' }}>
                ({pagination.total} entries)
              </span>
            )}
          </h3>
        </div>
        
        <div className="card-body" style={{ padding: 0 }}>
          {loading ? (
            <div className="loading">
              <div className="spinner"></div>
            </div>
          ) : logs.length === 0 ? (
            <div style={{ padding: '40px', textAlign: 'center', color: '#666' }}>
              <div style={{ fontSize: '48px', marginBottom: '16px' }}>📋</div>
              <h3>No audit logs found</h3>
              <p>No logs match your current filter criteria</p>
            </div>
          ) : (
            <div className="table-container">
              <table className="table">
                <thead>
                  <tr>
                    <th>Timestamp</th>
                    <th>User</th>
                    <th>Action</th>
                    <th>Resource</th>
                    <th>Status</th>
                    <th>Emergency</th>
                    <th>IP Address</th>
                  </tr>
                </thead>
                <tbody>
                  {logs.map((log) => (
                    <tr key={log.id}>
                      <td style={{ fontSize: '12px', fontFamily: 'monospace' }}>
                        {formatDateTime(log.timestamp)}
                      </td>
                      <td>
                        <div>
                          <div style={{ fontWeight: '600' }}>
                            {log.user?.name || 'System'}
                          </div>
                          <div style={{ fontSize: '12px', color: '#666' }}>
                            {log.user?.role && (
                              <span className={`role-badge ${log.user.role}`}>
                                {log.user.role}
                              </span>
                            )}
                          </div>
                        </div>
                      </td>
                      <td>
                        <div className="d-flex align-items-center gap-1">
                          <span style={{ fontSize: '16px' }}>
                            {getActionIcon(log.action)}
                          </span>
                          <span 
                            style={{ 
                              color: getActionColor(log.action, log.success),
                              fontWeight: '600',
                              fontSize: '12px'
                            }}
                          >
                            {log.action}
                          </span>
                        </div>
                      </td>
                      <td style={{ fontSize: '12px', fontFamily: 'monospace', wordBreak: 'break-word' }}>
                        {log.resource}
                      </td>
                      <td>
                        <div className="d-flex align-items-center gap-1">
                          <span>{getSuccessIcon(log.success)}</span>
                          <span style={{ 
                            color: log.success ? '#28a745' : '#dc3545',
                            fontWeight: '600',
                            fontSize: '12px'
                          }}>
                            {log.success ? 'Success' : 'Failed'}
                          </span>
                        </div>
                        {!log.success && log.error_message && (
                          <div style={{ fontSize: '11px', color: '#dc3545', marginTop: '2px' }}>
                            {log.error_message}
                          </div>
                        )}
                      </td>
                      <td>
                        {log.emergency_use ? (
                          <span className="badge" style={{ 
                            background: '#dc3545', 
                            color: 'white',
                            fontSize: '10px',
                            padding: '3px 6px'
                          }}>
                            🚨 EMERGENCY
                          </span>
                        ) : (
                          <span style={{ fontSize: '12px', color: '#666' }}>Regular</span>
                        )}
                      </td>
                      <td style={{ fontSize: '12px', fontFamily: 'monospace' }}>
                        {log.ip_address}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Pagination */}
        {!loading && logs.length > 0 && pagination.total_pages > 1 && (
          <div className="card-footer">
            <div className="d-flex justify-content-between align-items-center">
              <div style={{ fontSize: '14px', color: '#666' }}>
                Page {pagination.current_page} of {pagination.total_pages} 
                ({pagination.total} total entries)
              </div>
              
              <div className="d-flex gap-1">
                <button
                  onClick={() => handlePageChange(pagination.current_page - 1)}
                  disabled={pagination.current_page === 1}
                  className="btn btn-secondary"
                  style={{ fontSize: '12px', padding: '6px 12px' }}
                >
                  ← Previous
                </button>
                
                {/* Show some page numbers */}
                {Array.from({ length: Math.min(5, pagination.total_pages) }, (_, i) => {
                  const pageNum = i + 1;
                  return (
                    <button
                      key={pageNum}
                      onClick={() => handlePageChange(pageNum)}
                      className={`btn ${pageNum === pagination.current_page ? 'btn-primary' : 'btn-secondary'}`}
                      style={{ fontSize: '12px', padding: '6px 12px', minWidth: '36px' }}
                    >
                      {pageNum}
                    </button>
                  );
                })}
                
                <button
                  onClick={() => handlePageChange(pagination.current_page + 1)}
                  disabled={pagination.current_page === pagination.total_pages}
                  className="btn btn-secondary"
                  style={{ fontSize: '12px', padding: '6px 12px' }}
                >
                  Next →
                </button>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* HIPAA Compliance Notice */}
      <div className="card mt-4" style={{ background: '#e3f2fd', border: '1px solid #bbdefb' }}>
        <div className="card-body">
          <div className="d-flex align-items-center gap-2">
            <span style={{ fontSize: '20px' }}>🔒</span>
            <div style={{ fontSize: '14px', color: '#1976d2' }}>
              <strong>HIPAA Audit Trail:</strong> This log provides a complete audit trail of all system access 
              and modifications as required by HIPAA regulations. All entries are tamper-proof and permanently stored.
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AuditLogs;