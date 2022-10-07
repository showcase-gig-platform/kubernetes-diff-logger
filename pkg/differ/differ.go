package differ

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"regexp"

	"github.com/google/go-cmp/cmp"
	"github.com/showcase-gig-platform/kubernetes-diff-logger/pkg/config"
	"github.com/showcase-gig-platform/kubernetes-diff-logger/pkg/wrapper"
)

// Differ is responsible for subscribing to an informer an filtering out events
type Differ struct {
	wrap             wrapper.Wrap
	informer         cache.SharedInformer
	output           Output
	labelConfig      config.ExtraConfig
	annotationConfig config.ExtraConfig
	matchRegexp      string
	ignoreRegexp     string
}

// NewDiffer constructs a Differ
func NewDiffer(f wrapper.Wrap, i cache.SharedInformer, o Output, l, a config.ExtraConfig, mr, ir string) *Differ {
	d := &Differ{
		wrap:             f,
		informer:         i,
		output:           o,
		labelConfig:      l,
		annotationConfig: a,
		matchRegexp:      mr,
		ignoreRegexp:     ir,
	}

	return d
}

// Run sets up eventhandlers, sync informer caches and blocks until stop is closed
func (d *Differ) Run(stopCh <-chan struct{}) error {
	defer runtime.HandleCrash()

	if ok := cache.WaitForCacheSync(stopCh, d.informer.HasSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	d.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    d.added,
		UpdateFunc: d.updated,
		DeleteFunc: d.deleted,
	})

	<-stopCh

	return nil
}

func (d *Differ) added(added interface{}) {
	object := d.mustWrap(added)

	if d.matches(object) {
		meta := object.GetMetadata()

		d.output.WriteAdded(meta.Name, meta.Namespace, object.GetKind())
	}
}

func (d *Differ) updated(old interface{}, new interface{}) {
	oldObject := d.mustWrap(old)
	newObject := d.mustWrap(new)

	if d.matches(oldObject) || d.matches(newObject) {
		oldTarget := d.createDiffObject(oldObject)
		newTarget := d.createDiffObject(newObject)

		var r CmpDiffReporter
		cmp.Diff(oldTarget, newTarget, cmp.Reporter(&r))
		if len(r.diffs) > 0 {
			meta := newObject.GetMetadata()
			d.output.WriteUpdated(meta.Name, meta.Namespace, newObject.GetKind(), r.diffs)
		}
	}
}

func (d *Differ) deleted(deleted interface{}) {
	object := d.mustWrap(deleted)

	if d.matches(object) {
		meta := object.GetMetadata()

		d.output.WriteDeleted(meta.Name, meta.Namespace, object.GetKind())
	}
}

func (d *Differ) matches(o wrapper.KubernetesObject) bool {
	name := o.GetMetadata().Name
	if d.ignoreRegexp != "" {
		m, e := regexp.MatchString(d.ignoreRegexp, name)
		if e != nil {
			klog.Errorf("Failed to eval regexp : %v\n", e.Error())
		}
		if m {
			return false
		}
	}
	if d.matchRegexp != "" {
		m, e := regexp.MatchString(d.matchRegexp, name)
		if e != nil {
			klog.Errorf("Failed to eval regexp : %v\n", e.Error())
		}
		if !m {
			return false
		}
	}
	return true
}

func (d *Differ) mustWrap(i interface{}) wrapper.KubernetesObject {
	o, err := d.wrap(i)

	if err != nil {
		klog.Fatalf("Failed to wrap interface %v", err)
	}

	return o
}

func (d *Differ) createDiffObject(rawObj wrapper.KubernetesObject) map[string]interface{} {
	result := rawObj.GetRawObject()

	delete(result, "status")
	delete(result, "metadata")

	metadata := map[string]interface{}{}

	if d.labelConfig.Enable {
		metadata["labels"] = deleteKeys(rawObj.GetMetadata().Labels, d.labelConfig.IgnoreKeys)
	}

	if d.annotationConfig.Enable {
		metadata["annotations"] = deleteKeys(rawObj.GetMetadata().Annotations, d.annotationConfig.IgnoreKeys)
	}

	result["metadata"] = metadata

	return result
}

func deleteKeys(source map[string]string, target []string) map[string]string {
	for _, t := range target {
		delete(source, t)
	}
	return source
}
