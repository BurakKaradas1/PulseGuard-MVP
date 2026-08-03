import React from 'react';

interface SidebarProps {
  onOpenSettings: () => void;
}

export const Sidebar: React.FC<SidebarProps> = ({ onOpenSettings }) => {
  return (
    <aside style={{ width: '260px', flexShrink: 0, borderRight: '1px solid rgba(202, 138, 4, 0.5)', backgroundColor: '#111827', padding: '24px', display: 'flex', flexDirection: 'column', zIndex: 10 }}>
      <h1 style={{ fontSize: '28px', fontWeight: 'bold', color: '#eab308', margin: '0 0 8px 0', textShadow: '0 0 10px rgba(234, 179, 8, 0.5)' }}>PulseGuard</h1>
      <div style={{ borderBottom: '1px solid rgba(202, 138, 4, 0.3)', marginBottom: '24px', marginTop: '16px' }}></div>
      <nav style={{ display: 'flex', flexDirection: 'column', gap: '12px', flex: 1 }}>
        <div className="menu-item" style={{ padding: '12px', borderRadius: '4px', backgroundColor: 'rgba(202, 138, 4, 0.15)', color: '#fef08a', border: '1px solid rgba(202, 138, 4, 0.5)', display: 'flex', alignItems: 'center', gap: '12px' }}>
          <div className="pulse-dot"></div>
          <span style={{ fontWeight: 'bold', letterSpacing: '0.5px' }}>Fleet View</span>
        </div>
        <div className="menu-item" onClick={onOpenSettings} style={{ padding: '12px', borderRadius: '4px', color: '#9ca3af', border: '1px solid transparent', display: 'flex', alignItems: 'center', gap: '12px', cursor: 'pointer' }}>
          <span style={{ fontSize: '16px' }}>⚙️</span>
          <span style={{ fontWeight: 'bold', letterSpacing: '0.5px' }}>System Settings</span>
        </div>
      </nav>
    </aside>
  );
};