package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Herhangi bir alt komut girilmediğinde çalışır
var rootCmd = &cobra.Command{
	Use:   "pulsectl",
	Short: "PulseGuard Command and Control CLI Tool",
	Long: `PulseGuard (pulsectl) is a fast and modular command-line tool 
	used to monitor agents on the network, validate configurations, 
	and retrieve system reports.`,
	// 'pulsectl' yazıldığında yardım menüsü
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

}
