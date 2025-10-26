/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"gomificator/cmd"
	"gomificator/internal/storage"

	"github.com/pressly/goose/v3"
)

func main() {
	strg, err := storage.NewSqlliteStorage()
	if err != nil {
		panic(err)
	}

	goose.SetLogger(goose.NopLogger())
	goose.SetDialect("sqlite3")

	if err = storage.MigrateDb(strg); err != nil {
		panic(err)
	}

	cmd.Execute()
}
