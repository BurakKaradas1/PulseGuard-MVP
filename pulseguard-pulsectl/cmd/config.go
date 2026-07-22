package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage PulseGuard configurations",
	Long:  `View, modify, or validate the local configuration files used by the PulseGuard system.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate [file_path]",
	Short: "Validate the syntax of a configuration file",
	Long:  `Parses the specified configuration file and checks for valid JSON syntax.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		filePath := args[0]
		fmt.Printf("[+] Validating configuration file: %s\n", filePath)

		// 1. GERÇEK KONTROL: Dosya diskte var mı?
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: Configuration file not found at '%s'\n", filePath)
			os.Exit(1)
		}

		// 2. GERÇEK OKUMA: Dosyanın içeriğini belleğe al
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Could not read file.\nDetail: %v\n", err)
			os.Exit(1)
		}

		// 3. GERÇEK DOĞRULAMA: İçerik geçerli bir JSON formatında mı?
		var configData map[string]interface{}
		if err := json.Unmarshal(fileContent, &configData); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid JSON syntax in configuration file.\nDetail: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("[✓] File exists and is readable.")
		fmt.Println("[✓] JSON syntax validation passed.")
		fmt.Println("---------------------------------------------------------")
		fmt.Printf("Result: Configuration is VALID\n")
		fmt.Println("---------------------------------------------------------")
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(validateCmd)
}
