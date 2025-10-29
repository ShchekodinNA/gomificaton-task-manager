/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"gomificator/internal/imprt"
	"gomificator/internal/storage"
	"os"

	"github.com/spf13/cobra"
)

const (
	FileDestFlag   = "flile"
	SourceTypeFlag = "source"
	// DefaultSourceType = "superAppBackup"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import data from external sources",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fileDest, err := cmd.Flags().GetString(FileDestFlag)
		if err != nil {
			panic(err)
		}

		importerTypeStr, err := cmd.Flags().GetString(SourceTypeFlag)
		if err != nil {
			panic(err)
		}

		importerType, err := imprt.GetImporterTypeByString(importerTypeStr)
		if err != nil {
			panic(err)
		}

		file, err := os.Open(fileDest)
		if err != nil {
			panic(err)
		}

		var importer imprt.Importer
		switch importerType {
		case imprt.ImporterTypeSuperProductivityExport:
			importer = imprt.NewImporterFromSuperProductivityExportFile(file)
		case imprt.ImporterTypeSuperProductivityBackup:
			importer = imprt.NewImporterFromSuperProductivityBackupFile(file)
		default:
			panic(fmt.Sprintf("unsupported importer type: %s", importerType))
		}

		timers, err := importer.Import()
		if err != nil {
			panic(err)
		}
		fmt.Printf("Imported %d timers\n", len(timers))
		storageService, err := storage.NewSqlliteStorage()
		if err != nil {
			panic(err)
		}

		for _, timer := range timers {
			id, err := storageService.TimersRepo.Save(timer)
			if err != nil {
				panic(err)
			}

			fmt.Printf("Imported timer with id %d\n", id)
		}
		fmt.Println("done")
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.Flags().StringP(FileDestFlag, "F", "", "Path to source file for import")
	importCmd.MarkFlagRequired(FileDestFlag)

	importCmd.Flags().StringP(SourceTypeFlag, "S", imprt.ImporterTypeSuperProductivityExport.String(), "Type of source file")

	// importCmd.Flags().StringP(SourceTypeFlag, "S", DefaultSourceType, "Type of source file")
}
