'use client';

import { useState, useEffect } from 'react';
import Login from './components/login';
import DashboardLayout from './components/dashboard-layout';

interface UserInfo {
  email: string;
  password: string;
  role: string;
  name?: string;
  user_id?: number;
  token?: string;
}

export default function HomePage() {
  const [user, setUser] = useState<UserInfo | null>(null);

  useEffect(() => {
    const refreshToken = async () => {
      const refreshToken = localStorage.getItem('healthsecure_refresh_token');
      if (!refreshToken) return;

      try {
        const response = await fetch('/api/auth/refresh', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ refresh_token: refreshToken }),
        });

        if (response.ok) {
          const data = await response.json();
          localStorage.setItem('healthsecure_token', data.token);
          setUser(prevUser => prevUser ? { ...prevUser, token: data.token } : null);
        } else {
          handleLogout();
        }
      } catch (error) {
        console.error('Token refresh error:', error);
        handleLogout();
      }
    };

    const interval = setInterval(refreshToken, 14 * 60 * 1000); // Refresh every 14 minutes

    return () => clearInterval(interval);
  }, []);

  const handleLogin = (credentials: UserInfo) => {
    // User is already authenticated by the backend
    setUser(credentials);
  };

  const handleLogout = () => {
    // Clear stored tokens
    localStorage.removeItem('healthsecure_token');
    localStorage.removeItem('healthsecure_refresh_token');
    setUser(null);
  };

  if (!user) {
    return <Login onLogin={handleLogin} />;
  }

  return (
    <DashboardLayout
      user={{
        email: user.email,
        role: user.role,
        name: user.name || user.email.split('@')[0],
        user_id: user.user_id,
        token: user.token
      }}
      onLogout={handleLogout}
    >
      {/* Dashboard content is handled within DashboardLayout */}
    </DashboardLayout>
  );
}