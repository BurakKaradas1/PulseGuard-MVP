import React from 'react';
import type { HostStatus } from '../types';

interface ExportModalProps {
  isOpen: boolean;
  onClose: () => void;
  hosts: HostStatus[];
}

export const ExportModal: React.FC<ExportModalProps> = ({ isOpen, onClose, hosts }) => {
  if (!isOpen) return null;

  const handleExport = () => {
    const headers = "ID,Host Adi,Durum,CPU,RAM,DISK,Son Gorulme\n";
    const rows = hosts.map(h => `${h.id},${h.hostname},${h.status},${h.cpu_usage || 0},${h.ram_usage || 0},${h.disk_usage || 0},${h.last_seen}`).join("\n");
    const blob = new Blob([headers + rows], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', 'PulseGuard_Rapor.csv');
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    onClose();
  };

  return (
    <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, backgroundColor: 'rgba(0, 0, 0, 0.8)', zIndex: 100, display: 'flex', justifyContent: 'center', alignItems: 'center', backdropFilter: 'blur(4px)' }}>
      <div style={{ backgroundColor: '#111827', border: '1px solid rgba(202, 138, 4, 0.5)', borderRadius: '8px', width: '450px', boxShadow: '0 0 40px rgba(202, 138, 4, 0.1)', overflow: 'hidden' }}>
        <div style={{ padding: '20px 24px', borderBottom: '1px solid rgba(202, 138, 4, 0.3)', display: 'flex', justifyContent: 'space-between', alignItems: 'center', backgroundColor: '#030712' }}>
          <h2 style={{ margin: 0, fontSize: '18px', color: '#facc15' }}>📥 Veri Dışa Aktarımı</h2>
          <button onClick={onClose} style={{ background: 'none', border: 'none', color: '#9ca3af', fontSize: '20px', cursor: 'pointer' }}>×</button>
        </div>
        <div style={{ padding: '24px' }}>
          <p style={{ color: '#9ca3af', fontSize: '14px', marginBottom: '24px' }}>Sistem raporunu CSV formatında indirebilirsiniz.</p>
          <button onClick={handleExport} style={{ width: '100%', padding: '12px', backgroundColor: '#eab308', border: 'none', color: '#000', fontWeight: 'bold', borderRadius: '4px', cursor: 'pointer' }}>
            CSV Olarak İndir
          </button>
        </div>
      </div>
    </div>
  );
};