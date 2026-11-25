/*
   SPDX-FileCopyrightText: 2026 Siddhant Sharma <sharmasiddhant@proton.me>

   SPDX-License-Identifier: GPL-3.0-or-later
*/

package cmd

import "fmt"

type VersionCmd struct{}

func (cmd *VersionCmd) Run(ctx *AppContext) error {
	fmt.Println(ctx.Version.Detailed())
	return nil
}
