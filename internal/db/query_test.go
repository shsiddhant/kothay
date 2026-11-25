/*
   SPDX-FileCopyrightText: 2026 Siddhant Sharma <sharmasiddhant@proton.me>

   SPDX-License-Identifier: GPL-3.0-or-later
*/

package db

import (
	"context"
	"testing"
)

func TestTagAndFindResources(t *testing.T) {
	ctx := context.Background()
	database := testDB(t)

	items := []ResourceInput{
		{Path: "/math/Silverman.pdf", IsDir: false},
		{Path: "/math/Boston_Paper.pdf", IsDir: false},
		{Path: "/code/mone-pore", IsDir: true},
	}

	res1, err := database.TagResources(ctx, items[:1], []string{
		" //Math//Number_Theory/Elliptic_Curves/ ",
		"Type/Textbook",
		"math/number_theory/elliptic_curves",
	})
	if err != nil {
		t.Fatalf("TagResources failed: %v", err)
	}
	if res1.ResourcesTagged != 1 || res1.TagsAdded != 2 {
		t.Fatalf("expected 1 resource and 2 tags added, got %+v", res1)
	}

	resDup, err := database.TagResources(ctx, items[:1], []string{"TYPE/TEXTBOOK"})
	if err != nil {
		t.Fatalf("duplicate TagResources failed: %v", err)
	}
	if resDup.TagsAdded != 0 {
		t.Fatalf("expected 0 tags added on duplicate, got %d", resDup.TagsAdded)
	}

	if _, err := database.TagResources(
		ctx,
		items[1:2],
		[]string{"Math/Number_Theory", "Type/Paper"},
	); err != nil {
		t.Fatalf("tag Boston paper: %v", err)
	}
	if _, err := database.TagResources(
		ctx,
		items[2:3],
		[]string{"Lang/Go", "Books/Math"},
	); err != nil {
		t.Fatalf("tag code repo: %v", err)
	}

	tests := []struct {
		name     string
		segments []string
		want     []string
	}{
		{
			name:     "no filter returns all sorted by path",
			segments: nil,
			want: []string{
				"/code/mone-pore",
				"/math/Boston_Paper.pdf",
				"/math/Silverman.pdf",
			},
		},
		{
			name:     "case-insensitive mid-segment match with underscore",
			segments: []string{"number_theory"},
			want:     []string{"/math/Boston_Paper.pdf", "/math/Silverman.pdf"},
		},
		{
			name:     "multi-tag AND intersection",
			segments: []string{"number_theory", "paper"},
			want:     []string{"/math/Boston_Paper.pdf"},
		},
		{
			name:     "unanchored segment matches anywhere in hierarchy",
			segments: []string{"Math"},
			want: []string{
				"/code/mone-pore",
				"/math/Boston_Paper.pdf",
				"/math/Silverman.pdf",
			},
		},
		{
			name:     "leading slash anchors match to root",
			segments: []string{"/Math"},
			want:     []string{"/math/Boston_Paper.pdf", "/math/Silverman.pdf"},
		},
		{
			name:     "partial word should not match segment",
			segments: []string{"Theory"},
			want:     nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := database.FindResources(ctx, tc.segments)
			if err != nil {
				t.Fatalf("FindResources(%v) error: %v", tc.segments, err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf(
					"FindResources(%v) got %d results, want %d",
					tc.segments,
					len(got),
					len(tc.want),
				)
			}
			for i := range got {
				if got[i].Path != tc.want[i] {
					t.Errorf("result[%d] = %q, want %q", i, got[i].Path, tc.want[i])
				}
			}
		})
	}
}

func TestUntagAndOrphanCleanup(t *testing.T) {
	ctx := context.Background()
	database := testDB(t)

	items := []ResourceInput{
		{Path: "/math/A.pdf", IsDir: false},
		{Path: "/math/B.pdf", IsDir: false},
	}
	if _, err := database.TagResources(
		ctx,
		items,
		[]string{"Algebra", "Type/Paper"},
	); err != nil {
		t.Fatalf("seed tags: %v", err)
	}

	u1, err := database.UntagResources(
		ctx,
		[]string{"/math/A.pdf"},
		[]string{"algebra"},
	)
	if err != nil {
		t.Fatalf("UntagResources failed: %v", err)
	}
	if u1.TagsRemoved != 1 || u1.ResourcesRemoved != 0 {
		t.Fatalf("expected 1 tag removed and 0 resources deleted, got %+v", u1)
	}

	u2, err := database.UntagResources(
		ctx,
		[]string{"/math/A.pdf"},
		[]string{"Type/Paper"},
	)
	if err != nil {
		t.Fatalf("UntagResources last tag failed: %v", err)
	}
	if u2.TagsRemoved != 1 || u2.ResourcesRemoved != 1 {
		t.Fatalf("expected 1 tag removed and 1 resource deleted, got %+v", u2)
	}

	deleted, err := database.UntagResourcesAll(ctx, []string{"/math/B.pdf"})
	if err != nil {
		t.Fatalf("UntagAllResources failed: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("expected 1 deleted resource, got %d", deleted)
	}

	remaining, err := database.FindResources(ctx, nil)
	if err != nil {
		t.Fatalf("FindResources failed: %v", err)
	}
	if len(remaining) != 0 {
		t.Fatalf("expected 0 remaining resources, got %d", len(remaining))
	}
}

func TestGetTagsForResource(t *testing.T) {
	ctx := context.Background()
	database := testDB(t)

	items := []ResourceInput{
		{Path: "/math/Silverman.pdf", IsDir: false},
	}
	if _, err := database.TagResources(ctx, items, []string{
		"Type/Textbook",
		"Math/Number_Theory/Elliptic_Curves",
		"Algebra",
	}); err != nil {
		t.Fatalf("seed tags: %v", err)
	}

	got, err := database.GetTagsForResource(ctx, "/math/Silverman.pdf")
	if err != nil {
		t.Fatalf("GetTagsForResource failed: %v", err)
	}
	want := []string{
		"Algebra",
		"Math/Number_Theory/Elliptic_Curves",
		"Type/Textbook",
	}
	if len(got) != len(want) {
		t.Fatalf("got %d tags (%v), want %d (%v)", len(got), got, len(want), want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("tag[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	untracked, err := database.GetTagsForResource(ctx, "/math/DoesNotExist.pdf")
	if err != nil {
		t.Fatalf("GetTagsForResource untracked failed: %v", err)
	}
	if len(untracked) != 0 {
		t.Fatalf("expected 0 tags for untracked file, got %v", untracked)
	}
}

func TestGetTagCounts(t *testing.T) {
	ctx := context.Background()
	database := testDB(t)

	if _, err := database.TagResources(ctx, []ResourceInput{
		{Path: "/math/Silverman.pdf", IsDir: false},
	}, []string{"Math/Number_Theory/Elliptic_Curves", "Type/Textbook"}); err != nil {
		t.Fatalf("seed 1: %v", err)
	}
	if _, err := database.TagResources(ctx, []ResourceInput{
		{Path: "/math/Ireland_Rosen.djvu", IsDir: false},
	}, []string{"Math/Number_Theory", "Type/Textbook"}); err != nil {
		t.Fatalf("seed 2: %v", err)
	}
	if _, err := database.TagResources(ctx, []ResourceInput{
		{Path: "/cs/Koblitz_Crypto.pdf", IsDir: false},
	}, []string{"CS/Cryptography/Elliptic_Curves"}); err != nil {
		t.Fatalf("seed 3: %v", err)
	}

	tests := []struct {
		name    string
		segment string
		want    []TagCount
	}{
		{
			name:    "empty filter returns all tags and counts sorted",
			segment: "",
			want: []TagCount{
				{Tag: "CS/Cryptography/Elliptic_Curves", Count: 1},
				{Tag: "Math/Number_Theory", Count: 1},
				{Tag: "Math/Number_Theory/Elliptic_Curves", Count: 1},
				{Tag: "Type/Textbook", Count: 2},
			},
		},
		{
			name:    "case-insensitive mid-segment filter",
			segment: "number_theory",
			want: []TagCount{
				{Tag: "Math/Number_Theory", Count: 1},
				{Tag: "Math/Number_Theory/Elliptic_Curves", Count: 1},
			},
		},
		{
			name:    "leaf segment matches across different root trees",
			segment: "Elliptic_Curves",
			want: []TagCount{
				{Tag: "CS/Cryptography/Elliptic_Curves", Count: 1},
				{Tag: "Math/Number_Theory/Elliptic_Curves", Count: 1},
			},
		},
		{
			name:    "leading slash anchors filter to root",
			segment: "/Math",
			want: []TagCount{
				{Tag: "Math/Number_Theory", Count: 1},
				{Tag: "Math/Number_Theory/Elliptic_Curves", Count: 1},
			},
		},
		{
			name:    "non-matching segment returns empty slice",
			segment: "Topology",
			want:    nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := database.GetTagCounts(ctx, tc.segment)
			if err != nil {
				t.Fatalf("GetTagCounts(%q) error: %v", tc.segment, err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("GetTagCounts(%q) got %+v, want %+v", tc.segment, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("count[%d] = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}
