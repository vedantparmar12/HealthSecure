import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import Login from './Login';
import { AuthProvider } from '../contexts/AuthContext';
import { authAPI } from '../services/api';

// Mock the API
jest.mock('../services/api', () => ({
  authAPI: {
    login: jest.fn(),
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

// Mock react-router-dom navigate
const mockNavigate = jest.fn();
jest.mock('react-router-dom', () => ({
  ...jest.requireActual('react-router-dom'),
  useNavigate: () => mockNavigate,
}));

const renderLogin = () => {
  return render(
    <MemoryRouter>
      <AuthProvider>
        <Login />
      </AuthProvider>
    </MemoryRouter>
  );
};

describe('Login Component', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    localStorageMock.getItem.mockReturnValue(null);
  });

  test('renders login form', () => {
    renderLogin();

    expect(screen.getByText('HealthSecure Login')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Enter your email')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Enter your password')).toBeInTheDocument();
    expect(screen.getByText('🔐 Sign In')).toBeInTheDocument();
  });

  test('renders demo credentials', () => {
    renderLogin();

    expect(screen.getByText('Demo Credentials')).toBeInTheDocument();
    expect(screen.getByText('doctor@healthsecure.com')).toBeInTheDocument();
    expect(screen.getByText('nurse@healthsecure.com')).toBeInTheDocument();
    expect(screen.getByText('admin@healthsecure.com')).toBeInTheDocument();
  });

  test('handles input changes', () => {
    renderLogin();

    const emailInput = screen.getByPlaceholderText('Enter your email');
    const passwordInput = screen.getByPlaceholderText('Enter your password');

    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });

    expect(emailInput.value).toBe('test@example.com');
    expect(passwordInput.value).toBe('password123');
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

    renderLogin();

    const emailInput = screen.getByPlaceholderText('Enter your email');
    const passwordInput = screen.getByPlaceholderText('Enter your password');
    const loginButton = screen.getByText('🔐 Sign In');

    fireEvent.change(emailInput, { target: { value: 'doctor@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.click(loginButton);

    await waitFor(() => {
      expect(authAPI.login).toHaveBeenCalledWith({
        email: 'doctor@example.com',
        password: 'password123'
      });
      expect(mockNavigate).toHaveBeenCalledWith('/dashboard');
    });
  });

  test('handles login error', async () => {
    authAPI.login.mockRejectedValueOnce({
      response: {
        data: {
          message: 'Invalid credentials'
        }
      }
    });

    renderLogin();

    const emailInput = screen.getByPlaceholderText('Enter your email');
    const passwordInput = screen.getByPlaceholderText('Enter your password');
    const loginButton = screen.getByText('🔐 Sign In');

    fireEvent.change(emailInput, { target: { value: 'wrong@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'wrongpassword' } });
    fireEvent.click(loginButton);

    await waitFor(() => {
      expect(screen.getByText('⚠️ Invalid credentials')).toBeInTheDocument();
    });
  });

  test('prevents form submission with empty fields', () => {
    renderLogin();

    const loginButton = screen.getByText('🔐 Sign In');
    fireEvent.click(loginButton);

    // Form should not submit (no API call should be made)
    expect(authAPI.login).not.toHaveBeenCalled();
  });

  test('shows loading state during login', async () => {
    authAPI.login.mockImplementation(() => new Promise(() => {})); // Never resolves

    renderLogin();

    const emailInput = screen.getByPlaceholderText('Enter your email');
    const passwordInput = screen.getByPlaceholderText('Enter your password');
    const loginButton = screen.getByText('🔐 Sign In');

    fireEvent.change(emailInput, { target: { value: 'doctor@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.click(loginButton);

    await waitFor(() => {
      expect(screen.getByText('Signing in...')).toBeInTheDocument();
      expect(loginButton).toBeDisabled();
    });
  });

  test('demo credential buttons fill form', () => {
    renderLogin();

    const emailInput = screen.getByPlaceholderText('Enter your email');
    const passwordInput = screen.getByPlaceholderText('Enter your password');

    // Click doctor demo button
    const doctorButton = screen.getByText('👨‍⚕️ Doctor Access');
    fireEvent.click(doctorButton);

    expect(emailInput.value).toBe('doctor@healthsecure.com');
    expect(passwordInput.value).toBe('doctor123');

    // Click nurse demo button
    const nurseButton = screen.getByText('👩‍⚕️ Nurse Access');
    fireEvent.click(nurseButton);

    expect(emailInput.value).toBe('nurse@healthsecure.com');
    expect(passwordInput.value).toBe('nurse123');

    // Click admin demo button
    const adminButton = screen.getByText('👨‍💼 Admin Access');
    fireEvent.click(adminButton);

    expect(emailInput.value).toBe('admin@healthsecure.com');
    expect(passwordInput.value).toBe('admin123');
  });

  test('clears error when typing', async () => {
    authAPI.login.mockRejectedValueOnce({
      response: {
        data: {
          message: 'Invalid credentials'
        }
      }
    });

    renderLogin();

    const emailInput = screen.getByPlaceholderText('Enter your email');
    const passwordInput = screen.getByPlaceholderText('Enter your password');
    const loginButton = screen.getByText('🔐 Sign In');

    // Trigger error
    fireEvent.change(emailInput, { target: { value: 'wrong@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'wrongpassword' } });
    fireEvent.click(loginButton);

    await waitFor(() => {
      expect(screen.getByText('⚠️ Invalid credentials')).toBeInTheDocument();
    });

    // Error should clear when typing
    fireEvent.change(emailInput, { target: { value: 'new@example.com' } });

    expect(screen.queryByText('⚠️ Invalid credentials')).not.toBeInTheDocument();
  });
});