/*
   SPDX-FileCopyrightText: 2026 Siddhant Sharma <sharmasiddhant@proton.me>

   SPDX-License-Identifier: GPL-3.0-or-later
*/

package main

import "github.com/shsiddhant/kothay/cmd"

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func main() {
	cmd.Execute(
		cmd.VersionInfo{
			AppName: "kothay",
			Author:  "Siddhant Sharma",
			Version: Version,
			Commit:  Commit,
			Date:    Date,
		},
	)
}
