/*
   SPDX-FileCopyrightText: 2026 Siddhant Sharma <sharmasiddhant@proton.me>

   SPDX-License-Identifier: GPL-3.0-or-later
*/

package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/alecthomas/kong"

	"github.com/shsiddhant/kothay/internal/db"
)

type AppContext struct {
	DB      *db.DB
	Logger  *log.Logger
	Verbose bool
}

type CLI struct {
	DB string `help:"Path to the database file" type:"path"`

	Verbose bool `help:"Enable verbose output" short:"v"`

	Tag  TagCmd  `cmd:"" help:"Tag files or directories."`
	Find FindCmd `cmd:"" help:"Find tagged files and directories."`
	Ls   LsCmd   `cmd:"" help:"List tags for a file/directory, or all tags."`
}

func Execute() {
	var cli CLI
	cli.DB = filepath.Join(xdg.DataHome, "kothay", "index.db")

	if len(os.Args) == 1 {
		os.Args = append(os.Args, "--help")
	}

	ctx := kong.Parse(
		&cli,
		kong.Name("kothay"),
		kong.Description("A CLI for organising files and folders with tags."),
		kong.UsageOnError(),
	)

	logger := log.New(io.Discard, "", 0)
	if cli.Verbose {
		logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	database, err := openDatabase(cli.DB, logger)
	ctx.FatalIfErrorf(err)
	defer func() {
		if err := database.Close(); err != nil {
			logger.Printf("close database: %v", err)
		}
	}()

	appCtx := &AppContext{DB: database, Logger: logger, Verbose: cli.Verbose}
	ctx.FatalIfErrorf(ctx.Run(appCtx))
}

func openDatabase(dbPath string, logger *log.Logger) (*db.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}
	database, err := db.Open(
		db.SQLiteConfig{
			DBPath:      dbPath,
			BusyTimeout: 5000,
			JournalMode: "WAL",
			ForeignKeys: true,
			Synchronous: "NORMAL",
		},
	)
	if err != nil {
		return nil, err
	}

	if err := db.RunMigrations(database, logger); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("migration runner failed: %w", err)
	}

	return database, nil
}
