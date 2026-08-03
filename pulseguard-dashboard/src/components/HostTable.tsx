import React from 'react';
import type { HostStatus } from '../types';

interface HostTableProps {
  hosts: HostStatus[];
  selectedHostId?: string;
  onHostClick: (id: string) => void;
}

export const HostTable: React.FC<HostTableProps> = ({ hosts, selectedHostId, onHostClick }) => {
  const renderStatusBadge = (status: string) => {
    let config = { bg: 'rgba(34, 197, 94, 0.25)', border: '#22c55e', text: '#4ade80', shadow: 'rgba(34, 197, 94, 0.4)', label: 'ONLINE' };
    const s = status ? status.toLowerCase() : '';
    if (s === 'warning' || s === 'critical') {
      config = { bg: 'rgba(234, 179, 8, 0.25)', border: '#eab308', text: '#fef08a', shadow: 'rgba(234, 179, 8, 0.4)', label: 'WARNING' };
    } else if (s === 'offline') {
      config = { bg: 'rgba(239, 68, 68, 0.25)', border: '#ef4444', text: '#fca5a5', shadow: 'rgba(239, 68, 68, 0.4)', label: 'OFFLINE' };
    }
    return (
      <span style={{ display: 'inline-block', padding: '6px 14px', borderRadius: '9999px', backgroundColor: config.bg, border: `1px solid ${config.border}`, color: config.text, fontSize: '11px', fontWeight: 'bold', letterSpacing: '1px', boxShadow: `0 0 10px ${config.shadow}` }}>
        {config.label}
      </span>
    );
  };

  return (
    <div style={{ border: '1px solid rgba(202, 138, 4, 0.4)', borderRadius: '8px', backgroundColor: '#000', overflow: 'hidden', boxShadow: '0 0 20px rgba(202, 138, 4, 0.05)' }}>
      <table style={{ width: '100%', textAlign: 'left', borderCollapse: 'collapse' }}>
        <thead>
          <tr style={{ backgroundColor: '#111827', borderBottom: '1px solid rgba(202, 138, 4, 0.5)' }}>
            <th style={{ padding: '16px 12px', fontSize: '12px', color: '#6b7280', textTransform: 'uppercase', letterSpacing: '1px' }}>STATUS</th>
            <th style={{ padding: '16px 12px', fontSize: '12px', color: '#6b7280', textTransform: 'uppercase', letterSpacing: '1px' }}>HOSTNAME</th>
            <th style={{ padding: '16px 12px', fontSize: '12px', color: '#6b7280', textTransform: 'uppercase', letterSpacing: '1px' }}>LAST SEEN</th>
            <th style={{ padding: '16px 12px', fontSize: '12px', color: '#6b7280', textTransform: 'uppercase', letterSpacing: '1px', width: '18%' }}>CPU</th>
            <th style={{ padding: '16px 12px', fontSize: '12px', color: '#6b7280', textTransform: 'uppercase', letterSpacing: '1px', width: '18%' }}>RAM</th>
            <th style={{ padding: '16px 12px', fontSize: '12px', color: '#6b7280', textTransform: 'uppercase', letterSpacing: '1px', width: '18%' }}>DISK</th>
          </tr>
        </thead>
        <tbody>
          {hosts.length === 0 ? (
            <tr>
              <td colSpan={6} style={{ padding: '24px', textAlign: 'center', color: '#6b7280' }}>No registered hosts found or waiting for agent...</td>
            </tr>
          ) : (
            hosts.map(host => {
              const cpu = host.cpu_usage ?? 0;
              const ram = host.ram_usage ?? 0;
              const disk = host.disk_usage ?? 0;
              return (
                <tr 
                  key={host.id} 
                  className={`host-row ${selectedHostId === host.id ? 'selected-row' : ''}`}
                  style={{ borderBottom: '1px solid #111827' }}
                  onClick={() => onHostClick(host.id)} 
                >
                  <td style={{ padding: '16px 12px' }}>{renderStatusBadge(host.status)}</td>
                  <td style={{ padding: '16px 12px', color: '#d1d5db', fontWeight: 'bold' }}>{host.hostname}</td>
                  <td style={{ padding: '16px 12px', color: '#6b7280', fontSize: '12px' }}>{new Date(host.last_seen).toLocaleString('tr-TR')}</td>
                  <td style={{ padding: '16px 12px', color: '#9ca3af' }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                      <span style={{ width: '32px', fontWeight: 'bold', fontSize: '13px' }}>%{cpu}</span>
                      <div style={{ flex: 1, height: '6px', backgroundColor: 'rgba(255, 255, 255, 0.1)', borderRadius: '3px', overflow: 'hidden' }}>
                        <div style={{ width: `${cpu}%`, height: '100%', backgroundColor: cpu > 85 ? '#ef4444' : '#eab308' }}></div>
                      </div>
                    </div>
                  </td>
                  <td style={{ padding: '16px 12px', color: '#9ca3af' }}>
                     <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                      <span style={{ width: '32px', fontWeight: 'bold', fontSize: '13px' }}>%{ram}</span>
                      <div style={{ flex: 1, height: '6px', backgroundColor: 'rgba(255, 255, 255, 0.1)', borderRadius: '3px', overflow: 'hidden' }}>
                        <div style={{ width: `${ram}%`, height: '100%', backgroundColor: ram > 85 ? '#ef4444' : '#eab308' }}></div>
                      </div>
                    </div>
                  </td>
                  <td style={{ padding: '16px 12px', color: '#9ca3af' }}>
                     <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                      <span style={{ width: '32px', fontWeight: 'bold', fontSize: '13px' }}>%{disk}</span>
                      <div style={{ flex: 1, height: '6px', backgroundColor: 'rgba(255, 255, 255, 0.1)', borderRadius: '3px', overflow: 'hidden' }}>
                        <div style={{ width: `${disk}%`, height: '100%', backgroundColor: disk > 85 ? '#ef4444' : '#eab308' }}></div>
                      </div>
                    </div>
                  </td>
                </tr>
              );
            })
          )}
        </tbody>
      </table>
    </div>
  );
};