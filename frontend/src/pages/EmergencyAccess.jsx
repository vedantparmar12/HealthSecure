import React, { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { emergencyAPI, patientsAPI, handleAPIError } from '../services/api';

const EmergencyAccess = () => {
  const { user, isAdmin } = useAuth();
  const [activeRequests, setActiveRequests] = useState([]);
  const [userRequests, setUserRequests] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showRequestForm, setShowRequestForm] = useState(false);
  
  // Form state
  const [formData, setFormData] = useState({
    patient_id: '',
    reason: ''
  });
  const [patients, setPatients] = useState([]);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      setLoading(true);
      setError(null);

      // Load active emergency access (admin only)
      if (isAdmin()) {
        const activeResponse = await emergencyAPI.getActiveAccess();
        setActiveRequests(activeResponse.data.active_sessions || []);
      }

      // Load user's emergency access requests
      const userResponse = await emergencyAPI.getUserAccess(user.id);
      setUserRequests(userResponse.data.emergency_access || []);

      // Load patients for form
      const patientsResponse = await patientsAPI.getPatients({ limit: 100 });
      setPatients(patientsResponse.data.patients || []);

    } catch (err) {
      setError(handleAPIError(err));
    } finally {
      setLoading(false);
    }
  };

  const handleFormChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleSubmitRequest = async (e) => {
    e.preventDefault();
    
    if (!formData.patient_id || !formData.reason.trim()) {
      return;
    }

    if (formData.reason.length < 20) {
      setError('Reason must be at least 20 characters long');
      return;
    }

    try {
      setSubmitting(true);
      setError(null);

      await emergencyAPI.requestAccess({
        patient_id: parseInt(formData.patient_id),
        reason: formData.reason.trim()
      });

      // Reset form and reload data
      setFormData({ patient_id: '', reason: '' });
      setShowRequestForm(false);
      await loadData();

    } catch (err) {
      setError(handleAPIError(err));
    } finally {
      setSubmitting(false);
    }
  };

  const handleRevokeAccess = async (accessId) => {
    try {
      await emergencyAPI.revokeAccess(accessId);
      await loadData();
    } catch (err) {
      setError(handleAPIError(err));
    }
  };

  const formatDateTime = (dateString) => {
    return new Date(dateString).toLocaleString();
  };

  const getTimeRemaining = (expiresAt) => {
    const now = new Date();
    const expiry = new Date(expiresAt);
    const diffMs = expiry - now;
    
    if (diffMs <= 0) return 'Expired';
    
    const diffMins = Math.floor(diffMs / (1000 * 60));
    const diffHours = Math.floor(diffMins / 60);
    
    if (diffHours > 0) {
      return `${diffHours}h ${diffMins % 60}m remaining`;
    } else {
      return `${diffMins}m remaining`;
    }
  };

  const getStatusColor = (status) => {
    const colors = {
      pending: '#ffc107',
      active: '#28a745',
      used: '#17a2b8',
      expired: '#6c757d',
      revoked: '#dc3545'
    };
    return colors[status] || '#6c757d';
  };

  const getStatusIcon = (status) => {
    const icons = {
      pending: '⏳',
      active: '✅',
      used: '✔️',
      expired: '⏰',
      revoked: '❌'
    };
    return icons[status] || '⚪';
  };

  if (loading) {
    return (
      <div className="loading">
        <div className="spinner"></div>
      </div>
    );
  }

  return (
    <div className="emergency-access-page">
      <div className="d-flex justify-content-between align-items-center mb-4">
        <div>
          <h1 className="mb-1">🚨 Emergency Access</h1>
          <p className="text-muted mb-0">
            Request temporary access to patient data in emergency situations
          </p>
        </div>
        
        {!isAdmin() && (
          <button
            onClick={() => setShowRequestForm(true)}
            className="btn btn-danger"
          >
            🚨 Request Emergency Access
          </button>
        )}
      </div>

      {/* Error Message */}
      {error && (
        <div className="alert alert-error mb-4">
          ⚠️ {error}
        </div>
      )}

      {/* Emergency Access Request Form */}
      {showRequestForm && (
        <div className="card emergency-card mb-4">
          <div className="card-header">
            <h3 className="card-title">🚨 Request Emergency Access</h3>
          </div>
          <div className="card-body">
            <div className="alert alert-warning mb-3">
              ⚠️ <strong>Warning:</strong> Emergency access is for critical situations only. 
              All emergency access is logged, audited, and reviewed. Misuse may result in 
              disciplinary action.
            </div>

            <form onSubmit={handleSubmitRequest}>
              <div className="form-group mb-3">
                <label htmlFor="patient_id" className="form-label">
                  Patient *
                </label>
                <select
                  id="patient_id"
                  name="patient_id"
                  value={formData.patient_id}
                  onChange={handleFormChange}
                  className="form-input"
                  required
                >
                  <option value="">Select a patient...</option>
                  {patients.map(patient => (
                    <option key={patient.id} value={patient.id}>
                      {patient.first_name} {patient.last_name} (ID: {patient.id})
                    </option>
                  ))}
                </select>
              </div>

              <div className="form-group mb-4">
                <label htmlFor="reason" className="form-label">
                  Reason for Emergency Access * (minimum 20 characters)
                </label>
                <textarea
                  id="reason"
                  name="reason"
                  value={formData.reason}
                  onChange={handleFormChange}
                  className="form-input"
                  rows="4"
                  placeholder="Describe the emergency situation that requires immediate access to patient data..."
                  required
                  minLength="20"
                />
                <div style={{ fontSize: '12px', color: '#666', marginTop: '4px' }}>
                  {formData.reason.length}/20 characters minimum
                </div>
              </div>

              <div className="d-flex gap-2">
                <button
                  type="submit"
                  disabled={submitting || formData.reason.length < 20}
                  className={`btn btn-danger ${submitting ? 'btn-loading' : ''}`}
                >
                  {submitting ? (
                    <>
                      <div className="spinner" style={{ width: '16px', height: '16px', marginRight: '8px' }}></div>
                      Submitting Request...
                    </>
                  ) : (
                    '🚨 Submit Emergency Request'
                  )}
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setShowRequestForm(false);
                    setFormData({ patient_id: '', reason: '' });
                    setError(null);
                  }}
                  className="btn btn-secondary"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Active Emergency Sessions (Admin Only) */}
      {isAdmin() && (
        <div className="card mb-4">
          <div className="card-header">
            <h3 className="card-title">
              🔴 Active Emergency Sessions 
              {activeRequests.length > 0 && (
                <span className="badge" style={{ 
                  background: '#dc3545', 
                  color: 'white', 
                  marginLeft: '8px',
                  fontSize: '12px',
                  padding: '4px 8px'
                }}>
                  {activeRequests.length}
                </span>
              )}
            </h3>
          </div>
          <div className="card-body">
            {activeRequests.length === 0 ? (
              <div style={{ textAlign: 'center', padding: '20px', color: '#666' }}>
                <div style={{ fontSize: '32px', marginBottom: '12px' }}>✅</div>
                <p>No active emergency sessions</p>
              </div>
            ) : (
              <div className="table-container">
                <table className="table">
                  <thead>
                    <tr>
                      <th>User</th>
                      <th>Patient</th>
                      <th>Reason</th>
                      <th>Expires</th>
                      <th>Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {activeRequests.map(request => (
                      <tr key={request.id}>
                        <td>
                          <div>
                            <div style={{ fontWeight: '600' }}>
                              {request.user?.name}
                            </div>
                            <div style={{ fontSize: '12px', color: '#666' }}>
                              {request.user?.role}
                            </div>
                          </div>
                        </td>
                        <td>
                          <div>
                            <div style={{ fontWeight: '600' }}>
                              {request.patient?.first_name} {request.patient?.last_name}
                            </div>
                            <div style={{ fontSize: '12px', color: '#666' }}>
                              ID: {request.patient?.id}
                            </div>
                          </div>
                        </td>
                        <td>
                          <div style={{ maxWidth: '200px', wordBreak: 'break-word' }}>
                            {request.reason}
                          </div>
                        </td>
                        <td>
                          <div style={{ color: '#dc3545', fontWeight: '600' }}>
                            {getTimeRemaining(request.expires_at)}
                          </div>
                          <div style={{ fontSize: '12px', color: '#666' }}>
                            {formatDateTime(request.expires_at)}
                          </div>
                        </td>
                        <td>
                          <button
                            onClick={() => handleRevokeAccess(request.id)}
                            className="btn btn-danger"
                            style={{ fontSize: '12px', padding: '6px 12px' }}
                          >
                            🚫 Revoke
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Your Emergency Access Requests */}
      <div className="card">
        <div className="card-header">
          <h3 className="card-title">Your Emergency Access Requests</h3>
        </div>
        <div className="card-body">
          {userRequests.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '40px', color: '#666' }}>
              <div style={{ fontSize: '48px', marginBottom: '16px' }}>🚨</div>
              <h3>No Emergency Requests</h3>
              <p>You haven't made any emergency access requests yet</p>
            </div>
          ) : (
            <div className="requests-list">
              {userRequests.map(request => (
                <div key={request.id} className="card mb-3" style={{ border: '1px solid #dee2e6' }}>
                  <div className="card-body">
                    <div className="d-flex justify-content-between align-items-start mb-3">
                      <div>
                        <h4 className="mb-1">
                          {request.patient?.first_name} {request.patient?.last_name}
                        </h4>
                        <div style={{ fontSize: '14px', color: '#666' }}>
                          Request #{request.id} • {formatDateTime(request.created_at)}
                        </div>
                      </div>
                      <div className="d-flex align-items-center gap-2">
                        <span
                          className="badge"
                          style={{
                            background: getStatusColor(request.status),
                            color: 'white',
                            padding: '6px 12px',
                            borderRadius: '16px'
                          }}
                        >
                          {getStatusIcon(request.status)} {request.status}
                        </span>
                        {request.status === 'active' && (
                          <button
                            onClick={() => handleRevokeAccess(request.id)}
                            className="btn btn-danger"
                            style={{ fontSize: '12px', padding: '4px 8px' }}
                          >
                            🚫 Revoke
                          </button>
                        )}
                      </div>
                    </div>
                    
                    <div className="mb-3">
                      <label style={{ fontWeight: '600', display: 'block', marginBottom: '4px' }}>
                        Reason:
                      </label>
                      <p style={{ margin: '0', padding: '8px', background: '#f8f9fa', borderRadius: '4px' }}>
                        {request.reason}
                      </p>
                    </div>
                    
                    <div className="row">
                      <div className="col-md-6">
                        <div style={{ fontSize: '12px', color: '#666' }}>
                          <strong>Expires:</strong> {formatDateTime(request.expires_at)}
                        </div>
                      </div>
                      <div className="col-md-6">
                        {request.status === 'active' && (
                          <div style={{ fontSize: '12px', color: '#dc3545', fontWeight: '600' }}>
                            {getTimeRemaining(request.expires_at)}
                          </div>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Emergency Access Guidelines */}
      <div className="card mt-4" style={{ background: '#e3f2fd', border: '1px solid #bbdefb' }}>
        <div className="card-header" style={{ background: '#bbdefb', borderBottom: '1px solid #90caf9' }}>
          <h3 className="card-title" style={{ color: '#1976d2' }}>
            📋 Emergency Access Guidelines
          </h3>
        </div>
        <div className="card-body" style={{ color: '#1976d2' }}>
          <ul style={{ margin: '0', paddingLeft: '20px' }}>
            <li><strong>Use Only in True Emergencies:</strong> Emergency access should only be used when patient safety is at immediate risk.</li>
            <li><strong>Provide Detailed Justification:</strong> All requests must include a detailed reason explaining the emergency situation.</li>
            <li><strong>Time Limited:</strong> Emergency access automatically expires after 1 hour and cannot be extended.</li>
            <li><strong>Fully Audited:</strong> All emergency access is logged, monitored, and reviewed by administrators.</li>
            <li><strong>Accountability:</strong> Misuse of emergency access may result in disciplinary action and legal consequences.</li>
            <li><strong>Alternative Options:</strong> Consider contacting colleagues or supervisors before using emergency access.</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default EmergencyAccess;