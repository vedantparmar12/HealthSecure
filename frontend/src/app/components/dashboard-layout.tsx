'use client';

import { useState, ReactNode, useEffect } from 'react';
import ClientOnly from './client-only';
import AddPatientModal from './add-patient-modal';
import AuditLogViewer from './audit-log-viewer';
import EmergencyAccessModal from './emergency-access-modal';

interface User {
  email: string;
  role: string;
  name?: string;
  user_id?: number;
  token?: string;
}

interface DashboardLayoutProps {
  user: User;
  onLogout: () => void;
  children: ReactNode;
}

type TabType = 'dashboard' | 'patients' | 'emergency' | 'audit' | 'ai-assistant';

export default function DashboardLayout({ user, onLogout, children }: DashboardLayoutProps) {
  const [activeTab, setActiveTab] = useState<TabType>('dashboard');

  const tabs = [
    { id: 'dashboard', label: 'Dashboard', icon: '📊' },
    { id: 'patients', label: 'Patients', icon: '👥', roles: ['doctor', 'nurse'] },
    { id: 'emergency', label: 'Emergency', icon: '🚨', roles: ['doctor', 'nurse'] },
    { id: 'audit', label: 'Audit Logs', icon: '📋' },
    { id: 'ai-assistant', label: 'AI Assistant', icon: '🤖' }
  ] as const;

  const visibleTabs = tabs.filter(tab =>
    !('roles' in tab) || tab.roles.includes(user.role as any)
  );

  const getGreeting = () => {
    const hour = new Date().getHours();
    if (hour < 12) return 'Good morning';
    if (hour < 18) return 'Good afternoon';
    return 'Good evening';
  };

  return (
    <div className="h-screen flex flex-col bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="px-6 py-4">
          <div className="flex justify-between items-center">
            <div>
              <h1 className="text-2xl font-bold text-indigo-900">HealthSecure</h1>
              <p className="text-sm text-gray-600">
                {getGreeting()}, {user.name || user.email}!
              </p>
            </div>

            <div className="flex items-center gap-4">
              <div className="text-right">
                <p className="text-sm font-medium text-gray-900">{user.email}</p>
                <p className="text-xs text-gray-500 capitalize">{user.role}</p>
              </div>
              <button
                onClick={onLogout}
                className="px-4 py-2 text-sm bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-md transition-colors"
              >
                Logout
              </button>
            </div>
          </div>
        </div>

        {/* Navigation Tabs */}
        <div className="px-6">
          <nav className="flex space-x-8">
            {visibleTabs.map((tab) => (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id as TabType)}
                className={`
                  flex items-center gap-2 px-1 py-4 text-sm font-medium border-b-2 transition-colors
                  ${activeTab === tab.id
                    ? 'border-indigo-500 text-indigo-600'
                    : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                  }
                `}
              >
                <span>{tab.icon}</span>
                {tab.label}
              </button>
            ))}
          </nav>
        </div>
      </header>

      {/* Content */}
      <main className="flex-1 overflow-hidden">
        <ClientOnly fallback={<div className="flex items-center justify-center h-full"><div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div></div>}>
          <TabContent activeTab={activeTab} user={user} onTabChange={setActiveTab} />
        </ClientOnly>
      </main>

      {/* HIPAA Compliance Footer */}
      <footer className="bg-blue-50 border-t border-blue-200 px-6 py-3">
        <div className="flex items-center gap-2">
          <span className="text-blue-600">🔒</span>
          <p className="text-xs text-blue-800">
            <strong>HIPAA Compliance:</strong> This system maintains full HIPAA compliance.
            All access is logged and audited. Unauthorized access may result in legal action.
          </p>
        </div>
      </footer>
    </div>
  );
}

interface TabContentProps {
  activeTab: TabType;
  user: User;
  onTabChange: (tab: TabType) => void;
}

function TabContent({ activeTab, user, onTabChange }: TabContentProps) {
  switch (activeTab) {
    case 'dashboard':
      return <DashboardTab user={user} onTabChange={onTabChange} />;
    case 'patients':
      return <PatientsTab user={user} />;
    case 'emergency':
      return <EmergencyTab user={user} />;
    case 'audit':
      return <AuditTab user={user} />;
    case 'ai-assistant':
      return <AIAssistantTab user={user} />;
    default:
      return <DashboardTab user={user} onTabChange={onTabChange} />;
  }
}

function DashboardTab({ user, onTabChange }: { user: User; onTabChange: (tab: TabType) => void }) {
  const [stats, setStats] = useState({
    system_status: 'healthy',
    access_level: user.role === 'doctor' ? 'Full Access' : user.role === 'nurse' ? 'Medical Access' : 'Limited Access',
    recent_activity: new Date().toLocaleDateString()
  });

  const quickActions = [
    ...(user.role === 'doctor' || user.role === 'nurse' ? [
      { title: 'View Patients', description: 'Access patient records', icon: '👥', action: 'patients' },
      { title: 'Emergency Access', description: 'Request emergency access', icon: '🚨', action: 'emergency' },
    ] : []),
    { title: 'Audit Logs', description: 'View system audit trail', icon: '📋', action: 'audit' },
    { title: 'AI Assistant', description: 'Medical knowledge base chat', icon: '🤖', action: 'ai-assistant' },
  ];

  return (
    <div className="p-6">
      <div className="mb-6">
        <h2 className="text-xl font-semibold text-gray-900 mb-2">Dashboard Overview</h2>
        <p className="text-gray-600">
          Welcome to the HealthSecure medical data management system
        </p>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center">
            <div className="text-2xl mr-3">👤</div>
            <div>
              <p className="text-sm font-medium text-gray-600">Your Role</p>
              <p className={`text-lg font-semibold capitalize ${
                user.role === 'doctor' ? 'text-green-600' :
                user.role === 'nurse' ? 'text-blue-600' : 'text-gray-600'
              }`}>
                {user.role}
              </p>
            </div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center">
            <div className="text-2xl mr-3">🔐</div>
            <div>
              <p className="text-sm font-medium text-gray-600">Access Level</p>
              <p className="text-lg font-semibold text-gray-900">{stats.access_level}</p>
            </div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-lg shadow-sm border">
          <div className="flex items-center">
            <div className="text-2xl mr-3">
              {stats.system_status === 'healthy' ? '✅' : '⚠️'}
            </div>
            <div>
              <p className="text-sm font-medium text-gray-600">System Status</p>
              <p className={`text-lg font-semibold ${
                stats.system_status === 'healthy' ? 'text-green-600' : 'text-red-600'
              }`}>
                {stats.system_status === 'healthy' ? 'Healthy' : 'Warning'}
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="bg-white rounded-lg shadow-sm border">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900">Quick Actions</h3>
        </div>
        <div className="p-6">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {quickActions.map((action, index) => (
              <button
                key={index}
                className="p-4 border border-gray-200 rounded-lg hover:shadow-md transition-shadow text-left"
                onClick={() => onTabChange(action.action as TabType)}
              >
                <div className="text-3xl mb-3">{action.icon}</div>
                <h4 className="font-medium text-gray-900 mb-1">{action.title}</h4>
                <p className="text-sm text-gray-600">{action.description}</p>
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

function PatientsTab({ user }: { user: User }) {
  const [isModalOpen, setIsModalOpen] = useState(false);

  const handlePatientAdded = () => {
    // Here you might want to refresh the patient list
    alert('Patient added successfully!');
  };

  return (
    <div className="p-6">
      <div className="bg-white rounded-lg shadow-sm border">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900">Patient Management</h3>
        </div>
        <div className="p-6">
          <div className="text-center py-8">
            <div className="text-6xl mb-4">👥</div>
            <h3 className="text-xl font-medium text-gray-900 mb-2">Patient Management System</h3>
            <p className="text-gray-600 mb-6">
              Access and manage patient records, medical history, and treatment plans.
            </p>
            <button 
              onClick={() => setIsModalOpen(true)}
              className="bg-indigo-600 text-white px-6 py-2 rounded-md hover:bg-indigo-700 transition-colors"
            >
              Add New Patient
            </button>
          </div>
        </div>
      </div>
      {isModalOpen && (
        <AddPatientModal
          user={user}
          onClose={() => setIsModalOpen(false)}
          onPatientAdded={handlePatientAdded}
        />
      )}
    </div>
  );
}

function EmergencyTab({ user }: { user: User }) {
  const [isModalOpen, setIsModalOpen] = useState(false);

  const handleAccessRequested = () => {
    alert('Emergency access requested successfully!');
  };

  return (
    <div className="p-6">
      <div className="bg-red-50 rounded-lg border border-red-200">
        <div className="px-6 py-4 border-b border-red-200">
          <h3 className="text-lg font-medium text-red-900">Emergency Access</h3>
        </div>
        <div className="p-6">
          <div className="text-center py-8">
            <div className="text-6xl mb-4">🚨</div>
            <h3 className="text-xl font-medium text-red-900 mb-2">Emergency Break-Glass Access</h3>
            <p className="text-red-700 mb-6">
              Request immediate access to patient records in emergency situations.
              All emergency access is logged and audited.
            </p>
            <button 
              onClick={() => setIsModalOpen(true)}
              className="bg-red-600 text-white px-6 py-2 rounded-md hover:bg-red-700 transition-colors"
            >
              Request Emergency Access
            </button>
          </div>
        </div>
      </div>
      {isModalOpen && (
        <EmergencyAccessModal
          user={user}
          onClose={() => setIsModalOpen(false)}
          onAccessRequested={handleAccessRequested}
        />
      )}
    </div>
  );
}

function AuditTab({ user }: { user: User }) {
  const [isViewerOpen, setIsViewerOpen] = useState(false);

  return (
    <div className="p-6">
      <div className="bg-white rounded-lg shadow-sm border">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900">Audit Logs</h3>
        </div>
        <div className="p-6">
          <div className="text-center py-8">
            <div className="text-6xl mb-4">📋</div>
            <h3 className="text-xl font-medium text-gray-900 mb-2">System Audit Trail</h3>
            <p className="text-gray-600 mb-6">
              View comprehensive logs of all system access, modifications, and security events.
              HIPAA compliant audit trail for compliance reporting.
            </p>
            <button 
              onClick={() => setIsViewerOpen(true)}
              className="bg-gray-600 text-white px-6 py-2 rounded-md hover:bg-gray-700 transition-colors"
            >
              View Audit Logs
            </button>
          </div>
        </div>
      </div>
      {isViewerOpen && (
        <AuditLogViewer
          user={user}
          onClose={() => setIsViewerOpen(false)}
        />
      )}
    </div>
  );
}

function AIAssistantTab({ user }: { user: User }) {
  return (
    <div className="h-full">
      <AIAssistantContent user={user} />
    </div>
  );
}

function AIAssistantContent({ user }: { user: User }) {
  // Use simpler chat component instead of complex CopilotKit
  const [SimpleChat, setSimpleChat] = useState<any>(null);

  useEffect(() => {
    // Lazy load the simple chat component
    import('./simple-chat').then(module => {
      setSimpleChat(() => module.default);
    }).catch(err => {
      console.error('Failed to load chat component:', err);
    });
  }, []);

  if (!SimpleChat) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading AI Assistant...</p>
        </div>
      </div>
    );
  }

  return (
    <SimpleChat
      user={{ email: user.email, role: user.role }}
    />
  );
}