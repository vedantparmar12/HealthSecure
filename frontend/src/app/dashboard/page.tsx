'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import DashboardLayout from '../components/dashboard-layout';

interface UserInfo {
  email: string;
  password: string;
  role: string;
  name?: string;
  user_id?: number;
  token?: string;
}

export default function DashboardPage() {
  const [user, setUser] = useState<UserInfo | null>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    const checkAuth = async () => {
      const token = localStorage.getItem('healthsecure_token');
      const refreshToken = localStorage.getItem('healthsecure_refresh_token');
      
      if (!token && !refreshToken) {
        router.push('/login');
        return;
      }

      // Try to refresh token if needed
      if (refreshToken) {
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
            
            // Get user info
            const userResponse = await fetch('/api/auth/me', {
              headers: {
                Authorization: `Bearer ${data.token}`,
              },
            });

            if (userResponse.ok) {
              const userData = await userResponse.json();
              setUser({
                email: userData.user.email,
                role: userData.user.role,
                name: userData.user.name,
                user_id: userData.user.id,
                token: data.token,
                password: '', // Not needed
              });
            }
          } else {
            router.push('/login');
          }
        } catch (error) {
          console.error('Auth check error:', error);
          router.push('/login');
        }
      }
      
      setLoading(false);
    };

    checkAuth();

    // Setup token refresh interval
    const interval = setInterval(async () => {
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
        }
      } catch (error) {
        console.error('Token refresh error:', error);
      }
    }, 14 * 60 * 1000); // Refresh every 14 minutes

    return () => clearInterval(interval);
  }, [router]);

  const handleLogout = () => {
    localStorage.removeItem('healthsecure_token');
    localStorage.removeItem('healthsecure_refresh_token');
    setUser(null);
    router.push('/');
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="w-16 h-16 border-4 border-primary border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  if (!user) {
    return null;
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
