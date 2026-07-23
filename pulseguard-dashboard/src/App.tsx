import { useState, useEffect } from 'react';
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer } from 'recharts';

interface HostStatus {
  id: string;
  hostname: string;
  ip_address?: string;
  status: string;
  last_seen: string;
  cpu_usage?: number;
  ram_usage?: number;
  disk_usage?: number;
}

interface ThresholdConfig {
  max_cpu_usage: number;
  max_ram_usage: number;
  max_disk_usage: number;
  error_alert_limit: number;
}

interface HostDetail extends HostStatus {
  ip_address: string;
  os: string;
  threshold: ThresholdConfig;
}

export function App() {
  const [hosts, setHosts] = useState<HostStatus[]>([]);
  const [selectedHost, setSelectedHost] = useState<HostDetail | null>(null);
  const [apiStatus, setApiStatus] = useState<"connected" | "disconnected">("disconnected");
  
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);
  const [isExportOpen, setIsExportOpen] = useState(false);

  const [editThresholds, setEditThresholds] = useState<ThresholdConfig>({ max_cpu_usage: 90, max_ram_usage: 90, max_disk_usage: 90, error_alert_limit: 5 });
  const [saveStatus, setSaveStatus] = useState<string>("");

  // Grafik için anlık veri simülasyon state'i (Sistem akışını bozmaz, görsel zenginlik katar)
  const [cpuHistory, setCpuHistory] = useState<{ time: string; cpu: number }[]>([]);

  const API_BASE_URL = ''; 

  useEffect(() => {
    const fetchHosts = async () => {
      try {
        const response = await fetch(`${API_BASE_URL}/api/v1/dashboard/hosts`);
        if (!response.ok) throw new Error("Ağ hatası");
        const data = await response.json();
        setHosts(data || []);
        setApiStatus("connected");

        // Eğer seçili bir host varsa, grafik geçmişini anlık CPU verisiyle besle
        if (selectedHost) {
          const currentUpdatedHost = data.find((h: HostStatus) => h.id === selectedHost.id);
          if (currentUpdatedHost) {
            const nowTime = new Date().toLocaleTimeString('tr-TR', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
            setCpuHistory(prev => [
              ...prev.slice(-6), // Son 7 veriyi tut
              { time: nowTime, cpu: currentUpdatedHost.cpu_usage || 0 }
            ]);
          }
        }
      } catch (error) {
        setApiStatus("disconnected");
        setHosts([]);
      }
    };

    fetchHosts();
    const pollingInterval = setInterval(fetchHosts, 5000);
    return () => clearInterval(pollingInterval);
  }, [selectedHost?.id]);

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

    // Tıklandığı an ilk grafik verisini oluştur
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
        headers: { 
          'Content-Type': 'application/json',
          'Accept': 'application/json' 
        },
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

  const offlineCount = hosts.filter(h => h.status && h.status.toLowerCase() === 'offline').length;
  const avgCpu = hosts.length > 0 ? Math.round(hosts.reduce((acc, h) => acc + (h.cpu_usage || 0), 0) / hosts.length) : 0;
  const avgRam = hosts.length > 0 ? Math.round(hosts.reduce((acc, h) => acc + (h.ram_usage || 0), 0) / hosts.length) : 0;

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

      {/* SOL MENÜ */}
      <aside style={{ width: '260px', flexShrink: 0, borderRight: '1px solid rgba(202, 138, 4, 0.5)', backgroundColor: '#111827', padding: '24px', display: 'flex', flexDirection: 'column', zIndex: 10 }}>
        <h1 style={{ fontSize: '28px', fontWeight: 'bold', color: '#eab308', margin: '0 0 8px 0', textShadow: '0 0 10px rgba(234, 179, 8, 0.5)' }}>PulseGuard</h1>
        <span style={{ fontSize: '12px', color: '#6b7280', marginBottom: '32px', borderBottom: '1px solid rgba(202, 138, 4, 0.3)', paddingBottom: '16px' }}>Core v1.0.0</span>
        <nav style={{ display: 'flex', flexDirection: 'column', gap: '12px', flex: 1 }}>
          <div className="menu-item" style={{ padding: '12px', borderRadius: '4px', backgroundColor: 'rgba(202, 138, 4, 0.15)', color: '#fef08a', border: '1px solid rgba(202, 138, 4, 0.5)', display: 'flex', alignItems: 'center', gap: '12px' }}>
            <div className="pulse-dot"></div>
            <span style={{ fontWeight: 'bold', letterSpacing: '0.5px' }}>Filo Görünümü</span>
          </div>
          <div className="menu-item" onClick={() => setIsSettingsOpen(true)} style={{ padding: '12px', borderRadius: '4px', color: '#9ca3af', border: '1px solid transparent', display: 'flex', alignItems: 'center', gap: '12px' }}>
            <span style={{ fontSize: '16px' }}>⚙️</span>
            <span style={{ fontWeight: 'bold', letterSpacing: '0.5px' }}>Sistem Ayarları</span>
          </div>
        </nav>
      </aside>

      {/* ANA İÇERİK */}
      <main style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden', transition: 'margin-right 0.3s ease', marginRight: selectedHost ? '450px' : '0' }}>
        
        <header style={{ height: '64px', flexShrink: 0, borderBottom: '1px solid rgba(202, 138, 4, 0.5)', display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0 32px', backgroundColor: 'rgba(17, 24, 39, 0.5)' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
            {apiStatus === "connected" ? (
              <><div style={{ width: '12px', height: '12px', backgroundColor: '#3b82f6', borderRadius: '50%', boxShadow: '0 0 8px #3b82f6' }}></div><span style={{ fontSize: '14px', fontWeight: 'bold', color: '#3b82f6', textShadow: '0 0 5px rgba(59, 130, 246, 0.5)' }}>CANLI BAĞLANTI AKTİF</span></>
            ) : (
              <><div style={{ width: '12px', height: '12px', backgroundColor: '#22c55e', borderRadius: '50%', boxShadow: '0 0 8px #22c55e' }}></div><span style={{ fontSize: '14px', fontWeight: 'bold', color: '#22c55e', textShadow: '0 0 5px rgba(34, 197, 94, 0.5)' }}>SİMÜLASYON MODU (MOCK DATA)</span></>
            )}
          </div>
          <div style={{ fontSize: '12px', color: '#6b7280' }}>
            {apiStatus === "connected" ? `REST API: ${API_BASE_URL || 'Local Proxy'}` : "Bekleniyor..."}
          </div>
        </header>

        <div style={{ padding: '32px', flex: 1, overflowY: 'auto' }}>
          
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px', borderBottom: '1px solid rgba(202, 138, 4, 0.5)', paddingBottom: '8px' }}>
            <h2 style={{ fontSize: '20px', margin: 0, color: '#d1d5db', textShadow: '0 0 8px rgba(209, 213, 219, 0.3)' }}>Filo Metrikleri (Özet)</h2>
            <button className="action-btn" onClick={() => setIsExportOpen(true)}>📥 RAPOR AL</button>
          </div>
          
          {/* ÖZET KARTLARI */}
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, minmax(0, 1fr))', gap: '24px', marginBottom: '32px' }}>
             <div style={{ border: '1px solid rgba(202, 138, 4, 0.4)', padding: '24px', borderRadius: '8px', backgroundColor: 'rgba(17, 24, 39, 0.4)', boxShadow: '0 0 15px rgba(202, 138, 4, 0.1)', display: 'flex', flexDirection: 'column', gap: '8px' }}>
               <span style={{ fontSize: '13px', color: '#9ca3af', textTransform: 'uppercase', letterSpacing: '1px' }}>Toplam Host</span>
               <span style={{ fontSize: '48px', fontWeight: 'bold', color: '#fff', lineHeight: '1', textShadow: '0 0 15px rgba(255,255,255,0.3)' }}>{hosts.length}</span>
             </div>
             <div style={{ border: '1px solid rgba(202, 138, 4, 0.4)', padding: '24px', borderRadius: '8px', backgroundColor: 'rgba(17, 24, 39, 0.4)', boxShadow: '0 0 15px rgba(202, 138, 4, 0.1)', display: 'flex', flexDirection: 'column', gap: '8px' }}>
               <span style={{ fontSize: '13px', color: '#9ca3af', textTransform: 'uppercase', letterSpacing: '1px' }}>Ortalama CPU / RAM</span>
               <span style={{ fontSize: '36px', fontWeight: 'bold', color: '#fff', lineHeight: '1', textShadow: '0 0 15px rgba(255,255,255,0.3)' }}>%{avgCpu} / %{avgRam}</span>
             </div>
             <div style={{ border: '1px solid rgba(239, 68, 68, 0.5)', padding: '24px', borderRadius: '8px', backgroundColor: 'rgba(69, 10, 10, 0.2)', boxShadow: '0 0 25px rgba(239, 68, 68, 0.3)', display: 'flex', flexDirection: 'column', gap: '8px' }}>
               <span style={{ fontSize: '13px', color: '#fca5a5', textTransform: 'uppercase', letterSpacing: '1px' }}>Offline Host</span>
               <span style={{ fontSize: '48px', fontWeight: 'bold', color: '#f87171', lineHeight: '1', textShadow: '0 0 15px rgba(248, 113, 113, 0.6)' }}>{offlineCount}</span>
             </div>
          </div>

          <h2 style={{ fontSize: '20px', marginBottom: '16px', borderBottom: '1px solid rgba(202, 138, 4, 0.5)', paddingBottom: '8px', color: '#d1d5db', textShadow: '0 0 8px rgba(209, 213, 219, 0.3)' }}>Host Listesi</h2>
          
          <div style={{ border: '1px solid rgba(202, 138, 4, 0.4)', borderRadius: '8px', backgroundColor: '#000', overflow: 'hidden', boxShadow: '0 0 20px rgba(202, 138, 4, 0.05)' }}>
            <table style={{ width: '100%', textAlign: 'left', borderCollapse: 'collapse' }}>
              <thead>
                <tr style={{ backgroundColor: '#111827', borderBottom: '1px solid rgba(202, 138, 4, 0.5)' }}>
                  <th style={{ padding: '16px 12px', fontSize: '12px', color: '#6b7280', textTransform: 'uppercase', letterSpacing: '1px' }}>Durum</th>
                  <th style={{ padding: '16px 12px', fontSize: '12px', color: '#6b7280', textTransform: 'uppercase', letterSpacing: '1px' }}>Host Adı</th>
                  <th style={{ padding: '16px 12px', fontSize: '12px', color: '#6b7280', textTransform: 'uppercase', letterSpacing: '1px' }}>Son Görülme</th>
                  <th style={{ padding: '16px 12px', fontSize: '12px', color: '#6b7280', textTransform: 'uppercase', letterSpacing: '1px', width: '18%' }}>CPU</th>
                  <th style={{ padding: '16px 12px', fontSize: '12px', color: '#6b7280', textTransform: 'uppercase', letterSpacing: '1px', width: '18%' }}>RAM</th>
                  <th style={{ padding: '16px 12px', fontSize: '12px', color: '#6b7280', textTransform: 'uppercase', letterSpacing: '1px', width: '18%' }}>DISK</th>
                </tr>
              </thead>
              <tbody>
                {hosts.length === 0 ? (
                  <tr>
                    <td colSpan={6} style={{ padding: '24px', textAlign: 'center', color: '#6b7280' }}>Henüz kayıtlı host bulunamadı veya ajan bekleniyor...</td>
                  </tr>
                ) : (
                  hosts.map(host => {
                    const cpu = host.cpu_usage ?? 0;
                    const ram = host.ram_usage ?? 0;
                    const disk = host.disk_usage ?? 0;
                    return (
                      <tr 
                        key={host.id} 
                        className={`host-row ${selectedHost?.id === host.id ? 'selected-row' : ''}`}
                        style={{ borderBottom: '1px solid #111827' }}
                        onClick={() => handleHostClick(host.id)} 
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
        </div>
      </main>

      {/* SAĞ PANEL: DETAYLAR, GRAFİK VE EŞİK YÖNETİMİ */}
      <div style={{
        position: 'absolute', top: 0, right: 0, bottom: 0, width: '450px',
        backgroundColor: '#111827', borderLeft: '1px solid rgba(202, 138, 4, 0.5)',
        transform: selectedHost ? 'translateX(0)' : 'translateX(100%)',
        transition: 'transform 0.3s ease-in-out',
        boxShadow: '-10px 0 30px rgba(0,0,0,0.5)',
        display: 'flex', flexDirection: 'column', zIndex: 50, overflowY: 'auto'
      }}>
        {selectedHost && (
          <div style={{ padding: '24px', display: 'flex', flexDirection: 'column', gap: '24px' }}>
            
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', borderBottom: '1px solid rgba(202, 138, 4, 0.3)', paddingBottom: '16px' }}>
              <div>
                <div style={{ fontSize: '11px', color: '#6b7280', letterSpacing: '1px' }}>HOST DETAYI</div>
                <h2 style={{ fontSize: '22px', fontWeight: 'bold', color: '#facc15', margin: '4px 0 0 0' }}>{selectedHost.hostname}</h2>
                <div style={{ fontSize: '13px', color: '#9ca3af', marginTop: '2px' }}>{selectedHost.ip_address}</div>
              </div>
              <button onClick={() => setSelectedHost(null)} style={{ background: 'none', border: 'none', color: '#9ca3af', fontSize: '22px', cursor: 'pointer' }}>×</button>
            </div>

            <div>
              <h3 style={{ fontSize: '13px', color: '#d1d5db', textTransform: 'uppercase', marginBottom: '12px' }}>Sistem Bilgisi</h3>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '12px' }}>
                <div style={{ background: '#000', padding: '12px', borderRadius: '6px', border: '1px solid rgba(202, 138, 4, 0.2)' }}>
                  <div style={{ fontSize: '10px', color: '#6b7280' }}>İŞLETİM SİSTEMİ</div>
                  <div style={{ fontSize: '13px', color: '#fff', fontWeight: 'bold', marginTop: '4px' }}>{selectedHost.os}</div>
                </div>
                <div style={{ background: '#000', padding: '12px', borderRadius: '6px', border: '1px solid rgba(202, 138, 4, 0.2)' }}>
                  <div style={{ fontSize: '10px', color: '#6b7280' }}>DURUM</div>
                  <div style={{ fontSize: '13px', color: '#4ade80', fontWeight: 'bold', marginTop: '4px', textTransform: 'uppercase' }}>{selectedHost.status}</div>
                </div>
              </div>
            </div>

            {/* DİNAMİK CPU GRAFİĞİ (RECHARTS) */}
            <div>
              <h3 style={{ fontSize: '13px', color: '#d1d5db', textTransform: 'uppercase', marginBottom: '8px' }}>CPU Kullanım Trendi (Canlı)</h3>
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

            {/* DİSK VE CPU/RAM EŞİK YÖNETİMİ (POST) */}
            <div style={{ background: 'rgba(0,0,0,0.4)', padding: '16px', borderRadius: '8px', border: '1px solid rgba(202, 138, 4, 0.3)' }}>
              <h3 style={{ fontSize: '13px', color: '#facc15', textTransform: 'uppercase', margin: '0 0 12px 0' }}>Alarm Eşikleri (Threshold)</h3>
              
              <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span style={{ fontSize: '13px', color: '#9ca3af' }}>Max CPU (%):</span>
                  <input 
                    type="number" 
                    className="threshold-input"
                    value={editThresholds.max_cpu_usage} 
                    onChange={(e) => setEditThresholds({...editThresholds, max_cpu_usage: Number(e.target.value)})}
                  />
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span style={{ fontSize: '13px', color: '#9ca3af' }}>Max RAM (%):</span>
                  <input 
                    type="number" 
                    className="threshold-input"
                    value={editThresholds.max_ram_usage} 
                    onChange={(e) => setEditThresholds({...editThresholds, max_ram_usage: Number(e.target.value)})}
                  />
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span style={{ fontSize: '13px', color: '#9ca3af' }}>Max Disk (%):</span>
                  <input 
                    type="number" 
                    className="threshold-input"
                    value={editThresholds.max_disk_usage} 
                    onChange={(e) => setEditThresholds({...editThresholds, max_disk_usage: Number(e.target.value)})}
                  />
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span style={{ fontSize: '13px', color: '#9ca3af' }}>Hata Limiti:</span>
                  <input 
                    type="number" 
                    className="threshold-input"
                    value={editThresholds.error_alert_limit} 
                    onChange={(e) => setEditThresholds({...editThresholds, error_alert_limit: Number(e.target.value)})}
                  />
                </div>

                <button 
                  onClick={handleSaveThresholds}
                  style={{ marginTop: '8px', padding: '8px', backgroundColor: '#eab308', border: 'none', color: '#000', fontWeight: 'bold', borderRadius: '4px', cursor: 'pointer' }}
                >
                  Eşikleri Güncelle
                </button>
                {saveStatus && <div style={{ fontSize: '12px', color: '#4ade80', textAlign: 'center' }}>{saveStatus}</div>}
              </div>
            </div>

            {/* OLAY / ALARM GEÇMİŞİ */}
            <div>
              <h3 style={{ fontSize: '13px', color: '#d1d5db', textTransform: 'uppercase', marginBottom: '12px' }}>Olay / Alarm Geçmişi</h3>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '8px' }}>
                <div style={{ background: '#000', padding: '10px 14px', borderRadius: '6px', borderLeft: '3px solid #eab308' }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '11px', color: '#6b7280', marginBottom: '2px' }}>
                    <span style={{ color: '#eab308', fontWeight: 'bold' }}>WARNING</span>
                    <span>Son Kontrol</span>
                  </div>
                  <div style={{ fontSize: '12px', color: '#d1d5db' }}>Ajan sistem durumu kararlı, periyodik heartbeat alındı.</div>
                </div>
              </div>
            </div>

          </div>
        )}
      </div>

      {/* AYARLAR MODALI */}
      {isSettingsOpen && (
        <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, backgroundColor: 'rgba(0, 0, 0, 0.8)', zIndex: 100, display: 'flex', justifyContent: 'center', alignItems: 'center', backdropFilter: 'blur(4px)' }}>
          <div style={{ backgroundColor: '#111827', border: '1px solid rgba(202, 138, 4, 0.5)', borderRadius: '8px', width: '450px', boxShadow: '0 0 40px rgba(202, 138, 4, 0.1)', overflow: 'hidden' }}>
            <div style={{ padding: '20px 24px', borderBottom: '1px solid rgba(202, 138, 4, 0.3)', display: 'flex', justifyContent: 'space-between', alignItems: 'center', backgroundColor: '#030712' }}>
              <h2 style={{ margin: 0, fontSize: '18px', color: '#facc15' }}>⚙️ Sistem Ayarları</h2>
              <button onClick={() => setIsSettingsOpen(false)} style={{ background: 'none', border: 'none', color: '#9ca3af', fontSize: '20px', cursor: 'pointer' }}>×</button>
            </div>
            <div style={{ padding: '24px' }}>
              <p style={{ color: '#9ca3af', fontSize: '14px', marginBottom: '16px' }}>Genel filo eşik ve yapılandırma ayarları.</p>
              <div style={{ color: '#d1d5db', fontSize: '13px' }}>Ajan bazlı eşik yönetimi için sağ panelden bir host seçebilirsiniz.</div>
            </div>
            <div style={{ padding: '16px 24px', borderTop: '1px solid rgba(202, 138, 4, 0.3)', display: 'flex', justifyContent: 'flex-end', backgroundColor: '#030712' }}>
              <button onClick={() => setIsSettingsOpen(false)} style={{ padding: '8px 24px', backgroundColor: '#eab308', border: 'none', color: '#000', borderRadius: '4px', cursor: 'pointer', fontWeight: 'bold' }}>Kapat</button>
            </div>
          </div>
        </div>
      )}

      {/* RAPOR MODALI */}
      {isExportOpen && (
        <div style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, backgroundColor: 'rgba(0, 0, 0, 0.8)', zIndex: 100, display: 'flex', justifyContent: 'center', alignItems: 'center', backdropFilter: 'blur(4px)' }}>
          <div style={{ backgroundColor: '#111827', border: '1px solid rgba(202, 138, 4, 0.5)', borderRadius: '8px', width: '450px', boxShadow: '0 0 40px rgba(202, 138, 4, 0.1)', overflow: 'hidden' }}>
            <div style={{ padding: '20px 24px', borderBottom: '1px solid rgba(202, 138, 4, 0.3)', display: 'flex', justifyContent: 'space-between', alignItems: 'center', backgroundColor: '#030712' }}>
              <h2 style={{ margin: 0, fontSize: '18px', color: '#facc15' }}>📥 Veri Dışa Aktarımı</h2>
              <button onClick={() => setIsExportOpen(false)} style={{ background: 'none', border: 'none', color: '#9ca3af', fontSize: '20px', cursor: 'pointer' }}>×</button>
            </div>
            <div style={{ padding: '24px' }}>
              <p style={{ color: '#9ca3af', fontSize: '14px', marginBottom: '24px' }}>Sistem raporunu CSV formatında indirebilirsiniz.</p>
              <button onClick={() => {
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
                setIsExportOpen(false);
              }} style={{ width: '100%', padding: '12px', backgroundColor: '#eab308', border: 'none', color: '#000', fontWeight: 'bold', borderRadius: '4px', cursor: 'pointer' }}>
                CSV Olarak İndir
              </button>
            </div>
          </div>
        </div>
      )}

    </div>
  );
}

export default App;