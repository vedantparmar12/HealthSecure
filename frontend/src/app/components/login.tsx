'use client';

import { useState, useEffect } from 'react';

interface LoginProps {
  onLogin: (credentials: { email: string; password: string; role: string; name?: string; user_id?: number; token?: string }) => void;
}

export default function Login({ onLogin }: LoginProps) {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState('doctor');
  const [isLoading, setIsLoading] = useState(false);
  const [roleDefaults, setRoleDefaults] = useState<Record<string, {email: string, password: string}>>({});

  // Fetch default users from database on component mount
  useEffect(() => {
    const fetchDefaults = async () => {
      try {
        const response = await fetch('http://localhost:8080/api/auth/defaults');
        if (response.ok) {
          const data = await response.json();
          setRoleDefaults(data.defaults);

          // Set initial values for doctor role
          if (data.defaults.doctor) {
            setEmail(data.defaults.doctor.email);
            setPassword(data.defaults.doctor.password);
          }
        }
      } catch (error) {
        console.error('Failed to fetch default users:', error);
        // Fallback to hardcoded defaults if API fails
        const fallbackDefaults = {
          doctor: { email: 'dr.smith@hospital.local', password: 'Doctor123' },
          nurse: { email: 'nurse.wilson@hospital.local', password: 'Doctor123' },
          admin: { email: 'admin@healthsecure.local', password: 'Doctor123' }
        };
        setRoleDefaults(fallbackDefaults);
        setEmail(fallbackDefaults.doctor.email);
        setPassword(fallbackDefaults.doctor.password);
      }
    };

    fetchDefaults();
  }, []);

  // Update email and password when role changes
  const handleRoleChange = (newRole: string) => {
    setRole(newRole);
    const defaults = roleDefaults[newRole];
    if (defaults) {
      setEmail(defaults.email);
      setPassword(defaults.password);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      // Connect to Go backend for authentication
      const response = await fetch('http://localhost:8080/api/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email,
          password,
        }),
      });

      if (response.ok) {
        const data = await response.json();
        // Store JWT token
        localStorage.setItem('healthsecure_token', data.token);
        localStorage.setItem('healthsecure_refresh_token', data.refresh_token);

        // Pass user data to parent component
        onLogin({
          email,
          password,
          role: data.user.role,
          name: data.user.name,
          user_id: data.user.id,
          token: data.token
        });
      } else {
        const errorData = await response.json();
        alert(`Login failed: ${errorData.message || 'Invalid credentials'}`);
      }
    } catch (error) {
      console.error('Login error:', error);
      alert('Login failed: Network error. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-indigo-100">
      <div className="max-w-md w-full space-y-8 p-8 bg-white rounded-xl shadow-lg">
        <div className="text-center">
          <h2 className="mt-6 text-3xl font-extrabold text-gray-900">
            HealthSecure AI
          </h2>
          <p className="mt-2 text-sm text-gray-600">
            Sign in to access the Knowledge Base Explorer
          </p>
        </div>

        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          <div className="space-y-4">
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-gray-700">
                Email
              </label>
              <input
                id="email"
                name="email"
                type="email"
                autoComplete="email"
                required
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="mt-1 relative block w-full px-3 py-2 border border-gray-300 rounded-md placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                placeholder="Enter your email"
              />
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-gray-700">
                Password
              </label>
              <input
                id="password"
                name="password"
                type="password"
                autoComplete="current-password"
                required
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="mt-1 relative block w-full px-3 py-2 border border-gray-300 rounded-md placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                placeholder="Enter your password"
              />
            </div>

            <div>
              <label htmlFor="role" className="block text-sm font-medium text-gray-700">
                Role
              </label>
              <select
                id="role"
                name="role"
                value={role}
                onChange={(e) => handleRoleChange(e.target.value)}
                className="mt-1 relative block w-full px-3 py-2 border border-gray-300 rounded-md text-gray-900 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
              >
                <option value="doctor">Doctor</option>
                <option value="nurse">Nurse</option>
                <option value="admin">Administrator</option>
              </select>
            </div>
          </div>

          <div>
            <button
              type="submit"
              disabled={isLoading}
              className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? (
                <div className="flex items-center">
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Signing in...
                </div>
              ) : (
                'Sign in'
              )}
            </button>
          </div>

          <div className="text-center">
            <p className="text-xs text-gray-500">
              Demo credentials: dr.smith@hospital.local / Doctor123 (doctor)<br/>
              nurse.jane@hospital.local / Nurse123 (nurse)<br/>
              admin@hospital.local / admin123 (admin)
            </p>
          </div>
        </form>
      </div>
    </div>
  );
}