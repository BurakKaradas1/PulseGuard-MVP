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

type HostStats struct {
	ID       string `json:"id"`
	Hostname string `json:"hostname"`
	Status   string `json:"status"`
}

// Makine-okur rapor struct yapısı
type SystemReportJSON struct {
	GeneratedAt   string `json:"generated_at"`
	TotalAgents   int    `json:"total_agents"`
	OnlineAgents  int    `json:"online_agents"`
	OfflineAgents int    `json:"offline_agents"`
	StatusSummary string `json:"status_summary"`
}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a summary system report",
	Long:  `Connects to the PulseGuard REST API, analyzes real-time system metrics, and outputs a formatted summary report.`,
	Run: func(cmd *cobra.Command, args []string) {
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

		summary := "HEALTHY - All systems operational."
		if offlineCount > 0 {
			summary = "WARNING - Some agents are down!"
		} else if totalHosts == 0 {
			summary = "IDLE - No agents connected yet."
		}

		// Eğer --json istentiyse, makine-okur formatta struct'ı JSON'a çevirip bas
		if outputJSON {
			reportData := SystemReportJSON{
				GeneratedAt:   time.Now().Format("2006-01-02 15:04:05"),
				TotalAgents:   totalHosts,
				OnlineAgents:  onlineCount,
				OfflineAgents: offlineCount,
				StatusSummary: summary,
			}
			jsonOutput, _ := json.MarshalIndent(reportData, "", "  ")
			fmt.Println(string(jsonOutput))
			return
		}

		// İnsan-okur tablo/metin çıktısı
		fmt.Println("=========================================================")
		fmt.Println("             PULSEGUARD SYSTEM HEALTH REPORT             ")
		fmt.Println("=========================================================")
		fmt.Printf(" Generated At : %s\n", time.Now().Format("2006-01-02 15:04:05"))
		fmt.Println("---------------------------------------------------------")
		fmt.Printf(" Total Registered Agents : %d\n", totalHosts)
		fmt.Printf(" Online Agents           : %d [OK]\n", onlineCount)
		fmt.Printf(" Offline Agents          : %d [ALERT]\n", offlineCount)
		fmt.Println("---------------------------------------------------------")
		fmt.Printf(" Status Summary          : %s\n", summary)
		fmt.Println("=========================================================")
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
