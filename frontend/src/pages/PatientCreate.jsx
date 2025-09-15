import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { patientsAPI, handleAPIError } from '../services/api';
import './PatientCreate.css';

const PatientCreate = () => {
  const navigate = useNavigate();
  const [formData, setFormData] = useState({
    first_name: '',
    last_name: '',
    date_of_birth: '',
    ssn: '',
    phone: '',
    address: '',
    emergency_contact: '',
    description: '',
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
      await patientsAPI.createPatient(formData);
      navigate('/patients');
    } catch (err) {
      setError(handleAPIError(err));
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="patient-create-page">
      <div className="card">
        <div className="card-header">
          <h1 className="card-title">Create New Patient</h1>
        </div>
        <div className="card-body">
          <form onSubmit={handleSubmit}>
            <div className="form-row">
              <div className="form-group col-md-6">
                <label><i className="fas fa-user"></i> First Name</label>
                <input
                  type="text"
                  name="first_name"
                  className="form-control"
                  placeholder="Enter first name"
                  value={formData.first_name}
                  onChange={handleChange}
                  required
                />
              </div>
              <div className="form-group col-md-6">
                <label><i className="fas fa-user"></i> Last Name</label>
                <input
                  type="text"
                  name="last_name"
                  className="form-control"
                  placeholder="Enter last name"
                  value={formData.last_name}
                  onChange={handleChange}
                  required
                />
              </div>
            </div>
            <div className="form-row">
              <div className="form-group col-md-6">
                <label><i className="fas fa-calendar-alt"></i> Date of Birth</label>
                <input
                  type="date"
                  name="date_of_birth"
                  className="form-control"
                  value={formData.date_of_birth}
                  onChange={handleChange}
                  required
                />
              </div>
              <div className="form-group col-md-6">
                <label><i className="fas fa-id-card"></i> SSN</label>
                <input
                  type="text"
                  name="ssn"
                  className="form-control"
                  placeholder="Enter SSN"
                  value={formData.ssn}
                  onChange={handleChange}
                />
              </div>
            </div>
            <div className="form-row">
              <div className="form-group col-md-6">
                <label><i className="fas fa-phone"></i> Phone</label>
                <input
                  type="text"
                  name="phone"
                  className="form-control"
                  placeholder="Enter phone number"
                  value={formData.phone}
                  onChange={handleChange}
                />
              </div>
              <div className="form-group col-md-6">
                <label><i className="fas fa-map-marker-alt"></i> Address</label>
                <input
                  type="text"
                  name="address"
                  className="form-control"
                  placeholder="Enter address"
                  value={formData.address}
                  onChange={handleChange}
                />
              </div>
            </div>
            <div className="form-group">
              <label><i className="fas fa-first-aid"></i> Emergency Contact</label>
              <input
                type="text"
                name="emergency_contact"
                className="form-control"
                placeholder="Enter emergency contact"
                value={formData.emergency_contact}
                onChange={handleChange}
              />
            </div>
            <div className="form-group">
              <label><i className="fas fa-notes-medical"></i> Description</label>
              <textarea
                name="description"
                className="form-control"
                placeholder="Enter patient description"
                value={formData.description}
                onChange={handleChange}
              />
            </div>
            {error && <div className="alert alert-danger">{error}</div>}
            <button type="submit" className="btn btn-primary btn-lg btn-block" disabled={loading}>
              {loading ? 'Creating...' : 'Create Patient'}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
};

export default PatientCreate;