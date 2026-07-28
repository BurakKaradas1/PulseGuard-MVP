import { useState, useEffect } from 'react';
import type { HostStatus, HostDetail, ThresholdConfig } from './types';
import { Sidebar } from './components/Sidebar';
import { Header } from './components/Header';
import { SummaryCards } from './components/SummaryCards';
import { HostTable } from './components/HostTable';
import { HostDetailPanel } from './components/HostDetailPanel';
import { SettingsModal } from './components/SettingsModal';
import { ExportModal } from './components/ExportModal';

export function App() {
  const [hosts, setHosts] = useState<HostStatus[]>([]);
  const [selectedHost, setSelectedHost] = useState<HostDetail | null>(null);
  const [apiStatus, setApiStatus] = useState<"connected" | "disconnected">("disconnected");
  
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);
  const [isExportOpen, setIsExportOpen] = useState(false);

  const [editThresholds, setEditThresholds] = useState<ThresholdConfig>({ max_cpu_usage: 90, max_ram_usage: 90, max_disk_usage: 90, error_alert_limit: 5 });
  const [saveStatus, setSaveStatus] = useState<string>("");
  const [cpuHistory, setCpuHistory] = useState<{ time: string; cpu: number }[]>([]);
  const [refreshRate, setRefreshRate] = useState<number>(5000);

  const API_BASE_URL = ''; 

  useEffect(() => {
    let isMounted = true;
    let timeoutId: ReturnType<typeof setTimeout>;

    const fetchHostsSafely = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/api/v1/dashboard/hosts`);
        if (!response.ok) throw new Error("Ağ hatası");
        const data = await response.json();
        
        if (isMounted) {
          setHosts(data || []);
          setApiStatus("connected");

          if (selectedHost) {
            const currentUpdatedHost = data.find((h: HostStatus) => h.id === selectedHost.id);
            if (currentUpdatedHost) {
              const nowTime = new Date().toLocaleTimeString('tr-TR', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
              setCpuHistory(prev => [
                ...prev.slice(-6),
                { time: nowTime, cpu: currentUpdatedHost.cpu_usage || 0 }
              ]);
            }
          }
        }
      } catch (error) {
        if (isMounted) {
          setApiStatus("disconnected");
          setHosts([]);
        }
      } finally {
        if (isMounted) {
          timeoutId = setTimeout(fetchHostsSafely, refreshRate);
        }
      }
    };

    fetchHostsSafely();

    return () => {
      isMounted = false;
      clearTimeout(timeoutId);
    };
  }, [selectedHost?.id, refreshRate]);

  const handleHostClick = async (hostId: string) => {
    const baseHost = hosts.find(h => h.id === hostId);
    if (!baseHost) return;

    const initialDetail: HostDetail = {
      ...baseHost,
      ip_address: baseHost.ip_address || "192.168.1.X",
      os: "Yükleniyor...",
      threshold: { max_cpu_usage: 90, max_ram_usage: 90, max_disk_usage: 90, error_alert_limit: 5 }
    };
    setSelectedHost(initialDetail);
    setEditThresholds(initialDetail.threshold);
    setSaveStatus("");

    const nowTime = new Date().toLocaleTimeString('tr-TR', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    setCpuHistory([{ time: nowTime, cpu: baseHost.cpu_usage || 0 }]);

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/dashboard/hosts/detail?id=${hostId}`);
      if (response.ok) {
        const detailData: HostDetail = await response.json();
        setSelectedHost(detailData);
        setEditThresholds(detailData.threshold);
      }
    } catch (error) {
      setSelectedHost({
        ...baseHost,
        ip_address: baseHost.ip_address || "192.168.1.50",
        os: "Ubuntu 22.04 LTS",
        threshold: { max_cpu_usage: 90, max_ram_usage: 90, max_disk_usage: 90, error_alert_limit: 5 }
      });
    }
  };

  const handleSaveThresholds = async () => {
    if (!selectedHost) return;
    setSaveStatus("Kaydediliyor...");
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/dashboard/hosts/threshold?id=${selectedHost.id}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
        body: JSON.stringify(editThresholds)
      });
      
      if (response.ok) {
        setSaveStatus("Başarıyla kaydedildi!");
        setSelectedHost({...selectedHost, threshold: editThresholds}); 
        setTimeout(() => setSaveStatus(""), 3000);
      } else {
        const errorText = await response.text();
        setSaveStatus(`Kayıt başarısız: ${errorText}`);
      }
    } catch (err) {
      setSaveStatus("Bağlantı hatası.");
    }
  };

  return (
    <div style={{ display: 'flex', height: '100vh', backgroundColor: '#030712', color: '#facc15', fontFamily: 'monospace', position: 'relative', overflow: 'hidden' }}>
      
      <style>{`
        .host-row { transition: background-color 0.2s ease; cursor: pointer; position: relative; }
        .host-row:after { content: ''; position: absolute; left: 0; top: 0; bottom: 0; width: 0; background-color: #eab308; transition: width 0.2s ease; opacity: 0; }
        .host-row:hover:after { width: 6px; opacity: 1; }
        .host-row:hover { background-color: rgba(202, 138, 4, 0.25) !important; }
        .selected-row { background-color: rgba(202, 138, 4, 0.35) !important; }
        .selected-row:after { width: 6px; opacity: 1; }
        .menu-item { transition: all 0.2s ease; cursor: pointer; }
        .menu-item:hover { background-color: rgba(202, 138, 4, 0.2) !important; }
        @keyframes radar-pulse {
          0% { box-shadow: 0 0 0 0 rgba(239, 68, 68, 0.7); }
          70% { box-shadow: 0 0 0 6px rgba(239, 68, 68, 0); }
          100% { box-shadow: 0 0 0 0 rgba(239, 68, 68, 0); }
        }
        .pulse-dot { width: 10px; height: 10px; background-color: #ef4444; border-radius: 50%; animation: radar-pulse 2s infinite; }
        ::-webkit-scrollbar { width: 8px; }
        ::-webkit-scrollbar-track { background: rgba(0, 0, 0, 0.3); }
        ::-webkit-scrollbar-thumb { background: rgba(202, 138, 4, 0.3); border-radius: 4px; }
        .action-btn { padding: 6px 16px; background-color: transparent; border: 1px solid #eab308; color: #eab308; border-radius: 4px; cursor: pointer; font-weight: bold; font-family: monospace; transition: all 0.2s; }
        .action-btn:hover { background-color: rgba(234, 179, 8, 0.2); box-shadow: 0 0 10px rgba(234, 179, 8, 0.3); }
        .threshold-input { background-color: #000; border: 1px solid rgba(202, 138, 4, 0.5); color: #fff; padding: 6px 10px; border-radius: 4px; font-family: monospace; font-size: 14px; width: 70px; outline: none; }
        .threshold-input:focus { border-color: #eab308; box-shadow: 0 0 8px rgba(234, 179, 8, 0.3); }
      `}</style>

      <Sidebar onOpenSettings={() => setIsSettingsOpen(true)} />

      <main style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden', transition: 'margin-right 0.3s ease', marginRight: selectedHost ? '450px' : '0' }}>
        <Header apiStatus={apiStatus} apiBaseUrl={API_BASE_URL} />

        <div style={{ padding: '32px', flex: 1, overflowY: 'auto' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px', borderBottom: '1px solid rgba(202, 138, 4, 0.5)', paddingBottom: '8px' }}>
            <h2 style={{ fontSize: '20px', margin: 0, color: '#d1d5db', textShadow: '0 0 8px rgba(209, 213, 219, 0.3)' }}>Filo Metrikleri (Özet)</h2>
            <button className="action-btn" onClick={() => setIsExportOpen(true)}>📥 RAPOR AL</button>
          </div>
          
          <SummaryCards hosts={hosts} />

          <h2 style={{ fontSize: '20px', marginBottom: '16px', borderBottom: '1px solid rgba(202, 138, 4, 0.5)', paddingBottom: '8px', color: '#d1d5db', textShadow: '0 0 8px rgba(209, 213, 219, 0.3)' }}>Host Listesi</h2>
          
          <HostTable hosts={hosts} selectedHostId={selectedHost?.id} onHostClick={handleHostClick} />
        </div>
      </main>

      <HostDetailPanel 
        selectedHost={selectedHost} 
        onClose={() => setSelectedHost(null)}
        cpuHistory={cpuHistory}
        editThresholds={editThresholds}
        setEditThresholds={setEditThresholds}
        onSaveThresholds={handleSaveThresholds}
        saveStatus={saveStatus}
      />

      <SettingsModal 
        isOpen={isSettingsOpen} 
        onClose={() => setIsSettingsOpen(false)} 
        hosts={hosts}
        refreshRate={refreshRate}
        setRefreshRate={setRefreshRate}
        apiBaseUrl={API_BASE_URL}
      />
      <ExportModal isOpen={isExportOpen} onClose={() => setIsExportOpen(false)} hosts={hosts} />
    </div>
  );
}

export default App;