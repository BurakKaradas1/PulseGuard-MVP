import React, { useState, useEffect } from 'react';
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

  const [isVisible, setIsVisible] = useState(false);
  const [isClosing, setIsClosing] = useState(false);

  useEffect(() => {
    if (isOpen) {
      setIsClosing(false);
      const timer = setTimeout(() => setIsVisible(true), 10);
      return () => clearTimeout(timer);
    } else {
      setIsVisible(false);
    }
  }, [isOpen]);

  const handleCloseClick = () => {
    setIsClosing(true);
    setIsVisible(false);
    setTimeout(() => {
      setIsClosing(false); 
      onClose();
    }, 300);
  };

  if (!isOpen && !isClosing) return null;

  const handleApplyGlobalThresholds = async () => {
    if (hosts.length === 0) {
      setStatusMessage("No hosts found in the fleet.");
      return;
    }

    setIsApplying(true);
    setStatusMessage("Applying to all hosts...");

    try {
      const promises = hosts.map(host => 
        fetch(`${apiBaseUrl}/api/v1/dashboard/hosts/threshold?id=${host.id}`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
          body: JSON.stringify(globalThresholds)
        })
      );

      await Promise.all(promises);
      
      setStatusMessage("Successfully applied to the entire fleet!");
      setTimeout(() => setStatusMessage(""), 3000);
    } catch (error) {
      setStatusMessage("An error occurred during application.");
    } finally {
      setIsApplying(false);
    }
  };

  return (
    <div style={{ 
      position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, 
      backgroundColor: 'rgba(0, 0, 0, 0.8)', zIndex: 100, display: 'flex', justifyContent: 'center', alignItems: 'center', backdropFilter: 'blur(4px)',
      opacity: isVisible ? 1 : 0,
      transition: 'opacity 0.3s ease-in-out'
    }}>
      
      <style>{`
        .modern-select {
          appearance: none;
          -webkit-appearance: none;
          background-color: #111827;
          color: #facc15;
          border: 1px solid rgba(202, 138, 4, 0.5);
          padding: 8px 36px 8px 16px;
          border-radius: 6px;
          outline: none;
          cursor: pointer;
          font-size: 13px;
          font-family: monospace;
          font-weight: bold;
          background-image: url("data:image/svg+xml;charset=UTF-8,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%23facc15' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'%3E%3C/polyline%3E%3C/svg%3E");
          background-repeat: no-repeat;
          background-position: right 12px center;
          background-size: 16px;
          transition: all 0.2s ease;
          box-shadow: 0 2px 4px rgba(0,0,0,0.2);
        }
        .modern-select:hover {
          border-color: #facc15;
          background-color: rgba(202, 138, 4, 0.15);
        }
        .modern-select:focus {
          border-color: #facc15;
          box-shadow: 0 0 0 2px rgba(234, 179, 8, 0.3);
        }
        
        /* ÇÖZÜM BURADA: Açılır menü seçeneklerinin arka planını koyu yaptık */
        .modern-select option {
          background-color: #111827;
          color: #facc15;
          font-weight: bold;
        }
        
        .modern-input {
          background-color: #111827;
          border: 1px solid rgba(202, 138, 4, 0.5);
          color: #facc15;
          padding: 8px 12px;
          border-radius: 6px;
          font-family: monospace;
          font-size: 14px;
          font-weight: bold;
          width: 80px;
          outline: none;
          text-align: center;
          transition: all 0.2s ease;
          box-shadow: inset 0 2px 4px rgba(0,0,0,0.2);
        }
        .modern-input:hover {
          border-color: #facc15;
          background-color: rgba(202, 138, 4, 0.15);
        }
        .modern-input:focus {
          border-color: #facc15;
          box-shadow: 0 0 0 2px rgba(234, 179, 8, 0.3);
          background-color: #111827;
        }
        .modern-input::-webkit-outer-spin-button,
        .modern-input::-webkit-inner-spin-button {
          -webkit-appearance: none;
          margin: 0;
        }
        .modern-input[type=number] {
          -moz-appearance: textfield;
        }
      `}</style>

      <div style={{ 
        backgroundColor: '#111827', border: '1px solid rgba(202, 138, 4, 0.5)', borderRadius: '8px', width: '500px', boxShadow: '0 0 40px rgba(202, 138, 4, 0.1)', overflow: 'hidden',
        transform: isVisible ? 'scale(1) translateY(0)' : 'scale(0.95) translateY(20px)',
        opacity: isVisible ? 1 : 0,
        transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)'
      }}>
        
        <div style={{ padding: '20px 24px', borderBottom: '1px solid rgba(202, 138, 4, 0.3)', display: 'flex', justifyContent: 'space-between', alignItems: 'center', backgroundColor: '#030712' }}>
          <h2 style={{ margin: 0, fontSize: '18px', color: '#facc15' }}>⚙️ System Settings</h2>
          <button 
            onClick={handleCloseClick} 
            style={{ background: 'none', border: 'none', color: '#9ca3af', fontSize: '20px', cursor: 'pointer', transition: 'color 0.2s' }}
            onMouseEnter={(e) => e.currentTarget.style.color = '#facc15'}
            onMouseLeave={(e) => e.currentTarget.style.color = '#9ca3af'}
          >
            ×
          </button>
        </div>

        <div style={{ padding: '24px', display: 'flex', flexDirection: 'column', gap: '24px' }}>
          
          <div>
            <h3 style={{ fontSize: '13px', color: '#d1d5db', textTransform: 'uppercase', marginBottom: '12px' }}>Interface Refresh Rate</h3>
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', background: 'rgba(0,0,0,0.3)', padding: '12px', borderRadius: '6px', border: '1px solid rgba(202, 138, 4, 0.2)' }}>
              <span style={{ fontSize: '13px', color: '#9ca3af' }}>Data fetch interval:</span>
              <select 
                className="modern-select"
                value={refreshRate} 
                onChange={(e) => setRefreshRate(Number(e.target.value))}
              >
                <option value={3000}>3 Seconds (Aggressive)</option>
                <option value={5000}>5 Seconds (Standard)</option>
                <option value={15000}>15 Seconds</option>
                <option value={30000}>30 Seconds</option>
              </select>
            </div>
          </div>

          <div>
            <h3 style={{ fontSize: '13px', color: '#d1d5db', textTransform: 'uppercase', marginBottom: '12px' }}>Global Thresholds for Fleet</h3>
            <div style={{ background: 'rgba(0,0,0,0.3)', padding: '16px', borderRadius: '6px', border: '1px solid rgba(202, 138, 4, 0.2)', display: 'flex', flexDirection: 'column', gap: '12px' }}>
              
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span style={{ fontSize: '13px', color: '#9ca3af' }}>Max CPU (%):</span>
                <input type="number" className="modern-input" value={globalThresholds.max_cpu_usage} onChange={(e) => setGlobalThresholds({...globalThresholds, max_cpu_usage: Number(e.target.value)})} />
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span style={{ fontSize: '13px', color: '#9ca3af' }}>Max RAM (%):</span>
                <input type="number" className="modern-input" value={globalThresholds.max_ram_usage} onChange={(e) => setGlobalThresholds({...globalThresholds, max_ram_usage: Number(e.target.value)})} />
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span style={{ fontSize: '13px', color: '#9ca3af' }}>Max Disk (%):</span>
                <input type="number" className="modern-input" value={globalThresholds.max_disk_usage} onChange={(e) => setGlobalThresholds({...globalThresholds, max_disk_usage: Number(e.target.value)})} />
              </div>

              <button 
                onClick={handleApplyGlobalThresholds}
                disabled={isApplying}
                style={{ marginTop: '8px', padding: '10px', backgroundColor: isApplying ? '#9ca3af' : '#eab308', border: 'none', color: '#000', fontWeight: 'bold', borderRadius: '4px', cursor: isApplying ? 'not-allowed' : 'pointer' }}
              >
                {isApplying ? 'Applying...' : 'Apply to All Fleet'}
              </button>
              {statusMessage && <div style={{ fontSize: '12px', color: '#4ade80', textAlign: 'center', marginTop: '4px' }}>{statusMessage}</div>}

            </div>
          </div>

        </div>

        <div style={{ padding: '16px 24px', borderTop: '1px solid rgba(202, 138, 4, 0.3)', display: 'flex', justifyContent: 'flex-end', backgroundColor: '#030712' }}>
          <button 
            onClick={handleCloseClick} 
            style={{ padding: '8px 24px', backgroundColor: 'transparent', border: '1px solid #eab308', color: '#eab308', borderRadius: '4px', cursor: 'pointer', fontWeight: 'bold', transition: 'all 0.2s' }}
            onMouseEnter={(e) => { e.currentTarget.style.backgroundColor = '#eab308'; e.currentTarget.style.color = '#000'; }}
            onMouseLeave={(e) => { e.currentTarget.style.backgroundColor = 'transparent'; e.currentTarget.style.color = '#eab308'; }}
          >
            Close
          </button>
        </div>

      </div>
    </div>
  );
};