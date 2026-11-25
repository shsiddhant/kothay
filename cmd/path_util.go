/*
   SPDX-FileCopyrightText: 2026 Siddhant Sharma <sharmasiddhant@proton.me>

   SPDX-License-Identifier: GPL-3.0-or-later
*/

package cmd

import (
	"fmt"
	"path/filepath"
)

func resolvePath(p string) (string, error) {
	absPath, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("resolve path %q: %w", p, err)
	}

	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return "", fmt.Errorf("eval symlinks %q: %w", absPath, err)
	}

	return realPath, err
}
