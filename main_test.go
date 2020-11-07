package main

import (
	"io"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_parseModGraph(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    *Graph
		wantErr bool
	}{
		{
			name: "basic case",
			args: args{r: strings.NewReader(`A B
A C
A D
B E
B F
C F
D G`)},
			want: &Graph{
				Nodes: map[string]*Node{
					"A": {ID: "A", Label: "A", ParentIDs: map[string]struct{}{}, ChildIDs: map[string]struct{}{"B": {}, "C": {}, "D": {}}},
					"B": {ID: "B", Label: "B", ParentIDs: map[string]struct{}{"A": {}}, ChildIDs: map[string]struct{}{"E": {}, "F": {}}},
					"C": {ID: "C", Label: "C", ParentIDs: map[string]struct{}{"A": {}}, ChildIDs: map[string]struct{}{"F": {}}},
					"D": {ID: "D", Label: "D", ParentIDs: map[string]struct{}{"A": {}}, ChildIDs: map[string]struct{}{"G": {}}},
					"E": {ID: "E", Label: "E", ParentIDs: map[string]struct{}{"B": {}}, ChildIDs: map[string]struct{}{}},
					"F": {ID: "F", Label: "F", ParentIDs: map[string]struct{}{"B": {}, "C": {}}, ChildIDs: map[string]struct{}{}},
					"G": {ID: "G", Label: "G", ParentIDs: map[string]struct{}{"D": {}}, ChildIDs: map[string]struct{}{}},
				},
				Edges: []*Edge{
					{From: "A", To: "B"},
					{From: "A", To: "C"},
					{From: "A", To: "D"},
					{From: "B", To: "E"},
					{From: "B", To: "F"},
					{From: "C", To: "F"},
					{From: "D", To: "G"},
				},
			},
			wantErr: false,
		},
		{
			name:    "a case of invalid lines in the content to be read",
			args:    args{r: strings.NewReader(`AB`)},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseModGraph(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseModGraph() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// The order of edges doesn't matter.
			sortEdges(tt.want)
			sortEdges(got)

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseModGraph() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func sortEdges(g *Graph) {
	if g == nil {
		return
	}
	sort.SliceStable(g.Edges, func(i, j int) bool { return g.Edges[i].From < g.Edges[j].From })
}

func Test_getParentAndChildName(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name       string
		args       args
		wantParent string
		wantChild  string
		wantErr    bool
	}{
		{
			name:       "valid case",
			args:       args{"A B"},
			wantParent: "A",
			wantChild:  "B",
			wantErr:    false,
		},
		{
			name:       "invalid case: only one element",
			args:       args{"AB"},
			wantParent: "",
			wantChild:  "",
			wantErr:    true,
		},
		{
			name:       "invalid case: parent element begins with a letter other than the alphabet",
			args:       args{"0A B"},
			wantParent: "",
			wantChild:  "",
			wantErr:    true,
		},
		{
			name:       "invalid case: child element begins with a letter other than the alphabet",
			args:       args{"A 0B"},
			wantParent: "",
			wantChild:  "",
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotParent, gotChild, err := getParentAndChildName(tt.args.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("getParentAndChildName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotParent != tt.wantParent {
				t.Errorf("getParentAndChildName() gotParent = %v, want %v", gotParent, tt.wantParent)
			}
			if gotChild != tt.wantChild {
				t.Errorf("getParentAndChildName() gotChild = %v, want %v", gotChild, tt.wantChild)
			}
		})
	}
}
