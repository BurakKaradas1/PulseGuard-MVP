import React, { useState } from 'react';
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';
import type { HostDetail, ThresholdConfig } from '../types';

interface HostDetailPanelProps {
  selectedHost: HostDetail | null;
  onClose: () => void;
  metricsHistory: { time: string; cpu: number; ram: number; disk: number }[];
  editThresholds: ThresholdConfig;
  setEditThresholds: (t: ThresholdConfig) => void;
  onSaveThresholds: () => void;
  saveStatus: string;
}

export const HostDetailPanel: React.FC<HostDetailPanelProps> = ({
  selectedHost, onClose, metricsHistory, editThresholds, setEditThresholds, onSaveThresholds, saveStatus
}) => {
  const [isClosing, setIsClosing] = useState(false);
  const [activeTab, setActiveTab] = useState<'cpu' | 'ram' | 'disk'>('cpu');

  const handleCloseClick = () => {
    setIsClosing(true);
    setTimeout(() => {
      setIsClosing(false);
      onClose();
    }, 300);
  };

  if (!selectedHost) return null;

  const currentMetrics = metricsHistory.length > 0 ? metricsHistory[metricsHistory.length - 1] : { cpu: 0, ram: 0, disk: 0 };
  
  const alarms = [];
  if (currentMetrics.cpu > editThresholds.max_cpu_usage) alarms.push({ name: 'CPU', current: currentMetrics.cpu, limit: editThresholds.max_cpu_usage });
  if (currentMetrics.ram > editThresholds.max_ram_usage) alarms.push({ name: 'RAM', current: currentMetrics.ram, limit: editThresholds.max_ram_usage });
  if (currentMetrics.disk > editThresholds.max_disk_usage) alarms.push({ name: 'DISK', current: currentMetrics.disk, limit: editThresholds.max_disk_usage });

  const hasAnyAlarm = alarms.length > 0;
  const activeValue = currentMetrics[activeTab];
  const activeThreshold = activeTab === 'cpu' ? editThresholds.max_cpu_usage : activeTab === 'ram' ? editThresholds.max_ram_usage : editThresholds.max_disk_usage;
  const isActiveTabAlarm = activeValue > activeThreshold;

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
          background-color: #111827; border: 1px solid rgba(202, 138, 4, 0.5); color: #facc15;
          padding: 8px 12px; border-radius: 6px; font-family: monospace; font-size: 14px;
          font-weight: bold; width: 80px; outline: none; text-align: center;
          transition: all 0.2s ease; box-shadow: inset 0 2px 4px rgba(0,0,0,0.2);
        }
        .modern-input:hover { border-color: #facc15; background-color: rgba(202, 138, 4, 0.15); }
        .modern-input:focus { border-color: #facc15; box-shadow: 0 0 0 2px rgba(234, 179, 8, 0.3); background-color: #111827; }
        .modern-input::-webkit-outer-spin-button, .modern-input::-webkit-inner-spin-button { -webkit-appearance: none; margin: 0; }
        .modern-input[type=number] { -moz-appearance: textfield; }
        
        .tab-btn {
          flex: 1; padding: 6px 0; background-color: #111827; border: 1px solid rgba(202, 138, 4, 0.5);
          color: #9ca3af; font-family: monospace; font-size: 12px; font-weight: bold; cursor: pointer; transition: all 0.2s;
        }
        .tab-btn.active { background-color: #eab308; color: #000; border-color: #eab308; }
        .tab-btn:hover:not(.active) { background-color: rgba(202, 138, 4, 0.15); color: #eab308; }
        .tab-btn:first-child { border-radius: 4px 0 0 4px; }
        .tab-btn:nth-child(2) { border-left: none; border-right: none; }
        .tab-btn:last-child { border-radius: 0 4px 4px 0; }
        
        @keyframes flash-alarm { 0% { opacity: 1; } 50% { opacity: 0.6; } 100% { opacity: 1; } }
      `}</style>

      {hasAnyAlarm && (
        <div style={{ 
          backgroundColor: 'rgba(239, 68, 68, 0.2)', borderBottom: '1px solid #ef4444', 
          padding: '12px 24px', display: 'flex', alignItems: 'center', gap: '12px',
          animation: 'flash-alarm 2s infinite'
        }}>
          <span style={{ fontSize: '20px' }}>🚨</span>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
            {alarms.map(alarm => (
              <div key={alarm.name} style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                <div style={{ color: '#fca5a5', fontWeight: 'bold', fontSize: '13px', textTransform: 'uppercase' }}>{alarm.name} LIMIT EXCEEDED</div>
                <div style={{ color: '#ef4444', fontSize: '12px' }}>(Current: {alarm.current}% | Limit: {alarm.limit}%)</div>
              </div>
            ))}
          </div>
        </div>
      )}

      <div style={{ padding: '24px', display: 'flex', flexDirection: 'column', gap: '24px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', borderBottom: '1px solid rgba(202, 138, 4, 0.3)', paddingBottom: '16px' }}>
          <div>
            <div style={{ fontSize: '11px', color: '#6b7280', letterSpacing: '1px' }}>HOST DETAIL</div>
            <h2 style={{ fontSize: '22px', fontWeight: 'bold', color: '#facc15', margin: '4px 0 0 0' }}>{selectedHost.hostname}</h2>
            <div style={{ fontSize: '13px', color: '#9ca3af', marginTop: '2px' }}>{selectedHost.ip_address}</div>
          </div>
          <button onClick={handleCloseClick} style={{ background: 'none', border: 'none', color: '#9ca3af', fontSize: '22px', cursor: 'pointer', transition: 'color 0.2s' }} onMouseEnter={(e) => e.currentTarget.style.color = '#facc15'} onMouseLeave={(e) => e.currentTarget.style.color = '#9ca3af'}>×</button>
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
              <div style={{ fontSize: '13px', color: hasAnyAlarm ? '#ef4444' : '#4ade80', fontWeight: 'bold', marginTop: '4px', textTransform: 'uppercase' }}>
                {hasAnyAlarm ? 'WARNING' : selectedHost.status}
              </div>
            </div>
          </div>
        </div>

        <div>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '8px' }}>
            <h3 style={{ fontSize: '13px', color: '#d1d5db', textTransform: 'uppercase', margin: 0 }}>Metrics Trend (Live)</h3>
          </div>
          
          <div style={{ display: 'flex', width: '100%', marginBottom: '12px' }}>
            <button className={`tab-btn ${activeTab === 'cpu' ? 'active' : ''}`} onClick={() => setActiveTab('cpu')}>CPU</button>
            <button className={`tab-btn ${activeTab === 'ram' ? 'active' : ''}`} onClick={() => setActiveTab('ram')}>RAM</button>
            <button className={`tab-btn ${activeTab === 'disk' ? 'active' : ''}`} onClick={() => setActiveTab('disk')}>DISK</button>
          </div>

          <div style={{ background: '#000', padding: '12px 8px 4px 0', borderRadius: '8px', border: '1px solid rgba(202, 138, 4, 0.3)', height: '150px' }}>
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={metricsHistory}>
                <defs>
                  <linearGradient id="metricColor" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor={isActiveTabAlarm ? "#ef4444" : "#eab308"} stopOpacity={0.8}/>
                    <stop offset="95%" stopColor={isActiveTabAlarm ? "#ef4444" : "#eab308"} stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <XAxis dataKey="time" stroke="#6b7280" fontSize={10} tickLine={false} />
                <YAxis stroke="#6b7280" fontSize={10} domain={[0, 100]} tickLine={false} />
                <Tooltip contentStyle={{ backgroundColor: '#111827', borderColor: isActiveTabAlarm ? '#ef4444' : '#eab308', fontSize: '12px', color: '#fff' }} />
                <Area type="monotone" dataKey={activeTab} stroke={isActiveTabAlarm ? "#ef4444" : "#eab308"} fillOpacity={1} fill="url(#metricColor)" />
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
            {hasAnyAlarm ? (
              <div style={{ background: 'rgba(239, 68, 68, 0.1)', padding: '10px 14px', borderRadius: '6px', borderLeft: '3px solid #ef4444' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '11px', color: '#6b7280', marginBottom: '2px' }}>
                  <span style={{ color: '#ef4444', fontWeight: 'bold' }}>CRITICAL</span>
                  <span>Just Now</span>
                </div>
                <div style={{ fontSize: '12px', color: '#d1d5db' }}>System thresholds exceeded. Check banner for details.</div>
              </div>
            ) : (
              <div style={{ background: '#000', padding: '10px 14px', borderRadius: '6px', borderLeft: '3px solid #4ade80' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '11px', color: '#6b7280', marginBottom: '2px' }}>
                  <span style={{ color: '#4ade80', fontWeight: 'bold' }}>STABLE</span>
                  <span>Last Control</span>
                </div>
                <div style={{ fontSize: '12px', color: '#d1d5db' }}>Agent system status stable, all metrics within normal ranges.</div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};