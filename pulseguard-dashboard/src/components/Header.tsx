import React from 'react';

interface HeaderProps {
  apiStatus: "connected" | "disconnected";
  apiBaseUrl: string;
}

export const Header: React.FC<HeaderProps> = ({ apiStatus, apiBaseUrl }) => {
  return (
    <header style={{ height: '64px', flexShrink: 0, borderBottom: '1px solid rgba(202, 138, 4, 0.5)', display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0 32px', backgroundColor: 'rgba(17, 24, 39, 0.5)' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
        {apiStatus === "connected" && (
          <>
            <div style={{ width: '12px', height: '12px', backgroundColor: '#3b82f6', borderRadius: '50%', boxShadow: '0 0 8px #3b82f6' }}></div>
            <span style={{ fontSize: '14px', fontWeight: 'bold', color: '#3b82f6', textShadow: '0 0 5px rgba(59, 130, 246, 0.5)' }}>
              LIVE CONNECTION ACTIVE
            </span>
          </>
        )}
      </div>
      <div style={{ fontSize: '12px', color: '#6b7280' }}>
        {apiStatus === "connected" ? `REST API: ${apiBaseUrl || 'Local Proxy'}` : "Waiting..."}
      </div>
    </header>
  );
};