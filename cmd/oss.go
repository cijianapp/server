package cmd

import (
	"fmt"

	"github.com/cijianapp/server/oss"
	"github.com/spf13/cobra"
)

// ossCmd represents the oss command
var ossCmd = &cobra.Command{
	Use:   "oss",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("oss called")

		oss.PutObject("")
	},
}

func init() {
	rootCmd.AddCommand(ossCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// ossCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// ossCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
