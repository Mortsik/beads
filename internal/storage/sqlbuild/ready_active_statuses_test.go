package sqlbuild

import (
	"reflect"
	"strings"
	"testing"

	"github.com/steveyegge/beads/internal/types"
)

// TestBuildReadyWorkWhereActiveCustomStatuses pins the three readings of the
// singular StatusOpen pin: unresolved names keep the custom_statuses
// subquery; resolved names are bound as parameters (#8045 on this line).
func TestBuildReadyWorkWhereActiveCustomStatuses(t *testing.T) {
	filter := types.WorkFilter{Status: types.StatusOpen}
	for _, tc := range []struct {
		name       string
		active     []string
		wantClause string
		wantArgs   []any
	}{
		{name: "unresolved keeps the subquery", active: nil,
			wantClause: "(status = ? OR status IN (SELECT name FROM custom_statuses WHERE category = 'active'))", wantArgs: []any{"open"}},
		{name: "none active binds equality", active: []string{},
			wantClause: "status = ?", wantArgs: []any{"open"}},
		{name: "active names are bound, deduped, empties dropped", active: []string{"review", "open", "", "review", "qa"},
			wantClause: "status IN (?,?,?)", wantArgs: []any{"open", "review", "qa"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			where, args, err := BuildReadyWorkWhere(filter, IssuesFilterTables, ReadyWorkWhereInputs{ActiveCustomStatuses: tc.active})
			if err != nil {
				t.Fatalf("BuildReadyWorkWhere: %v", err)
			}
			if !strings.HasPrefix(where, "WHERE "+tc.wantClause+" AND ") {
				t.Fatalf("where = %q, want prefix %q", where, tc.wantClause)
			}
			if got := args[:len(tc.wantArgs)]; !reflect.DeepEqual(got, tc.wantArgs) {
				t.Fatalf("status args = %v, want %v", got, tc.wantArgs)
			}
		})
	}
}

// TestBuildReadyWorkWhereActiveCustomStatusesOnlyForOpen pins that resolved
// names never widen any other selector.
func TestBuildReadyWorkWhereActiveCustomStatusesOnlyForOpen(t *testing.T) {
	where, args, err := BuildReadyWorkWhere(types.WorkFilter{Status: types.StatusInProgress}, IssuesFilterTables, ReadyWorkWhereInputs{ActiveCustomStatuses: []string{"review"}})
	if err != nil {
		t.Fatalf("BuildReadyWorkWhere: %v", err)
	}
	if !strings.HasPrefix(where, "WHERE status = ? AND ") || args[0] != "in_progress" {
		t.Fatalf("where = %q args = %v, want exact in_progress", where, args)
	}
}
