import React, { useState } from 'react';
import type { HostStatus, ThresholdConfig } from '../types';

interface SettingsModalProps {
  isOpen: boolean;
  onClose: () => void;
  hosts: HostStatus[];
  refreshRate: number;
  setRefreshRate: (rate: number) => void;
  apiBaseUrl: string;
}

export const SettingsModal: React.FC<SettingsModalProps> = ({ isOpen, onClose, hosts, refreshRate, setRefreshRate, apiBaseUrl }) => {
  const [globalThresholds, setGlobalThresholds] = useState<ThresholdConfig>({ max_cpu_usage: 90, max_ram_usage: 90, max_disk_usage: 90, error_alert_limit: 5 });
  const [isApplying, setIsApplying] = useState(false);
  const [statusMessage, setStatusMessage] = useState<string>("");

  if (!isOpen) return null;

  const handleApplyGlobalThresholds = async () => {
    if (hosts.length === 0) {
      setStatusMessage("Filoda host bulunmuyor.");
      return;
    }

    setIsApplying(true);
    setStatusMessage("Tüm hostlara uygulanıyor...");

    try {
      // Tüm hostlara sırasıyla eşik güncelleme isteği atıyoruz (Toplu İşlem)
      const promises = hosts.map(host => 
        fetch(`${apiBaseUrl}/api/v1/dashboard/hosts/threshold?id=${host.id}`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
          body: JSON.stringify(globalThresholds)
        })
      );

      await Promise.all(promises);
      
      setStatusMessage("Tüm filoya başarıyla uygulandı!");
      setTimeout(() => setStatusMessage(""), 3000);
    } catch (error) {
      setStatusMessage("Uygulama sırasında bir hata oluştu.");
    } finally {
      setIsApplying(false);
    }
  };

  return (
    <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, backgroundColor: 'rgba(0, 0, 0, 0.8)', zIndex: 100, display: 'flex', justifyContent: 'center', alignItems: 'center', backdropFilter: 'blur(4px)' }}>
      <div style={{ backgroundColor: '#111827', border: '1px solid rgba(202, 138, 4, 0.5)', borderRadius: '8px', width: '500px', boxShadow: '0 0 40px rgba(202, 138, 4, 0.1)', overflow: 'hidden' }}>
        
        <div style={{ padding: '20px 24px', borderBottom: '1px solid rgba(202, 138, 4, 0.3)', display: 'flex', justifyContent: 'space-between', alignItems: 'center', backgroundColor: '#030712' }}>
          <h2 style={{ margin: 0, fontSize: '18px', color: '#facc15' }}>⚙️ Sistem Ayarları</h2>
          <button onClick={onClose} style={{ background: 'none', border: 'none', color: '#9ca3af', fontSize: '20px', cursor: 'pointer' }}>×</button>
        </div>

        <div style={{ padding: '24px', display: 'flex', flexDirection: 'column', gap: '24px' }}>
          
          {/* YENİLEME SIKLIĞI AYARI */}
          <div>
            <h3 style={{ fontSize: '13px', color: '#d1d5db', textTransform: 'uppercase', marginBottom: '12px' }}>Arayüz Yenileme Sıklığı</h3>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', background: 'rgba(0,0,0,0.3)', padding: '12px', borderRadius: '6px', border: '1px solid rgba(202, 138, 4, 0.2)' }}>
              <span style={{ fontSize: '13px', color: '#9ca3af' }}>Veri çekme aralığı:</span>
              <select 
                value={refreshRate} 
                onChange={(e) => setRefreshRate(Number(e.target.value))}
                style={{ backgroundColor: '#000', color: '#fff', border: '1px solid rgba(202, 138, 4, 0.5)', padding: '6px 12px', borderRadius: '4px', outline: 'none', cursor: 'pointer' }}
              >
                <option value={3000}>3 Saniye (Agresif)</option>
                <option value={5000}>5 Saniye (Standart)</option>
                <option value={15000}>15 Saniye</option>
                <option value={30000}>30 Saniye</option>
              </select>
            </div>
          </div>

          {/* GLOBAL EŞİK AYARLARI */}
          <div>
            <h3 style={{ fontSize: '13px', color: '#d1d5db', textTransform: 'uppercase', marginBottom: '12px' }}>Tüm Filo İçin Global Eşikler</h3>
            <div style={{ background: 'rgba(0,0,0,0.3)', padding: '16px', borderRadius: '6px', border: '1px solid rgba(202, 138, 4, 0.2)', display: 'flex', flexDirection: 'column', gap: '12px' }}>
              
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span style={{ fontSize: '13px', color: '#9ca3af' }}>Max CPU (%):</span>
                <input type="number" className="threshold-input" value={globalThresholds.max_cpu_usage} onChange={(e) => setGlobalThresholds({...globalThresholds, max_cpu_usage: Number(e.target.value)})} />
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span style={{ fontSize: '13px', color: '#9ca3af' }}>Max RAM (%):</span>
                <input type="number" className="threshold-input" value={globalThresholds.max_ram_usage} onChange={(e) => setGlobalThresholds({...globalThresholds, max_ram_usage: Number(e.target.value)})} />
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span style={{ fontSize: '13px', color: '#9ca3af' }}>Max Disk (%):</span>
                <input type="number" className="threshold-input" value={globalThresholds.max_disk_usage} onChange={(e) => setGlobalThresholds({...globalThresholds, max_disk_usage: Number(e.target.value)})} />
              </div>

              <button 
                onClick={handleApplyGlobalThresholds}
                disabled={isApplying}
                style={{ marginTop: '8px', padding: '10px', backgroundColor: isApplying ? '#9ca3af' : '#eab308', border: 'none', color: '#000', fontWeight: 'bold', borderRadius: '4px', cursor: isApplying ? 'not-allowed' : 'pointer' }}
              >
                {isApplying ? 'Uygulanıyor...' : 'Tüm Filoya Uygula'}
              </button>
              {statusMessage && <div style={{ fontSize: '12px', color: '#4ade80', textAlign: 'center', marginTop: '4px' }}>{statusMessage}</div>}

            </div>
          </div>

        </div>

        <div style={{ padding: '16px 24px', borderTop: '1px solid rgba(202, 138, 4, 0.3)', display: 'flex', justifyContent: 'flex-end', backgroundColor: '#030712' }}>
          <button onClick={onClose} style={{ padding: '8px 24px', backgroundColor: 'transparent', border: '1px solid #eab308', color: '#eab308', borderRadius: '4px', cursor: 'pointer', fontWeight: 'bold' }}>Kapat</button>
        </div>

      </div>
    </div>
  );
};