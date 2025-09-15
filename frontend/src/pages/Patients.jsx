import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { patientsAPI, handleAPIError } from '../services/api';
import './Patients.css';

const Patients = () => {
  const { isDoctor, isNurse } = useAuth();
  const [patients, setPatients] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [pagination, setPagination] = useState({
    current_page: 1,
    total_pages: 1,
    total: 0,
    limit: 20
  });

  useEffect(() => {
    loadPatients();
  }, [pagination.current_page, searchTerm]);

  const loadPatients = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const params = {
        page: pagination.current_page,
        limit: pagination.limit
      };

      // Add search parameters
      if (searchTerm.trim()) {
        const response = await patientsAPI.searchPatients(searchTerm);
        setPatients(response.data.patients || []);
        setPagination(prev => ({ ...prev, total: response.data.patients?.length || 0 }));
      } else {
        const response = await patientsAPI.getPatients(params);
        setPatients(response.data.patients || []);
        setPagination(response.data.pagination || pagination);
      }
    } catch (err) {
      setError(handleAPIError(err));
      setPatients([]);
    } finally {
      setLoading(false);
    }
  };

  const handleSearch = (e) => {
    e.preventDefault();
    setPagination(prev => ({ ...prev, current_page: 1 }));
    loadPatients();
  };

  const handlePageChange = (newPage) => {
    setPagination(prev => ({ ...prev, current_page: newPage }));
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString();
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

  const canCreatePatient = isDoctor();

  return (
    <div className="patients-page">
      <div className="d-flex justify-content-between align-items-center mb-4">
        <div>
          <h1 className="mb-1">Patient Records</h1>
          <p className="text-muted mb-0">
            Manage and view patient information
          </p>
        </div>
      </div>

      {/* Search Bar */}
      <div className="card mb-4">
        <div className="card-body">
          <form onSubmit={handleSearch}>
            <div className="d-flex gap-2">
              <input
                type="text"
                className="form-input"
                placeholder="Search patients by name..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                style={{ flex: 1 }}
              />
              <button type="submit" className="btn btn-primary">
                🔍 Search
              </button>
              {searchTerm && (
                <button
                  type="button"
                  onClick={() => {
                    setSearchTerm('');
                    setPagination(prev => ({ ...prev, current_page: 1 }));
                  }}
                  className="btn btn-secondary"
                >
                  Clear
                </button>
              )}
            </div>
          </form>
        </div>
      </div>

      {/* Error Message */}
      {error && (
        <div className="alert alert-error mb-4">
          ⚠️ {error}
        </div>
      )}

      {/* Patients Table */}
      <div className="card">
        <div className="card-header">
          <h3 className="card-title">
            Patient List 
            {pagination.total > 0 && (
              <span style={{ fontWeight: 'normal', fontSize: '14px', color: '#666' }}>
                ({pagination.total} patients)
              </span>
            )}
          </h3>
        </div>
        
        <div className="card-body" style={{ padding: 0 }}>
          {loading ? (
            <div className="loading">
              <div className="spinner"></div>
            </div>
          ) : patients.length === 0 ? (
            <div style={{ padding: '40px', textAlign: 'center', color: '#666' }}>
              {searchTerm ? (
                <>
                  <div style={{ fontSize: '48px', marginBottom: '16px' }}>🔍</div>
                  <h3>No patients found</h3>
                  <p>No patients match your search criteria</p>
                </>
              ) : (
                <>
                  <div style={{ fontSize: '48px', marginBottom: '16px' }}>👥</div>
                  <h3>No patients yet</h3>
                  <p>Start by adding your first patient</p>
                  {canCreatePatient && (
                    <Link to="/patients/new" className="btn btn-primary mt-2">
                      Add First Patient
                    </Link>
                  )}
                </>
              )}
            </div>
          ) : (
            <div className="table-container">
              <table className="table">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Age</th>
                    <th>Phone</th>
                    {isDoctor() && <th>SSN</th>}
                    <th>Last Updated</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {patients.map((patient) => (
                    <tr key={patient.id}>
                      <td>
                        <div>
                          <div style={{ fontWeight: '600' }}>
                            {patient.first_name} {patient.last_name}
                          </div>
                          <div style={{ fontSize: '12px', color: '#666' }}>
                            ID: {patient.id}
                          </div>
                        </div>
                      </td>
                      <td>
                        {patient.date_of_birth ? getAge(patient.date_of_birth) : 'N/A'}
                      </td>
                      <td>{patient.phone || 'N/A'}</td>
                      {isDoctor() && (
                        <td style={{ fontFamily: 'monospace' }}>
                          {patient.ssn || 'N/A'}
                        </td>
                      )}
                      <td>{formatDate(patient.updated_at)}</td>
                      <td>
                        <div className="d-flex gap-1">
                          <Link
                            to={`/patients/${patient.id}`}
                            className="btn btn-secondary"
                            style={{ fontSize: '12px', padding: '6px 12px' }}
                          >
                            👁️ View
                          </Link>
                          {(isDoctor() || isNurse()) && (
                            <Link
                              to={`/patients/${patient.id}/edit`}
                              className="btn btn-primary"
                              style={{ fontSize: '12px', padding: '6px 12px' }}
                            >
                              ✏️ Edit
                            </Link>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Pagination */}
        {!loading && patients.length > 0 && pagination.total_pages > 1 && (
          <div className="card-footer">
            <div className="d-flex justify-content-between align-items-center">
              <div style={{ fontSize: '14px', color: '#666' }}>
                Page {pagination.current_page} of {pagination.total_pages}
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
                
                {/* Page numbers */}
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

      {/* Access Level Notice */}
      <div className="card mt-4" style={{ background: '#fff3cd', border: '1px solid #ffeaa7' }}>
        <div className="card-body">
          <div className="d-flex align-items-center gap-2">
            <span>ℹ️</span>
            <div style={{ fontSize: '14px', color: '#856404' }}>
              <strong>Data Access Level:</strong> {isDoctor() ? 'Full access including SSN and sensitive data' : 'Limited access - sensitive data restricted'}
            </div>
          </div>
        </div>
      </div>

      {canCreatePatient && (
        <Link to="/patients/new" className="fab">
          +
        </Link>
      )}
    </div>
  );
};

export default Patients;
