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

type Host struct {
	ID        string `json:"id"`
	Hostname  string `json:"hostname"`
	Status    string `json:"status"`
	UpdatedAt string `json:"last_seen"`
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the connection status of agents",
	Long:  `Queries the real-time status of registered agents by connecting to the PulseGuard REST API.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("[+] Connecting to PulseGuard API...")

		apiURL := serverURL + "/api/v1/dashboard/hosts"

		// client burada tanımlandı
		client := http.Client{Timeout: 5 * time.Second}

		resp, err := client.Get(apiURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Could not connect to API.\nDetail: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			// %d formatı eklendi
			fmt.Fprintf(os.Stderr, "Error: Received unsuccessful response from API (HTTP %d)\n", resp.StatusCode)
			os.Exit(1)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			// fmt.fprintf -> fmt.Fprintf olarak düzeltildi ve sondaki fazla tırnak kaldırıldı
			fmt.Fprintf(os.Stderr, "Error: Could not read API response.\nDetail: %v\n", err)
			os.Exit(1)
		}

		var hosts []Host
		if err := json.Unmarshal(body, &hosts); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Could not parse JSON data.\nDetail: %v\n", err)
			os.Exit(1)
		}

		// Eğer kullanıcı --json bayrağını verdiyse, tabloyla uğraşmadan ham JSON'ı bas ve çık
		if outputJSON {
			fmt.Println(string(body))
			return
		}

		fmt.Println("---------------------------------------------------------")
		fmt.Printf("%-20s %-15s %-20s\n", "HOSTNAME", "STATUS", "LAST SEEN")
		fmt.Println("---------------------------------------------------------")

		if len(hosts) == 0 {
			fmt.Println("No agents registered in the system yet.")
		} else {
			for _, h := range hosts {
				statusDisplay := fmt.Sprintf("[%s]", h.Status)
				fmt.Printf("%-20s %-15s %-20s\n", h.Hostname, statusDisplay, h.UpdatedAt)
			}
		}
		fmt.Println("---------------------------------------------------------")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
