import React, { useState, useEffect } from 'react';
import jsPDF from 'jspdf';
import autoTable from 'jspdf-autotable';
import type { HostStatus } from '../types';

interface ExportModalProps {
  isOpen: boolean;
  onClose: () => void;
  hosts: HostStatus[];
}

export const ExportModal: React.FC<ExportModalProps> = ({ isOpen, onClose, hosts }) => {
  // Animasyon state'leri
  const [isVisible, setIsVisible] = useState(false);
  const [isClosing, setIsClosing] = useState(false);

  // Modal açıldığında giriş animasyonunu tetikle
  useEffect(() => {
    if (isOpen) {
      setIsClosing(false);
      // DOM'a eklendikten 10ms sonra animasyonu başlat
      const timer = setTimeout(() => setIsVisible(true), 10);
      return () => clearTimeout(timer);
    } else {
      setIsVisible(false);
    }
  }, [isOpen]);

  // Zarif kapanış fonksiyonu
  const handleCloseClick = () => {
    setIsClosing(true);
    setIsVisible(false); // Çıkış animasyonunu başlat
    setTimeout(() => {
      setIsClosing(false); 
      onClose(); // Animasyon bittikten sonra (300ms) ana state'i kapat
    }, 300);
  };

  // Eğer kapalıysa ve kapanma animasyonunda değilse render etme
  if (!isOpen && !isClosing) return null;

  // Mevcut CSV İndirme Fonksiyonu
  const handleExportCSV = () => {
    const headers = "ID,Hostname,Status,CPU,RAM,DISK,Last Seen\n";
    const rows = hosts.map(h => `${h.id},${h.hostname},${h.status},${h.cpu_usage || 0},${h.ram_usage || 0},${h.disk_usage || 0},${h.last_seen}`).join("\n");
    const blob = new Blob([headers + rows], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', 'PulseGuard_Report.csv');
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    handleCloseClick(); // Anında kapanmak yerine animasyonla kapat
  };

  // PDF İndirme Fonksiyonu
  const handleExportPDF = () => {
    const doc = new jsPDF();
    
    // Başlık Ekleme
    doc.setFontSize(16);
    doc.text("PulseGuard Fleet Status Report", 14, 15);
    
    const tableColumn = ["Hostname", "Status", "CPU (%)", "RAM (%)", "DISK (%)", "Last Seen"];
    const tableRows = hosts.map(host => [
      host.hostname,
      host.status.toUpperCase(),
      host.cpu_usage?.toString() || "0",
      host.ram_usage?.toString() || "0",
      host.disk_usage?.toString() || "0",
      new Date(host.last_seen).toLocaleString('en-US')
    ]);

    // Tabloyu PDF'e Çizdirme
    autoTable(doc, {
      head: [tableColumn],
      body: tableRows,
      startY: 25, 
      theme: 'grid',
      styles: { fontSize: 10, cellPadding: 3 },
      headStyles: { fillColor: [234, 179, 8], textColor: [0, 0, 0], fontStyle: 'bold' }, 
    });

    // Dosyayı Kaydet
    doc.save('PulseGuard_Report.pdf');
    handleCloseClick(); // Anında kapanmak yerine animasyonla kapat
  };

  return (
    <div style={{ 
      position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, 
      backgroundColor: 'rgba(0, 0, 0, 0.8)', zIndex: 100, display: 'flex', justifyContent: 'center', alignItems: 'center', backdropFilter: 'blur(4px)',
      // Arka plan kararma animasyonu
      opacity: isVisible ? 1 : 0,
      transition: 'opacity 0.3s ease-in-out'
    }}>
      <div style={{ 
        backgroundColor: '#111827', border: '1px solid rgba(202, 138, 4, 0.5)', borderRadius: '8px', width: '450px', boxShadow: '0 0 40px rgba(202, 138, 4, 0.1)', overflow: 'hidden',
        // Modalın kendisinin büyüme ve yukarı kayma animasyonu
        transform: isVisible ? 'scale(1) translateY(0)' : 'scale(0.95) translateY(20px)',
        opacity: isVisible ? 1 : 0,
        transition: 'all 0.3s cubic-bezier(0.4, 0, 0.2, 1)'
      }}>
        <div style={{ padding: '20px 24px', borderBottom: '1px solid rgba(202, 138, 4, 0.3)', display: 'flex', justifyContent: 'space-between', alignItems: 'center', backgroundColor: '#030712' }}>
          <h2 style={{ margin: 0, fontSize: '18px', color: '#facc15' }}>📥 Export Data</h2>
          <button 
            onClick={handleCloseClick} 
            style={{ background: 'none', border: 'none', color: '#9ca3af', fontSize: '20px', cursor: 'pointer', transition: 'color 0.2s' }}
            onMouseEnter={(e) => e.currentTarget.style.color = '#facc15'}
            onMouseLeave={(e) => e.currentTarget.style.color = '#9ca3af'}
          >
            ×
          </button>
        </div>
        <div style={{ padding: '24px', display: 'flex', flexDirection: 'column', gap: '16px' }}>
          <p style={{ color: '#9ca3af', fontSize: '14px', margin: 0 }}>You can download the system report in your preferred format.</p>
          
          <button onClick={handleExportCSV} style={{ width: '100%', padding: '12px', backgroundColor: '#374151', border: '1px solid #9ca3af', color: '#fff', fontWeight: 'bold', borderRadius: '4px', cursor: 'pointer', transition: 'all 0.2s' }}>
            📄 Export as CSV
          </button>
          
          <button onClick={handleExportPDF} style={{ width: '100%', padding: '12px', backgroundColor: '#eab308', border: 'none', color: '#000', fontWeight: 'bold', borderRadius: '4px', cursor: 'pointer', transition: 'all 0.2s' }}>
            📕 Export as PDF
          </button>
        </div>
      </div>
    </div>
  );
};