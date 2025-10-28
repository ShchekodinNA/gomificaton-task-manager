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
	Short: "See where settings are stored and their validation status",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		path, err := settings.GetDefaultConfigPath()
		if err != nil {
			panic("Can't get default path")
		}

		fmt.Println("Settings path: ", path)

		fmt.Printf("Validation status: ")
		if _, err = settings.LoadConfig(&path); err != nil {
			fmt.Println("invalid")
			panic(err)
		}
		fmt.Println("valid")
	},
}

func init() {
	rootCmd.AddCommand(settingsCmd)
}
