/*
   SPDX-FileCopyrightText: 2026 Siddhant Sharma <sharmasiddhant@proton.me>

   SPDX-License-Identifier: GPL-3.0-or-later
*/

package db

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Resource struct {
	ID        int64
	Path      string
	IsDir     bool
	CreatedAt time.Time
}

type ResourceTag struct {
	ResourceID int64
	Tag        string
}

type ResourceInput struct {
	Path  string
	IsDir bool
}

type TagResult struct {
	ResourcesTagged int
	TagsAdded       int
}

type UntagResult struct {
	ResourcesRemoved int
	TagsRemoved      int
}

// TagResources upserts resources and tags them with provided tags after
// normalizing them.
func (db *DB) TagResources(
	ctx context.Context,
	items []ResourceInput,
	rawTags []string,
) (TagResult, error) {
	var result TagResult

	if len(items) == 0 || len(rawTags) == 0 {
		return result, nil
	}

	tags := NormalizeTags(rawTags)

	if len(tags) == 0 {
		return result, nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return result, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	upsertResourceStmt, err := tx.PrepareContext(
		ctx,
		`INSERT INTO resources (path, is_dir)
		VALUES (?, ?)
		ON CONFLICT(path) DO UPDATE SET is_dir = excluded.is_dir
		RETURNING id;`,
	)
	if err != nil {
		return result, fmt.Errorf("prepare upsert resource: %w", err)
	}
	defer upsertResourceStmt.Close()

	insertTagStmt, err := tx.PrepareContext(
		ctx,
		`INSERT INTO resource_tags (resource_id, tag)
		VALUES (?, ?)
		ON CONFLICT(resource_id, tag) DO NOTHING`,
	)
	if err != nil {
		return result, fmt.Errorf("prepare insert tag: %w", err)
	}
	defer insertTagStmt.Close()

	for _, item := range items {
		isDirInt := 0
		if item.IsDir {
			isDirInt = 1
		}

		var resourceID int64
		if err := upsertResourceStmt.QueryRowContext(ctx, item.Path, isDirInt).
			Scan(&resourceID); err != nil {
			return result, fmt.Errorf("upsert resource %q: %w", item.Path, err)
		}
		result.ResourcesTagged++

		for _, tag := range tags {
			execRes, err := insertTagStmt.ExecContext(ctx, resourceID, tag)
			if err != nil {
				return result, fmt.Errorf(
					"insert tag %q for %q: %w",
					tag,
					item.Path,
					err,
				)
			}
			rows, err := execRes.RowsAffected()
			if err == nil {
				result.TagsAdded += int(rows)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return result, fmt.Errorf("commit tx: %w", err)
	}

	return result, nil

}

// NormalizeTag cleans up slashes and whitespace in tags.
// For example, "//Algebra/Group_Theory" -> "Algebra/Group_Theory"
func NormalizeTag(raw string) string {
	parts := strings.Split(raw, "/")
	cleaned := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	return strings.Join(cleaned, "/")
}

// NormalizeTags normalizes and deduplicates a slice of tag strings.
func NormalizeTags(rawTags []string) []string {
	seen := make(map[string]struct{})
	normalized := make([]string, 0, len(rawTags))

	for _, raw := range rawTags {
		norm := NormalizeTag(raw)
		if norm == "" {
			continue
		}
		key := strings.ToLower(norm)
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			normalized = append(normalized, norm)
		}
	}

	return normalized
}

// UntagResources removes given tags from given paths.
// Orphan resources, i.e. paths with no resource_tags row, are also removed.
func (db *DB) UntagResources(
	ctx context.Context,
	paths []string,
	rawTags []string,
) (UntagResult, error) {
	var result UntagResult

	if len(paths) == 0 || len(rawTags) == 0 {
		return result, nil
	}

	tags := NormalizeTags(rawTags)
	if len(tags) == 0 {
		return result, nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return result, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	deleteTagStmt, err := tx.PrepareContext(
		ctx,
		`DELETE FROM resource_tags
		WHERE resource_id = (SELECT id FROM resources WHERE path = ?)
		  AND tag = ?;`,
	)
	if err != nil {
		return result, fmt.Errorf("prepare delete tag: %w", err)
	}
	defer deleteTagStmt.Close()

	cleanupResourcesStmt, err := tx.PrepareContext(
		ctx,
		`DELETE FROM resources
		WHERE path = ?
		  AND NOT EXISTS (SELECT 1 FROM resource_tags WHERE resource_id = resources.id);`,
	)
	if err != nil {
		return result, fmt.Errorf("prepare cleanup resources: %w", err)
	}
	defer cleanupResourcesStmt.Close()

	for _, path := range paths {
		for _, tag := range tags {
			execRes, err := deleteTagStmt.ExecContext(ctx, path, tag)
			if err != nil {
				return result, fmt.Errorf("delete tag %q from %q: %w", tag, path, err)
			}
			if rows, err := execRes.RowsAffected(); err == nil {
				result.TagsRemoved += int(rows)
			}
		}

		cleanRes, err := cleanupResourcesStmt.ExecContext(ctx, path)
		if err != nil {
			return result, fmt.Errorf("cleanup resource %q: %w", path, err)
		}
		if rows, err := cleanRes.RowsAffected(); err == nil {
			result.ResourcesRemoved += int(rows)
		}
	}

	if err := tx.Commit(); err != nil {
		return result, fmt.Errorf("commit tx: %w", err)
	}
	return result, nil
}

// UntagResourcesAll completely removes given paths.
//
// ON DELETE CASCADE ensures corresponding resource_tags are removed
func (db *DB) UntagResourcesAll(ctx context.Context, paths []string) (int, error) {
	if len(paths) == 0 {
		return 0, nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `DELETE FROM resources WHERE path = ?;`)
	if err != nil {
		return 0, fmt.Errorf("prepare delete resource: %w", err)
	}
	defer stmt.Close()

	var deleted int
	for _, path := range paths {
		res, err := stmt.ExecContext(ctx, path)
		if err != nil {
			return 0, fmt.Errorf("delete resource %q: %w", path, err)
		}
		if rows, err := res.RowsAffected(); err == nil {
			deleted += int(rows)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return deleted, nil
}

// FindResources searches for resources from given raw tag segments.
func (db *DB) FindResources(
	ctx context.Context,
	rawTagSegments []string,
) ([]Resource, error) {
	var query strings.Builder
	query.WriteString(`SELECT id, path, is_dir, created_at FROM resources r`)

	args := make([]any, 0, len(rawTagSegments))
	conditions := make([]string, 0, len(rawTagSegments))

	for _, raw := range rawTagSegments {
		norm := NormalizeTag(raw)
		if norm == "" {
			continue
		}
		escaped := escapeLike(norm)

		var pattern string
		if strings.HasPrefix(strings.TrimSpace(raw), "/") {
			pattern = "/" + escaped + "/%"
		} else {
			pattern = "%/" + escaped + "/%"
		}

		conditions = append(conditions, `EXISTS (
			SELECT 1 FROM resource_tags rt
			WHERE rt.resource_id = r.id
			AND '/' || rt.tag || '/' LIKE ? ESCAPE '\'
		)`)
		args = append(args, pattern)
	}

	if len(conditions) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(conditions, " AND "))
	}

	query.WriteString(" ORDER BY r.path ASC;")

	rows, err := db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("find resources: %w", err)
	}
	defer rows.Close()

	var resources []Resource
	for rows.Next() {
		var res Resource
		var isDirInt int
		var createdUnix int64

		if err := rows.Scan(&res.ID, &res.Path, &isDirInt, &createdUnix); err != nil {
			return nil, fmt.Errorf("scan resource: %w", err)
		}
		res.IsDir = isDirInt == 1
		res.CreatedAt = time.Unix(createdUnix, 0)
		resources = append(resources, res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate resources: %w", err)
	}

	return resources, nil
}

// escapeLike escapes SQLite LIKE wildcards (\, %, _) so segment names
// such as "Number_Theory" match literally.
func escapeLike(s string) string {
	r := strings.NewReplacer(
		`\`, `\\`,
		`%`, `\%`,
		`_`, `\_`,
	)
	return r.Replace(s)
}
