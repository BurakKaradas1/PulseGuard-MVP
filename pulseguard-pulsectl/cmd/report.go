package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// HostStats struct, API'den dönecek verileri eşlemek için kullanılır
type HostStats struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a summary system report",
	Long:  `Connects to the PulseGuard REST API, analyzes real-time system metrics, and outputs a formatted summary report.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[+] Fetching system metrics from PulseGuard API...")

		apiURL := "http://localhost:8080/api/v1/dashboard/hosts"

		client := http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(apiURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Could not connect to API.\nDetail: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Fprintf(os.Stderr, "Error: Received unsuccessful response from API (HTTP %d)\n", resp.StatusCode)
			os.Exit(1)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Could not read API response.\nDetail: %v\n", err)
			os.Exit(1)
		}

		var hosts []HostStats
		if err := json.Unmarshal(body, &hosts); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Could not parse JSON data.\nDetail: %v\n", err)
			os.Exit(1)
		}

		// Metrikleri Hesaplama
		totalHosts := len(hosts)
		onlineCount := 0
		offlineCount := 0

		for _, h := range hosts {
			if h.Status == "ONLINE" || h.Status == "online" {
				onlineCount++
			} else {
				offlineCount++
			}
		}

		// Profesyonel Rapor Çıktısı
		fmt.Println("=========================================================")
		fmt.Println("             PULSEGUARD SYSTEM HEALTH REPORT             ")
		fmt.Println("=========================================================")
		fmt.Printf(" Generated At : %s\n", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Println("---------------------------------------------------------")
		fmt.Printf(" Total Registered Agents : %d\n", totalHosts)
		fmt.Printf(" Online Agents           : %d [OK]\n", onlineCount)
		fmt.Printf(" Offline Agents          : %d [ALERT]\n", offlineCount)
		fmt.Println("---------------------------------------------------------")

		if offlineCount > 0 {
			fmt.Println(" Status Summary          : WARNING - Some agents are down!")
		} else if totalHosts == 0 {
			fmt.Println(" Status Summary          : IDLE - No agents connected yet.")
		} else {
			fmt.Println(" Status Summary          : HEALTHY - All systems operational.")
		}
		fmt.Println("=========================================================")
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
