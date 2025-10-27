/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"gomificator/internal/utils"

	"github.com/spf13/cobra"
)

const (
	FileDestFlag      = "flile"
	SourceTypeFlag    = "source"
	DefaultSourceType = "superAppBackup"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import data from external sources",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("import called")

		fileDest, err := cmd.Flags().GetString(FileDestFlag)
		if err != nil {
			panic(err)
		}
		isSourceFileExists, err := utils.FileExists(fileDest)
		if err != nil {
			panic(err)
		}

		if !isSourceFileExists {
			fmt.Println("source file does not exist")
		}
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.Flags().StringP(FileDestFlag, "F", "", "Path to source file for import")
	importCmd.MarkFlagRequired(FileDestFlag)

	importCmd.Flags().StringP(SourceTypeFlag, "S", DefaultSourceType, "Type of source file")
}
