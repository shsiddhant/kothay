/*
   SPDX-FileCopyrightText: 2026 Siddhant Sharma <sharmasiddhant@proton.me>

   SPDX-License-Identifier: GPL-3.0-or-later
*/

package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shsiddhant/kothay/internal/db"
)

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
