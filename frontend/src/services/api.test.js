import axios from 'axios';
import { authAPI, patientsAPI, emergencyAPI, auditAPI, handleAPIError } from './api';

// Mock axios
jest.mock('axios');
const mockedAxios = axios;

// Mock localStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
};
global.localStorage = localStorageMock;

describe('API Service', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorageMock.getItem.mockReturnValue('fake-token');
  });

  describe('authAPI', () => {
    test('login makes correct API call', async () => {
      const mockResponse = {
        data: {
          user: { id: 1, name: 'Test User' },
          access_token: 'token',
          refresh_token: 'refresh'
        }
      };

      mockedAxios.post.mockResolvedValueOnce(mockResponse);

      const result = await authAPI.login({
        email: 'test@example.com',
        password: 'password'
      });

      expect(mockedAxios.post).toHaveBeenCalledWith('/auth/login', {
        email: 'test@example.com',
        password: 'password'
      });
      expect(result).toEqual(mockResponse);
    });

    test('refreshToken makes correct API call', async () => {
      const mockResponse = {
        data: {
          access_token: 'new-token',
          refresh_token: 'new-refresh'
        }
      };

      mockedAxios.post.mockResolvedValueOnce(mockResponse);

      const result = await authAPI.refreshToken({
        refresh_token: 'old-refresh'
      });

      expect(mockedAxios.post).toHaveBeenCalledWith('/auth/refresh', {
        refresh_token: 'old-refresh'
      });
      expect(result).toEqual(mockResponse);
    });

    test('logout makes correct API call', async () => {
      const mockResponse = { data: { message: 'Logged out' } };

      mockedAxios.post.mockResolvedValueOnce(mockResponse);

      const result = await authAPI.logout();

      expect(mockedAxios.post).toHaveBeenCalledWith('/auth/logout');
      expect(result).toEqual(mockResponse);
    });

    test('changePassword makes correct API call', async () => {
      const mockResponse = { data: { message: 'Password changed' } };

      mockedAxios.post.mockResolvedValueOnce(mockResponse);

      const result = await authAPI.changePassword({
        current_password: 'old',
        new_password: 'new'
      });

      expect(mockedAxios.post).toHaveBeenCalledWith('/auth/change-password', {
        current_password: 'old',
        new_password: 'new'
      });
      expect(result).toEqual(mockResponse);
    });
  });

  describe('patientsAPI', () => {
    test('getPatients makes correct API call', async () => {
      const mockResponse = {
        data: {
          patients: [{ id: 1, name: 'Patient 1' }],
          pagination: { total: 1 }
        }
      };

      mockedAxios.get.mockResolvedValueOnce(mockResponse);

      const result = await patientsAPI.getPatients({
        page: 1,
        limit: 20
      });

      expect(mockedAxios.get).toHaveBeenCalledWith('/patients', {
        params: { page: 1, limit: 20 }
      });
      expect(result).toEqual(mockResponse);
    });

    test('getPatient makes correct API call', async () => {
      const mockResponse = {
        data: {
          patient: { id: 1, name: 'Patient 1' }
        }
      };

      mockedAxios.get.mockResolvedValueOnce(mockResponse);

      const result = await patientsAPI.getPatient(1);

      expect(mockedAxios.get).toHaveBeenCalledWith('/patients/1', {
        params: {}
      });
      expect(result).toEqual(mockResponse);
    });

    test('getPatient with emergency token', async () => {
      const mockResponse = {
        data: {
          patient: { id: 1, name: 'Patient 1' }
        }
      };

      mockedAxios.get.mockResolvedValueOnce(mockResponse);

      const result = await patientsAPI.getPatient(1, 'emergency-token');

      expect(mockedAxios.get).toHaveBeenCalledWith('/patients/1', {
        params: { emergency_token: 'emergency-token' }
      });
      expect(result).toEqual(mockResponse);
    });

    test('createPatient makes correct API call', async () => {
      const mockResponse = {
        data: {
          patient: { id: 1, name: 'New Patient' }
        }
      };

      const patientData = {
        first_name: 'John',
        last_name: 'Doe'
      };

      mockedAxios.post.mockResolvedValueOnce(mockResponse);

      const result = await patientsAPI.createPatient(patientData);

      expect(mockedAxios.post).toHaveBeenCalledWith('/patients', patientData);
      expect(result).toEqual(mockResponse);
    });

    test('updatePatient makes correct API call', async () => {
      const mockResponse = {
        data: {
          patient: { id: 1, name: 'Updated Patient' }
        }
      };

      const updateData = {
        first_name: 'Jane'
      };

      mockedAxios.put.mockResolvedValueOnce(mockResponse);

      const result = await patientsAPI.updatePatient(1, updateData);

      expect(mockedAxios.put).toHaveBeenCalledWith('/patients/1', updateData);
      expect(result).toEqual(mockResponse);
    });

    test('searchPatients makes correct API call', async () => {
      const mockResponse = {
        data: {
          patients: [{ id: 1, name: 'John Doe' }]
        }
      };

      mockedAxios.get.mockResolvedValueOnce(mockResponse);

      const result = await patientsAPI.searchPatients('John');

      expect(mockedAxios.get).toHaveBeenCalledWith('/patients/search', {
        params: { q: 'John' }
      });
      expect(result).toEqual(mockResponse);
    });
  });

  describe('emergencyAPI', () => {
    test('requestAccess makes correct API call', async () => {
      const mockResponse = {
        data: {
          access: { id: 1, patient_id: 1 }
        }
      };

      const requestData = {
        patient_id: 1,
        reason: 'Emergency situation'
      };

      mockedAxios.post.mockResolvedValueOnce(mockResponse);

      const result = await emergencyAPI.requestAccess(requestData);

      expect(mockedAxios.post).toHaveBeenCalledWith('/emergency/request', requestData);
      expect(result).toEqual(mockResponse);
    });

    test('getActiveAccess makes correct API call', async () => {
      const mockResponse = {
        data: {
          active_sessions: [{ id: 1, user_id: 1 }]
        }
      };

      mockedAxios.get.mockResolvedValueOnce(mockResponse);

      const result = await emergencyAPI.getActiveAccess();

      expect(mockedAxios.get).toHaveBeenCalledWith('/emergency/active');
      expect(result).toEqual(mockResponse);
    });

    test('getUserAccess makes correct API call', async () => {
      const mockResponse = {
        data: {
          emergency_access: [{ id: 1, user_id: 1 }]
        }
      };

      mockedAxios.get.mockResolvedValueOnce(mockResponse);

      const result = await emergencyAPI.getUserAccess(1);

      expect(mockedAxios.get).toHaveBeenCalledWith('/emergency/user/1');
      expect(result).toEqual(mockResponse);
    });

    test('revokeAccess makes correct API call', async () => {
      const mockResponse = {
        data: { message: 'Access revoked' }
      };

      mockedAxios.delete.mockResolvedValueOnce(mockResponse);

      const result = await emergencyAPI.revokeAccess(1);

      expect(mockedAxios.delete).toHaveBeenCalledWith('/emergency/1');
      expect(result).toEqual(mockResponse);
    });
  });

  describe('auditAPI', () => {
    test('getLogs makes correct API call', async () => {
      const mockResponse = {
        data: {
          audit_logs: [{ id: 1, action: 'LOGIN' }],
          pagination: { total: 1 }
        }
      };

      const filters = {
        action: 'LOGIN',
        page: 1,
        limit: 50
      };

      mockedAxios.get.mockResolvedValueOnce(mockResponse);

      const result = await auditAPI.getLogs(filters);

      expect(mockedAxios.get).toHaveBeenCalledWith('/audit', {
        params: filters
      });
      expect(result).toEqual(mockResponse);
    });
  });

  describe('handleAPIError', () => {
    test('handles error with response data message', () => {
      const error = {
        response: {
          data: {
            message: 'Custom error message'
          }
        }
      };

      const result = handleAPIError(error);
      expect(result).toBe('Custom error message');
    });

    test('handles error with response data error', () => {
      const error = {
        response: {
          data: {
            error: 'Custom error'
          }
        }
      };

      const result = handleAPIError(error);
      expect(result).toBe('Custom error');
    });

    test('handles error with status code', () => {
      const error = {
        response: {
          status: 404
        }
      };

      const result = handleAPIError(error);
      expect(result).toBe('Request failed with status 404');
    });

    test('handles network error', () => {
      const error = {
        request: {}
      };

      const result = handleAPIError(error);
      expect(result).toBe('Network error - please check your connection');
    });

    test('handles generic error', () => {
      const error = {
        message: 'Something went wrong'
      };

      const result = handleAPIError(error);
      expect(result).toBe('Something went wrong');
    });

    test('handles unknown error', () => {
      const error = {};

      const result = handleAPIError(error);
      expect(result).toBe('An unexpected error occurred');
    });
  });
});