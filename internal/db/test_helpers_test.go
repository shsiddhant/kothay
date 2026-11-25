/*
   SPDX-FileCopyrightText: 2026 Siddhant Sharma <sharmasiddhant@proton.me>

   SPDX-License-Identifier: GPL-3.0-or-later
*/

package db

import "testing"

func testDB(t *testing.T) *DB {
	t.Helper()

	cfg := SQLiteConfig{
		DBPath:      ":memory:",
		ForeignKeys: true,
	}

	database, err := Open(cfg)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		database.Close()
	})

	if err := RunMigrations(database, nil); err != nil {
		t.Fatal(err)
	}

	return database
}
