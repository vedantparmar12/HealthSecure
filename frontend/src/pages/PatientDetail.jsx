import React, { useState, useEffect } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { patientsAPI, medicalRecordsAPI, handleAPIError } from '../services/api';

const PatientDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const { isDoctor, isNurse } = useAuth();
  
  const [patient, setPatient] = useState(null);
  const [medicalRecords, setMedicalRecords] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [activeTab, setActiveTab] = useState('overview');
  const [emergencyToken, setEmergencyToken] = useState(null);

  useEffect(() => {
    loadPatientData();
  }, [id, emergencyToken]);

  const loadPatientData = async () => {
    try {
      setLoading(true);
      setError(null);

      // Load patient data
      const patientResponse = await patientsAPI.getPatient(id, emergencyToken);
      setPatient(patientResponse.data.patient);

      // Load medical records
      const recordsResponse = await medicalRecordsAPI.getPatientRecords(id, {}, emergencyToken);
      setMedicalRecords(recordsResponse.data.records || []);

    } catch (err) {
      setError(handleAPIError(err));
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString();
  };

  const formatDateTime = (dateString) => {
    return new Date(dateString).toLocaleString();
  };

  const getAge = (birthDate) => {
    const today = new Date();
    const birth = new Date(birthDate);
    let age = today.getFullYear() - birth.getFullYear();
    const monthDiff = today.getMonth() - birth.getMonth();
    
    if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birth.getDate())) {
      age--;
    }
    
    return age;
  };

  const getSeverityColor = (severity) => {
    const colors = {
      low: '#28a745',
      medium: '#ffc107',
      high: '#fd7e14',
      critical: '#dc3545'
    };
    return colors[severity] || '#6c757d';
  };

  const getSeverityIcon = (severity) => {
    const icons = {
      low: '🟢',
      medium: '🟡',
      high: '🟠',
      critical: '🔴'
    };
    return icons[severity] || '⚪';
  };

  if (loading) {
    return (
      <div className="loading">
        <div className="spinner"></div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="error-page">
        <div className="alert alert-error">
          ⚠️ {error}
        </div>
        <button onClick={() => navigate('/patients')} className="btn btn-primary">
          ← Back to Patients
        </button>
      </div>
    );
  }

  if (!patient) {
    return (
      <div className="error-page">
        <div className="alert alert-error">
          Patient not found
        </div>
        <button onClick={() => navigate('/patients')} className="btn btn-primary">
          ← Back to Patients
        </button>
      </div>
    );
  }

  return (
    <div className="patient-detail">
      {/* Emergency Access Banner */}
      {emergencyToken && (
        <div className="emergency-banner mb-3">
          ⚠️ Emergency Access Mode Active - All access is being logged and audited
        </div>
      )}

      {/* Header */}
      <div className="d-flex justify-content-between align-items-start mb-4">
        <div>
          <div className="d-flex align-items-center gap-2 mb-2">
            <button
              onClick={() => navigate('/patients')}
              className="btn btn-secondary"
              style={{ fontSize: '12px', padding: '8px 12px' }}
            >
              ← Back
            </button>
            <h1 className="mb-0">
              {patient.first_name} {patient.last_name}
            </h1>
            <span className="role-badge" style={{ background: '#e3f2fd', color: '#1976d2' }}>
              ID: {patient.id}
            </span>
          </div>
          <p className="text-muted mb-0">
            Patient Details and Medical History
          </p>
        </div>
        
        <div className="d-flex gap-2">
          {isDoctor() && (
            <Link
              to={`/patients/${id}/records/new`}
              className="btn btn-success"
            >
              📝 Add Medical Record
            </Link>
          )}
          {(isDoctor() || isNurse()) && (
            <Link
              to={`/patients/${id}/edit`}
              className="btn btn-primary"
            >
              ✏️ Edit Patient
            </Link>
          )}
        </div>
      </div>

      {/* Tab Navigation */}
      <div className="card mb-4">
        <div className="card-header" style={{ padding: '0' }}>
          <div className="d-flex" style={{ borderBottom: '1px solid #dee2e6' }}>
            {[
              { id: 'overview', label: 'Overview', icon: '👤' },
              { id: 'medical', label: 'Medical Records', icon: '📋', count: medicalRecords.length },
              { id: 'audit', label: 'Access Log', icon: '📊' }
            ].map(tab => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
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
                {tab.count !== undefined && (
                  <span className="badge" style={{ 
                    background: '#2c5aa0', 
                    color: 'white', 
                    marginLeft: '8px',
                    fontSize: '11px',
                    padding: '2px 6px',
                    borderRadius: '10px'
                  }}>
                    {tab.count}
                  </span>
                )}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Tab Content */}
      {activeTab === 'overview' && (
        <div className="row">
          <div className="col-md-6">
            <div className="card">
              <div className="card-header">
                <h3 className="card-title">Personal Information</h3>
              </div>
              <div className="card-body">
                <div className="patient-info">
                  <div className="info-row mb-3">
                    <label>Full Name:</label>
                    <span>{patient.first_name} {patient.last_name}</span>
                  </div>
                  
                  <div className="info-row mb-3">
                    <label>Age:</label>
                    <span>{getAge(patient.date_of_birth)} years old</span>
                  </div>
                  
                  <div className="info-row mb-3">
                    <label>Date of Birth:</label>
                    <span>{formatDate(patient.date_of_birth)}</span>
                  </div>
                  
                  {isDoctor() && patient.ssn && (
                    <div className="info-row mb-3">
                      <label>SSN:</label>
                      <span style={{ fontFamily: 'monospace', color: '#dc3545' }}>
                        {patient.ssn}
                      </span>
                    </div>
                  )}
                  
                  <div className="info-row mb-3">
                    <label>Phone:</label>
                    <span>{patient.phone || 'Not provided'}</span>
                  </div>
                  
                  <div className="info-row mb-3">
                    <label>Address:</label>
                    <span>{patient.address || 'Not provided'}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
          
          <div className="col-md-6">
            <div className="card">
              <div className="card-header">
                <h3 className="card-title">Emergency Contact</h3>
              </div>
              <div className="card-body">
                <div className="info-row">
                  <span>{patient.emergency_contact || 'No emergency contact on file'}</span>
                </div>
              </div>
            </div>
            
            <div className="card mt-4">
              <div className="card-header">
                <h3 className="card-title">Record Information</h3>
              </div>
              <div className="card-body">
                <div className="info-row mb-3">
                  <label>Created:</label>
                  <span>{formatDateTime(patient.created_at)}</span>
                </div>
                <div className="info-row">
                  <label>Last Updated:</label>
                  <span>{formatDateTime(patient.updated_at)}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {activeTab === 'medical' && (
        <div>
          {medicalRecords.length === 0 ? (
            <div className="card">
              <div className="card-body text-center" style={{ padding: '40px' }}>
                <div style={{ fontSize: '48px', marginBottom: '16px' }}>📋</div>
                <h3>No Medical Records</h3>
                <p className="text-muted">No medical records have been added for this patient yet.</p>
                {isDoctor() && (
                  <Link
                    to={`/patients/${id}/records/new`}
                    className="btn btn-primary"
                  >
                    📝 Add First Record
                  </Link>
                )}
              </div>
            </div>
          ) : (
            <div className="medical-records">
              {medicalRecords.map((record, index) => (
                <div key={record.id} className="card mb-3">
                  <div className="card-header">
                    <div className="d-flex justify-content-between align-items-center">
                      <div className="d-flex align-items-center gap-2">
                        <h4 className="mb-0">Record #{medicalRecords.length - index}</h4>
                        <span 
                          className="badge"
                          style={{
                            background: getSeverityColor(record.severity),
                            color: 'white',
                            fontSize: '12px'
                          }}
                        >
                          {getSeverityIcon(record.severity)} {record.severity}
                        </span>
                      </div>
                      <div style={{ fontSize: '14px', color: '#666' }}>
                        {formatDateTime(record.created_at)}
                      </div>
                    </div>
                  </div>
                  <div className="card-body">
                    <div className="row">
                      <div className="col-md-6">
                        <div className="mb-3">
                          <label style={{ fontWeight: '600', display: 'block', marginBottom: '4px' }}>
                            Diagnosis:
                          </label>
                          <p>{record.diagnosis || 'Not specified'}</p>
                        </div>
                        <div className="mb-3">
                          <label style={{ fontWeight: '600', display: 'block', marginBottom: '4px' }}>
                            Treatment:
                          </label>
                          <p>{record.treatment || 'Not specified'}</p>
                        </div>
                      </div>
                      <div className="col-md-6">
                        <div className="mb-3">
                          <label style={{ fontWeight: '600', display: 'block', marginBottom: '4px' }}>
                            Medications:
                          </label>
                          <p>{record.medications || 'None prescribed'}</p>
                        </div>
                        <div className="mb-3">
                          <label style={{ fontWeight: '600', display: 'block', marginBottom: '4px' }}>
                            Notes:
                          </label>
                          <p>{record.notes || 'No additional notes'}</p>
                        </div>
                      </div>
                    </div>
                    {record.doctor && (
                      <div style={{ 
                        marginTop: '12px', 
                        paddingTop: '12px', 
                        borderTop: '1px solid #dee2e6',
                        fontSize: '12px',
                        color: '#666'
                      }}>
                        Created by: Dr. {record.doctor.name}
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {activeTab === 'audit' && (
        <div className="card">
          <div className="card-header">
            <h3 className="card-title">Access Audit Log</h3>
          </div>
          <div className="card-body">
            <div style={{ padding: '40px', textAlign: 'center', color: '#666' }}>
              <div style={{ fontSize: '48px', marginBottom: '16px' }}>📊</div>
              <h3>Audit logs would appear here</h3>
              <p>This would show all access to this patient's records</p>
            </div>
          </div>
        </div>
      )}

      {/* CSS for info rows */}
      <style jsx>{`
        .info-row {
          display: flex;
          justify-content: space-between;
          align-items: flex-start;
        }
        .info-row label {
          font-weight: 600;
          color: #333;
          min-width: 120px;
        }
        .info-row span {
          flex: 1;
          text-align: right;
        }
        .col-md-6 {
          flex: 1;
          margin-right: 20px;
        }
        .col-md-6:last-child {
          margin-right: 0;
        }
        .row {
          display: flex;
          gap: 0;
        }
        @media (max-width: 768px) {
          .row {
            flex-direction: column;
          }
          .col-md-6 {
            margin-right: 0;
            margin-bottom: 20px;
          }
        }
      `}</style>
    </div>
  );
};

export default PatientDetail;