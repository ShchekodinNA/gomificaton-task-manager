/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"gomificator/internal/settings"

	"github.com/spf13/cobra"
)

var settingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Where yaml file stored",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		path, err := settings.GetDefaultConfigPath()
		if err != nil {
			panic("Can't get default path")
		}

		fmt.Println("Settings stored here:")
		fmt.Println(path)
	},
}

func init() {
	rootCmd.AddCommand(settingsCmd)
}
