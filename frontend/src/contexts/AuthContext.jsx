import React, { createContext, useContext, useReducer, useEffect } from 'react';
import { authAPI, tokenManager, handleAPIError } from '../services/api';

// Auth context
const AuthContext = createContext();

// Auth actions
const AUTH_ACTIONS = {
  LOGIN_START: 'LOGIN_START',
  LOGIN_SUCCESS: 'LOGIN_SUCCESS',
  LOGIN_FAILURE: 'LOGIN_FAILURE',
  LOGOUT: 'LOGOUT',
  SET_USER: 'SET_USER',
  SET_LOADING: 'SET_LOADING',
};

// Auth reducer
const authReducer = (state, action) => {
  switch (action.type) {
    case AUTH_ACTIONS.LOGIN_START:
      return {
        ...state,
        loading: true,
        error: null,
      };
    case AUTH_ACTIONS.LOGIN_SUCCESS:
      return {
        ...state,
        user: action.payload.user,
        isAuthenticated: true,
        loading: false,
        error: null,
      };
    case AUTH_ACTIONS.LOGIN_FAILURE:
      return {
        ...state,
        user: null,
        isAuthenticated: false,
        loading: false,
        error: action.payload,
      };
    case AUTH_ACTIONS.LOGOUT:
      return {
        ...state,
        user: null,
        isAuthenticated: false,
        loading: false,
        error: null,
      };
    case AUTH_ACTIONS.SET_USER:
      return {
        ...state,
        user: action.payload,
        isAuthenticated: !!action.payload,
        loading: false,
      };
    case AUTH_ACTIONS.SET_LOADING:
      return {
        ...state,
        loading: action.payload,
      };
    default:
      return state;
  }
};

// Initial state
const initialState = {
  user: null,
  isAuthenticated: false,
  loading: true,
  error: null,
};

// Auth provider component
export const AuthProvider = ({ children }) => {
  const [state, dispatch] = useReducer(authReducer, initialState);

  // Check for existing session on mount
  useEffect(() => {
    checkAuthStatus();
  }, []);

  const checkAuthStatus = async () => {
    try {
      const token = tokenManager.getAccessToken();
      if (!token) {
        dispatch({ type: AUTH_ACTIONS.SET_LOADING, payload: false });
        return;
      }

      const response = await authAPI.getCurrentUser();
      dispatch({
        type: AUTH_ACTIONS.SET_USER,
        payload: response.data.user,
      });
    } catch (error) {
      // Token is invalid or expired
      tokenManager.clearTokens();
      dispatch({ type: AUTH_ACTIONS.SET_LOADING, payload: false });
    }
  };

  const login = async (credentials) => {
    try {
      dispatch({ type: AUTH_ACTIONS.LOGIN_START });

      const response = await authAPI.login(credentials);
      const { access_token, refresh_token, user } = response.data;

      // Store tokens
      tokenManager.setTokens(access_token, refresh_token);

      dispatch({
        type: AUTH_ACTIONS.LOGIN_SUCCESS,
        payload: { user },
      });

      return { success: true };
    } catch (error) {
      const errorMessage = handleAPIError(error);
      dispatch({
        type: AUTH_ACTIONS.LOGIN_FAILURE,
        payload: errorMessage,
      });
      return { success: false, error: errorMessage };
    }
  };

  const logout = async () => {
    try {
      await authAPI.logout();
    } catch (error) {
      // Ignore logout errors - still clear local state
      console.error('Logout error:', error);
    } finally {
      tokenManager.clearTokens();
      dispatch({ type: AUTH_ACTIONS.LOGOUT });
    }
  };

  const updateUser = (updatedUser) => {
    dispatch({
      type: AUTH_ACTIONS.SET_USER,
      payload: updatedUser,
    });
  };

  const clearError = () => {
    dispatch({
      type: AUTH_ACTIONS.LOGIN_FAILURE,
      payload: null,
    });
  };

  // Helper functions for role checking
  const isAdmin = () => state.user?.role === 'admin';
  const isDoctor = () => state.user?.role === 'doctor';
  const isNurse = () => state.user?.role === 'nurse';
  const isMedicalStaff = () => isDoctor() || isNurse();
  const hasRole = (role) => state.user?.role === role;
  const hasAnyRole = (roles) => roles.includes(state.user?.role);

  const value = {
    ...state,
    login,
    logout,
    updateUser,
    clearError,
    // Role helpers
    isAdmin,
    isDoctor,
    isNurse,
    isMedicalStaff,
    hasRole,
    hasAnyRole,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

// Custom hook to use auth context
export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

// HOC for role-based access
export const withAuth = (WrappedComponent, requiredRole = null) => {
  return (props) => {
    const { user, loading } = useAuth();

    if (loading) {
      return (
        <div className="loading">
          <div className="spinner"></div>
        </div>
      );
    }

    if (!user) {
      return <div>Access denied - please log in</div>;
    }

    if (requiredRole && !user.role === requiredRole) {
      return <div>Access denied - insufficient permissions</div>;
    }

    return <WrappedComponent {...props} />;
  };
};

export default AuthContext;