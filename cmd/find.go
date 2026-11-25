/*
   SPDX-FileCopyrightText: 2026 Siddhant Sharma <sharmasiddhant@proton.me>

   SPDX-License-Identifier: GPL-3.0-or-later
*/

package cmd

import (
	"context"
	"fmt"
	"os"
)

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
