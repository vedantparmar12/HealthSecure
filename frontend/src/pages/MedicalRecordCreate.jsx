import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { medicalRecordsAPI, handleAPIError } from '../services/api';
import './MedicalRecordCreate.css';

const MedicalRecordCreate = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    diagnosis: '',
    treatment: '',
    notes: '',
    medications: '',
    severity: 'low',
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleChange = (e) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      await medicalRecordsAPI.createRecord(id, formData);
      navigate(`/patients/${id}`);
    } catch (err) {
      setError(handleAPIError(err));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="medical-record-create-page">
      <div className="card">
        <div className="card-header">
          <h1 className="card-title">Create New Medical Record</h1>
        </div>
        <div className="card-body">
          <form onSubmit={handleSubmit}>
            <div className="form-group">
              <label>Diagnosis</label>
              <textarea
                name="diagnosis"
                className="form-control"
                value={formData.diagnosis}
                onChange={handleChange}
                required
              />
            </div>
            <div className="form-group">
              <label>Treatment</label>
              <textarea
                name="treatment"
                className="form-control"
                value={formData.treatment}
                onChange={handleChange}
                required
              />
            </div>
            <div className="form-group">
              <label>Medications</label>
              <textarea
                name="medications"
                className="form-control"
                value={formData.medications}
                onChange={handleChange}
              />
            </div>
            <div className="form-group">
              <label>Notes</label>
              <textarea
                name="notes"
                className="form-control"
                value={formData.notes}
                onChange={handleChange}
              />
            </div>
            <div className="form-group">
              <label>Severity</label>
              <select
                name="severity"
                className="form-control"
                value={formData.severity}
                onChange={handleChange}
              >
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
                <option value="critical">Critical</option>
              </select>
            </div>
            {error && <div className="alert alert-danger">{error}</div>}
            <button type="submit" className="btn btn-primary btn-lg btn-block" disabled={loading}>
              {loading ? 'Creating...' : 'Create Record'}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
};

export default MedicalRecordCreate;
