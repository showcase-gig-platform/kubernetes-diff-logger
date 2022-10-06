package differ

import (
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
)

type CmpDiffReporter struct {
	path  cmp.Path
	diffs []string
}

func (r *CmpDiffReporter) PushStep(ps cmp.PathStep) {
	r.path = append(r.path, ps)
}

func (r *CmpDiffReporter) Report(rs cmp.Result) {
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

func (r *CmpDiffReporter) PopStep() {
	r.path = r.path[:len(r.path)-1]
}

func (r *CmpDiffReporter) String(sep string) string {
	return strings.Join(r.diffs, sep)
}

func (r *CmpDiffReporter) MapIndexString() string {
	ps := []string{""}
	for _, s := range r.path {
		if mi, ok := s.(cmp.MapIndex); ok {
			val := fmt.Sprintf("%v", mi.Key())
			if strings.Contains(val, ".") { // labelとかのkeyにドットが入ってると見分けが付かないので囲む
				val = fmt.Sprintf("[%v]", val)
			}
			ps = append(ps, fmt.Sprintf(".%v", val))
		}
		if si, ok := s.(cmp.SliceIndex); ok {
			var i int
			vx, vy := si.SplitKeys()
			switch {
			case vx >= 0 && vy == -1:
				i = vx
			case vx == -1 && vy >= 0:
				i = vy
			default:
				i = vx
			}
			ps = append(ps, fmt.Sprintf("[%v]", i))
		}
	}
	return strings.Join(ps, "")
}
