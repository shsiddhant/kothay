/*
   SPDX-FileCopyrightText: 2026 Siddhant Sharma <sharmasiddhant@proton.me>

   SPDX-License-Identifier: GPL-3.0-or-later
*/

package main

import (
	"context"
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

type TagCmd struct {
	Tags []string `short:"t" required:"" help:"Tags to attach (repeatable or comma-separated)"` // nolint:lll

	Paths []string `arg:"" required:"" type:"existingpath" help:"Files or directories to tag"` //nolint:lll
}

func (cmd *TagCmd) Run(ctx *AppContext) error {
	items := make([]db.ResourceInput, 0, len(cmd.Paths))

	for _, p := range cmd.Paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			return fmt.Errorf("resolve path %q: %w", p, err)
		}

		realPath, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return fmt.Errorf("eval symlinks %q: %w", absPath, err)
		}

		info, err := os.Stat(realPath)
		if err != nil {
			return fmt.Errorf("stat %q: %w", realPath, err)
		}

		items = append(items, db.ResourceInput{
			Path:  realPath,
			IsDir: info.IsDir(),
		})
	}

	res, err := ctx.DB.TagResources(context.Background(), items, cmd.Tags)
	if err != nil {
		return err
	}

	if ctx.Verbose {
		for _, item := range items {
			fmt.Fprintf(os.Stderr, "tagged %s %v\n", item.Path, cmd.Tags)
		}
	}

	resStr := "resources"
	if res.ResourcesTagged == 1 {
		resStr = "resource"
	}

	tagStr := "tags"
	if res.TagsAdded == 1 {
		tagStr = "tag"
	}

	fmt.Fprintf(
		os.Stderr,
		"Tagged %d %s (%d new %s added)\n",
		res.ResourcesTagged,
		resStr,
		res.TagsAdded,
		tagStr,
	)
	return nil
}

type FindCmd struct {
	Tags []string `arg:"" optional:"" help:"Tag segments to match. Omit to list all tagged files/directories."` //nolint:lll

	Files bool `short:"f" xor:"kind" help:"Only show files."`
	Dirs  bool `short:"d" xor:"kind" help:"Only show directories."`
}

func (cmd *FindCmd) Run(ctx *AppContext) error {
	resources, err := ctx.DB.FindResources(context.Background(), cmd.Tags)
	if err != nil {
		return err
	}

	var matched int
	for _, r := range resources {
		if cmd.Files && r.IsDir {
			continue
		}
		if cmd.Dirs && !r.IsDir {
			continue
		}

		fmt.Println(r.Path)
		matched++
	}

	if ctx.Verbose {
		resStr := "resources"
		if matched == 1 {
			resStr = "resource"
		}
		fmt.Fprintf(os.Stderr, "Found %d %s\n", matched, resStr)
	}
	return nil
}

type CLI struct {
	DB string `help:"Path to the database file" type:"path"`

	Verbose bool `help:"Enable verbose output" short:"v"`

	Tag  TagCmd  `cmd:"" help:"Tag files or directories."`
	Find FindCmd `cmd:"" help:"Find tagged files and directories."`
}

func main() {
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
