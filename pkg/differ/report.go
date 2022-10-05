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
		switch true {
		case !vx.IsValid() && vy.IsValid():
			r.diffs = append(r.diffs, fmt.Sprintf("%v: %v (added)", r.MapIndexString(), vy))
		case vx.IsValid() && !vy.IsValid():
			r.diffs = append(r.diffs, fmt.Sprintf("%v: %v (deleted)", r.MapIndexString(), vx))
		default:
			r.diffs = append(r.diffs, fmt.Sprintf("%v: %v -> %v", r.MapIndexString(), vx, vy))
		}
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
		if mi, ok := s.(cmp.MapIndex); ok {
			ps = append(ps, fmt.Sprintf(".%v", mi.Key()))
		}
		if si, ok := s.(cmp.SliceIndex); ok {
			var i int
			vx, vy := si.SplitKeys()
			switch {
			case vx > 0 && vy == -1:
				i = vx
			case vx == -1 && vy > 0:
				i = vy
			default:
				i = vx
			}
			ps = append(ps, fmt.Sprintf("[%v]", i))
		}
	}
	return strings.Join(ps, "")
}
