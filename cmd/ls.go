/*
   SPDX-FileCopyrightText: 2026 Siddhant Sharma <sharmasiddhant@proton.me>

   SPDX-License-Identifier: GPL-3.0-or-later
*/

package cmd

import (
	"context"
	"fmt"
	"path/filepath"
)

type LsCmd struct {
	Path string `arg:"" optional:"" type:"existingpath" help:"File or directory to check. Omit to list all tags."` //nolint:lll
}

func (cmd *LsCmd) Run(ctx *AppContext) error {
	runCtx := context.Background()

	if cmd.Path != "" {
		absPath, err := filepath.Abs(cmd.Path)
		if err != nil {
			return fmt.Errorf("resolve path %q: %w", cmd.Path, err)
		}

		realPath, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return fmt.Errorf("eval symlinks %q: %w", absPath, err)
		}

		tags, err := ctx.DB.GetTagsForResource(runCtx, realPath)
		if err != nil {
			return err
		}
		for _, tag := range tags {
			fmt.Println(tag)
		}
		return nil
	}

	counts, err := ctx.DB.GetTagCounts(runCtx, "")
	if err != nil {
		return err
	}
	for _, tc := range counts {
		if ctx.Verbose {
			fmt.Printf("%d\t%s\n", tc.Count, tc.Tag)
		} else {
			fmt.Println(tc.Tag)
		}
	}
	return nil
}
