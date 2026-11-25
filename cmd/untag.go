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

type UntagCmd struct {
	Tags []string `short:"t" xor:"target" required:"" help:"Tags to remove (repeatable or comma-separated)"` // nolint:lll
	All  bool     `short:"a" xor:"target" required:"" help:"Remove all tags from the given paths."`          //nolint:lll

	Paths []string `arg:"" required:"" type:"existingpath" help:"Files or directories to untag"` //nolint:lll
}

func (cmd *UntagCmd) Run(ctx *AppContext) error {
	paths := make([]string, 0, len(cmd.Paths))

	for _, p := range cmd.Paths {
		realPath, err := resolvePath(p)
		if err != nil {
			return err
		}
		paths = append(paths, realPath)
	}

	runCtx := context.Background()

	if cmd.All {
		count, err := ctx.DB.UntagResourcesAll(runCtx, paths)
		if err != nil {
			return err
		}
		if ctx.Verbose {
			for _, p := range paths {
				fmt.Fprintf(os.Stderr, "untagged all from %s\n", p)
			}
		}
		resStr := "resources"
		if count == 1 {
			resStr = "resource"
		}
		fmt.Fprintf(os.Stderr, "Untagged %d %s\n", count, resStr)
		return nil
	}

	res, err := ctx.DB.UntagResources(runCtx, paths, cmd.Tags)
	if err != nil {
		return err
	}

	if ctx.Verbose {
		for _, p := range paths {
			fmt.Fprintf(os.Stderr, "untagged %s %v\n", p, cmd.Tags)
		}
	}

	resStr := "resources"
	if res.ResourcesRemoved == 1 {
		resStr = "resource"
	}

	tagStr := "tags"
	if res.TagsRemoved == 1 {
		tagStr = "tag"
	}

	fmt.Fprintf(
		os.Stderr,
		"Removed %d %s (%d %s removed)\n",
		res.TagsRemoved,
		tagStr,
		res.ResourcesRemoved,
		resStr,
	)
	return nil
}
