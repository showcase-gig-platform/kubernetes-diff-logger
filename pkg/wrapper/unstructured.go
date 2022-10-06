package wrapper

import (
	"encoding/json"
	"fmt"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

type Unstructured struct {
	d *unstructured.Unstructured
}

func WrapUnstructured(i interface{}) (KubernetesObject, error) {
	d, ok := i.(*unstructured.Unstructured)

	if !ok {
		return nil, fmt.Errorf("expected Unstructured received %T", i)
	}

	return &Unstructured{
		d: d,
	}, nil
}

func (d *Unstructured) GetMetadata() v1meta.ObjectMeta {
	jmeta, err := json.Marshal(d.d.Object["metadata"])
	if err != nil {
		klog.Errorf("failed to parse metadata: %v", err)
		return v1meta.ObjectMeta{}
	}
	var meta v1meta.ObjectMeta
	if err := json.Unmarshal(jmeta, &meta); err != nil {
		klog.Errorf("failed to convert metadata to meta.ObjectMeta: %v", err)
		return v1meta.ObjectMeta{}
	}
	return meta
}

func (d *Unstructured) GetObjectSpec() interface{} {
	return d.d.Object["spec"]
}

func (d *Unstructured) GetKind() string {
	t := "Unstructured"
	if k, ok, _ := unstructured.NestedString(d.d.Object, "kind"); ok {
		t = k
	}
	return t
}

func (d *Unstructured) GetRawObject() map[string]interface{} {
	rawJson, err := json.Marshal(d.d.Object)
	if err != nil {
		klog.Errorf("failed to parse raw object: %v", err)
		return map[string]interface{}{}
	}
	var cp map[string]interface{}
	if err := json.Unmarshal(rawJson, &cp); err != nil {
		klog.Errorf("failed to convert raw object: %v", err)
		return map[string]interface{}{}
	}
	return cp
}
