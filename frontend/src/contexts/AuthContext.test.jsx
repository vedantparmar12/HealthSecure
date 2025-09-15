import React from 'react';
import { render, screen, act, waitFor } from '@testing-library/react';
import { AuthProvider, useAuth } from './AuthContext';
import { authAPI } from '../services/api';

// Mock the API
jest.mock('../services/api', () => ({
  authAPI: {
    login: jest.fn(),
    refreshToken: jest.fn(),
    logout: jest.fn(),
  }
}));

// Mock localStorage
const localStorageMock = {
  getItem: jest.fn(),
  setItem: jest.fn(),
  removeItem: jest.fn(),
  clear: jest.fn(),
};
global.localStorage = localStorageMock;

// Test component to access context
const TestComponent = () => {
  const { user, login, logout, isLoading, isAdmin, isDoctor, isNurse } = useAuth();
  
  return (
    <div>
      <div data-testid="loading">{isLoading ? 'loading' : 'loaded'}</div>
      <div data-testid="user">{user ? user.name : 'no user'}</div>
      <div data-testid="admin">{isAdmin() ? 'admin' : 'not admin'}</div>
      <div data-testid="doctor">{isDoctor() ? 'doctor' : 'not doctor'}</div>
      <div data-testid="nurse">{isNurse() ? 'nurse' : 'not nurse'}</div>
      <button onClick={() => login('test@example.com', 'password')}>Login</button>
      <button onClick={logout}>Logout</button>
    </div>
  );
};

describe('AuthContext', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorageMock.getItem.mockReturnValue(null);
  });

  test('provides initial state', () => {
    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    expect(screen.getByTestId('loading')).toHaveTextContent('loaded');
    expect(screen.getByTestId('user')).toHaveTextContent('no user');
    expect(screen.getByTestId('admin')).toHaveTextContent('not admin');
    expect(screen.getByTestId('doctor')).toHaveTextContent('not doctor');
    expect(screen.getByTestId('nurse')).toHaveTextContent('not nurse');
  });

  test('loads user from localStorage on mount', async () => {
    const mockUser = {
      id: 1,
      name: 'Test Admin',
      email: 'admin@example.com',
      role: 'admin'
    };

    localStorageMock.getItem.mockImplementation((key) => {
      if (key === 'user') return JSON.stringify(mockUser);
      if (key === 'accessToken') return 'fake-token';
      return null;
    });

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.getByTestId('user')).toHaveTextContent('Test Admin');
      expect(screen.getByTestId('admin')).toHaveTextContent('admin');
    });
  });

  test('handles successful login', async () => {
    const mockUser = {
      id: 1,
      name: 'Test Doctor',
      email: 'doctor@example.com',
      role: 'doctor'
    };

    const mockResponse = {
      data: {
        user: mockUser,
        access_token: 'access-token',
        refresh_token: 'refresh-token'
      }
    };

    authAPI.login.mockResolvedValueOnce(mockResponse);

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    await act(async () => {
      screen.getByText('Login').click();
    });

    await waitFor(() => {
      expect(screen.getByTestId('user')).toHaveTextContent('Test Doctor');
      expect(screen.getByTestId('doctor')).toHaveTextContent('doctor');
    });

    expect(authAPI.login).toHaveBeenCalledWith({
      email: 'test@example.com',
      password: 'password'
    });

    expect(localStorageMock.setItem).toHaveBeenCalledWith('user', JSON.stringify(mockUser));
    expect(localStorageMock.setItem).toHaveBeenCalledWith('accessToken', 'access-token');
    expect(localStorageMock.setItem).toHaveBeenCalledWith('refreshToken', 'refresh-token');
  });

  test('handles login error', async () => {
    authAPI.login.mockRejectedValueOnce(new Error('Invalid credentials'));

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    await act(async () => {
      screen.getByText('Login').click();
    });

    await waitFor(() => {
      expect(screen.getByTestId('user')).toHaveTextContent('no user');
    });
  });

  test('handles logout', async () => {
    const mockUser = {
      id: 1,
      name: 'Test User',
      email: 'test@example.com',
      role: 'nurse'
    };

    localStorageMock.getItem.mockImplementation((key) => {
      if (key === 'user') return JSON.stringify(mockUser);
      if (key === 'accessToken') return 'fake-token';
      return null;
    });

    authAPI.logout.mockResolvedValueOnce({});

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    // Wait for initial load
    await waitFor(() => {
      expect(screen.getByTestId('user')).toHaveTextContent('Test User');
    });

    await act(async () => {
      screen.getByText('Logout').click();
    });

    await waitFor(() => {
      expect(screen.getByTestId('user')).toHaveTextContent('no user');
    });

    expect(authAPI.logout).toHaveBeenCalled();
    expect(localStorageMock.removeItem).toHaveBeenCalledWith('user');
    expect(localStorageMock.removeItem).toHaveBeenCalledWith('accessToken');
    expect(localStorageMock.removeItem).toHaveBeenCalledWith('refreshToken');
  });

  test('role checking functions work correctly', async () => {
    const adminUser = {
      id: 1,
      name: 'Admin User',
      email: 'admin@example.com',
      role: 'admin'
    };

    localStorageMock.getItem.mockImplementation((key) => {
      if (key === 'user') return JSON.stringify(adminUser);
      return null;
    });

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.getByTestId('admin')).toHaveTextContent('admin');
      expect(screen.getByTestId('doctor')).toHaveTextContent('not doctor');
      expect(screen.getByTestId('nurse')).toHaveTextContent('not nurse');
    });
  });

  test('doctor role checking', async () => {
    const doctorUser = {
      id: 1,
      name: 'Doctor User',
      email: 'doctor@example.com',
      role: 'doctor'
    };

    localStorageMock.getItem.mockImplementation((key) => {
      if (key === 'user') return JSON.stringify(doctorUser);
      return null;
    });

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.getByTestId('admin')).toHaveTextContent('not admin');
      expect(screen.getByTestId('doctor')).toHaveTextContent('doctor');
      expect(screen.getByTestId('nurse')).toHaveTextContent('not nurse');
    });
  });

  test('nurse role checking', async () => {
    const nurseUser = {
      id: 1,
      name: 'Nurse User',
      email: 'nurse@example.com',
      role: 'nurse'
    };

    localStorageMock.getItem.mockImplementation((key) => {
      if (key === 'user') return JSON.stringify(nurseUser);
      return null;
    });

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.getByTestId('admin')).toHaveTextContent('not admin');
      expect(screen.getByTestId('doctor')).toHaveTextContent('not doctor');
      expect(screen.getByTestId('nurse')).toHaveTextContent('nurse');
    });
  });
});