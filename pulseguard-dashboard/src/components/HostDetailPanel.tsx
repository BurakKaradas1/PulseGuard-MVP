import React from 'react';
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';
import type { HostDetail, ThresholdConfig } from '../types';

interface HostDetailPanelProps {
  selectedHost: HostDetail | null;
  onClose: () => void;
  cpuHistory: { time: string; cpu: number }[];
  editThresholds: ThresholdConfig;
  setEditThresholds: (t: ThresholdConfig) => void;
  onSaveThresholds: () => void;
  saveStatus: string;
}

export const HostDetailPanel: React.FC<HostDetailPanelProps> = ({
  selectedHost, onClose, cpuHistory, editThresholds, setEditThresholds, onSaveThresholds, saveStatus
}) => {
  const [isClosing, setIsClosing] = React.useState(false);

  const handleCloseClick = () => {
    setIsClosing(true);
    setTimeout(() => {
      setIsClosing(false);
      onClose();
    }, 300);
  };

  if (!selectedHost) return null;

  return (
    <div style={{
      position: 'absolute', top: 0, right: 0, bottom: 0, width: '450px',
      backgroundColor: '#111827', borderLeft: '1px solid rgba(202, 138, 4, 0.5)',
      transform: isClosing ? 'translateX(100%)' : 'translateX(0)',
      transition: 'transform 0.3s cubic-bezier(0.4, 0, 0.2, 1)',
      boxShadow: '-10px 0 30px rgba(0,0,0,0.5)',
      display: 'flex', flexDirection: 'column', zIndex: 50, overflowY: 'auto'
    }}>
      
      <style>{`
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

      <div style={{ padding: '24px', display: 'flex', flexDirection: 'column', gap: '24px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', borderBottom: '1px solid rgba(202, 138, 4, 0.3)', paddingBottom: '16px' }}>
          <div>
            <div style={{ fontSize: '11px', color: '#6b7280', letterSpacing: '1px' }}>HOST DETAIL</div>
            <h2 style={{ fontSize: '22px', fontWeight: 'bold', color: '#facc15', margin: '4px 0 0 0' }}>{selectedHost.hostname}</h2>
            <div style={{ fontSize: '13px', color: '#9ca3af', marginTop: '2px' }}>{selectedHost.ip_address}</div>
          </div>
          <button 
            onClick={handleCloseClick} 
            style={{ background: 'none', border: 'none', color: '#9ca3af', fontSize: '22px', cursor: 'pointer', transition: 'color 0.2s' }}
            onMouseEnter={(e) => e.currentTarget.style.color = '#facc15'}
            onMouseLeave={(e) => e.currentTarget.style.color = '#9ca3af'}
          >
            ×
          </button>
        </div>

        <div>
          <h3 style={{ fontSize: '13px', color: '#d1d5db', textTransform: 'uppercase', marginBottom: '12px' }}>System Information</h3>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
            <div style={{ background: '#000', padding: '12px', borderRadius: '6px', border: '1px solid rgba(202, 138, 4, 0.2)' }}>
              <div style={{ fontSize: '10px', color: '#6b7280' }}>OPERATING SYSTEM</div>
              <div style={{ fontSize: '13px', color: '#fff', fontWeight: 'bold', marginTop: '4px' }}>{selectedHost.os}</div>
            </div>
            <div style={{ background: '#000', padding: '12px', borderRadius: '6px', border: '1px solid rgba(202, 138, 4, 0.2)' }}>
              <div style={{ fontSize: '10px', color: '#6b7280' }}>STATUS</div>
              <div style={{ fontSize: '13px', color: '#4ade80', fontWeight: 'bold', marginTop: '4px', textTransform: 'uppercase' }}>{selectedHost.status}</div>
            </div>
          </div>
        </div>

        <div>
          <h3 style={{ fontSize: '13px', color: '#d1d5db', textTransform: 'uppercase', marginBottom: '8px' }}>CPU Usage Trend (Live)</h3>
          <div style={{ background: '#000', padding: '12px 8px 4px 0', borderRadius: '8px', border: '1px solid rgba(202, 138, 4, 0.3)', height: '150px' }}>
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={cpuHistory}>
                <defs>
                  <linearGradient id="cpuColor" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#eab308" stopOpacity={0.8}/>
                    <stop offset="95%" stopColor="#eab308" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <XAxis dataKey="time" stroke="#6b7280" fontSize={10} tickLine={false} />
                <YAxis stroke="#6b7280" fontSize={10} domain={[0, 100]} tickLine={false} />
                <Tooltip contentStyle={{ backgroundColor: '#111827', borderColor: '#eab308', fontSize: '12px', color: '#fff' }} />
                <Area type="monotone" dataKey="cpu" stroke="#eab308" fillOpacity={1} fill="url(#cpuColor)" />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        <div style={{ background: 'rgba(0,0,0,0.4)', padding: '16px', borderRadius: '8px', border: '1px solid rgba(202, 138, 4, 0.3)' }}>
          <h3 style={{ fontSize: '13px', color: '#facc15', textTransform: 'uppercase', margin: '0 0 12px 0' }}>Alarm Thresholds</h3>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span style={{ fontSize: '13px', color: '#9ca3af' }}>Max CPU (%):</span>
              <input type="number" className="modern-input" value={editThresholds.max_cpu_usage} onChange={(e) => setEditThresholds({...editThresholds, max_cpu_usage: Number(e.target.value)})} />
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span style={{ fontSize: '13px', color: '#9ca3af' }}>Max RAM (%):</span>
              <input type="number" className="modern-input" value={editThresholds.max_ram_usage} onChange={(e) => setEditThresholds({...editThresholds, max_ram_usage: Number(e.target.value)})} />
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span style={{ fontSize: '13px', color: '#9ca3af' }}>Max Disk (%):</span>
              <input type="number" className="modern-input" value={editThresholds.max_disk_usage} onChange={(e) => setEditThresholds({...editThresholds, max_disk_usage: Number(e.target.value)})} />
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span style={{ fontSize: '13px', color: '#9ca3af' }}>Error Limit:</span>
              <input type="number" className="modern-input" value={editThresholds.error_alert_limit} onChange={(e) => setEditThresholds({...editThresholds, error_alert_limit: Number(e.target.value)})} />
            </div>
            <button onClick={onSaveThresholds} style={{ marginTop: '8px', padding: '8px', backgroundColor: '#eab308', border: 'none', color: '#000', fontWeight: 'bold', borderRadius: '4px', cursor: 'pointer' }}>Update Thresholds</button>
            {saveStatus && <div style={{ fontSize: '12px', color: '#4ade80', textAlign: 'center' }}>{saveStatus}</div>}
          </div>
        </div>

        <div>
          <h3 style={{ fontSize: '13px', color: '#d1d5db', textTransform: 'uppercase', marginBottom: '12px' }}>Event / Alarm History</h3>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
            <div style={{ background: '#000', padding: '10px 14px', borderRadius: '6px', borderLeft: '3px solid #eab308' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '11px', color: '#6b7280', marginBottom: '2px' }}>
                <span style={{ color: '#eab308', fontWeight: 'bold' }}>WARNING</span>
                <span>Last Control</span>
              </div>
              <div style={{ fontSize: '12px', color: '#d1d5db' }}>Agent system status stable, periodic heartbeat received.</div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};