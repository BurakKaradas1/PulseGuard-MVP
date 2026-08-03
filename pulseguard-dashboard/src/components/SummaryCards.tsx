import React from 'react';
import type { HostStatus } from '../types';

interface SummaryCardsProps {
  hosts: HostStatus[];
}

export const SummaryCards: React.FC<SummaryCardsProps> = ({ hosts }) => {
  const offlineCount = hosts.filter(h => h.status && h.status.toLowerCase() === 'offline').length;
  const avgCpu = hosts.length > 0 ? Math.round(hosts.reduce((acc, h) => acc + (h.cpu_usage || 0), 0) / hosts.length) : 0;
  const avgRam = hosts.length > 0 ? Math.round(hosts.reduce((acc, h) => acc + (h.ram_usage || 0), 0) / hosts.length) : 0;

  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, minmax(0, 1fr))', gap: '24px', marginBottom: '32px' }}>
      <div style={{ border: '1px solid rgba(202, 138, 4, 0.4)', padding: '24px', borderRadius: '8px', backgroundColor: 'rgba(17, 24, 39, 0.4)', boxShadow: '0 0 15px rgba(202, 138, 4, 0.1)', display: 'flex', flexDirection: 'column', gap: '8px' }}>
        <span style={{ fontSize: '13px', color: '#9ca3af', textTransform: 'uppercase', letterSpacing: '1px' }}>TOTAL HOSTS</span>
        <span style={{ fontSize: '48px', fontWeight: 'bold', color: '#fff', lineHeight: '1', textShadow: '0 0 15px rgba(255,255,255,0.3)' }}>{hosts.length}</span>
      </div>
      <div style={{ border: '1px solid rgba(202, 138, 4, 0.4)', padding: '24px', borderRadius: '8px', backgroundColor: 'rgba(17, 24, 39, 0.4)', boxShadow: '0 0 15px rgba(202, 138, 4, 0.1)', display: 'flex', flexDirection: 'column', gap: '8px' }}>
        <span style={{ fontSize: '13px', color: '#9ca3af', textTransform: 'uppercase', letterSpacing: '1px' }}>AVG CPU / RAM</span>
        {/* Yüzde işaretleri sayının sonuna alındı: {avgCpu}% / {avgRam}% */}
        <span style={{ fontSize: '36px', fontWeight: 'bold', color: '#fff', lineHeight: '1', textShadow: '0 0 15px rgba(255,255,255,0.3)' }}>{avgCpu}% / {avgRam}%</span>
      </div>
      <div style={{ border: '1px solid rgba(239, 68, 68, 0.5)', padding: '24px', borderRadius: '8px', backgroundColor: 'rgba(69, 10, 10, 0.2)', boxShadow: '0 0 25px rgba(239, 68, 68, 0.3)', display: 'flex', flexDirection: 'column', gap: '8px' }}>
        <span style={{ fontSize: '13px', color: '#fca5a5', textTransform: 'uppercase', letterSpacing: '1px' }}>OFFLINE HOSTS</span>
        <span style={{ fontSize: '48px', fontWeight: 'bold', color: '#f87171', lineHeight: '1', textShadow: '0 0 15px rgba(248, 113, 113, 0.6)' }}>{offlineCount}</span>
      </div>
    </div>
  );
};