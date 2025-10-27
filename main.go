/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"
	"gomificator/cmd"
	"gomificator/internal/storage"

	"github.com/pressly/goose/v3"
)

func main() {
	isStorageExists, err := storage.IsStorageExists()
	if err != nil {
		panic(err)
	}

	if !isStorageExists {

		strg, err := storage.NewSqlliteStorage()
		if err != nil {
			panic(err)
		}

		goose.SetDialect("sqlite3")
		fmt.Println(">> first launch migrations:")
		if err = storage.MigrateDb(strg); err != nil {
			panic(err)
		}
		fmt.Println(">> end of first launch migrations\n\n\n")

	}

	cmd.Execute()
}
