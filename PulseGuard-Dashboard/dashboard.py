import requests
import time
import os

def clear_screen():
    # İşletim sistemine göre terminali temizler
    os.system('cls' if os.name == 'nt' else 'clear')

def main():
    url = "http://localhost:8080/stats"
    
    clear_screen()
    print("[*] PulseGuard C2 Dashboard Başlatılıyor...")
    time.sleep(1)

    while True:
        try:
            # Go C2 sunucumuzdan JSON verisini çekiyoruz
            response = requests.get(url)
            
            if response.status_code == 200:
                data = response.json()
                
                clear_screen()
                print("=" * 60)
                print(" 🛡️  PULSEGUARD COMMAND & CONTROL (C2) DASHBOARD 🛡️")
                print("=" * 60)
                print(f" 📡 Bağlantı Durumu     : AKTİF (Şifreli Tünel)")
                print(f" 💻 Hedef CPU Kullanımı : % {data.get('CPUUsage', 0):.2f}")
                print(f" 🧠 Hedef RAM Kullanımı : % {data.get('RAMUsage', 0):.2f}")
                print(f" 🔌 Açık Port Sayısı    : {len(data.get('OpenPorts') or [])}")
                print("-" * 60)
                
                # Ekstra bir analiz şovu:
                if data.get('CPUUsage', 0) > 80:
                    print(" [!] ALARM: Hedef sistemde aşırı CPU yüklenmesi tespit edildi!")
                
                print(" [*] C2 Sunucusu dinleniyor... (Canlı Güncelleme)")
                print("=" * 60)
            else:
                print(f"[!] Hata: Sunucu {response.status_code} kodu döndürdü.")
                
        except requests.exceptions.ConnectionError:
            clear_screen()
            print("[!] C2 Sunucusuna (Go) ulaşılamıyor. Lütfen önce C2'yi başlatın!")
            print(f"Hedef URL: {url}")

        # Sistemi yormamak için saniyede 1 kez güncelle
        time.sleep(1)

if __name__ == "__main__":
    main()