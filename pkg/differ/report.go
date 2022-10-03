package differ

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
)

type SpecDiffReporter struct {
	path  cmp.Path
	diffs []string
}

func (r *SpecDiffReporter) PushStep(ps cmp.PathStep) {
	r.path = append(r.path, ps)
}

func (r *SpecDiffReporter) Report(rs cmp.Result) {
	if !rs.Equal() {
		vx, vy := r.path.Last().Values()
		r.diffs = append(r.diffs, fmt.Sprintf("%v: %v -> %v", r.MapIndexString(), vx, vy))
	}
}

func (r *SpecDiffReporter) PopStep() {
	r.path = r.path[:len(r.path)-1]
}

func (r *SpecDiffReporter) String(sep string) string {
	return strings.Join(r.diffs, sep)
}

func (r *SpecDiffReporter) MapIndexString() string {
	ps := []string{"spec"}
	for _, s := range r.path {
		if i, ok := s.(cmp.MapIndex); ok {
			ps = append(ps, fmt.Sprintf("%v", i.Key()))
		}
	}
	return strings.Join(ps, ".")
}
