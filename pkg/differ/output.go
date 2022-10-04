package differ

import (
	"bytes"
	"encoding/json"
	"fmt"
	"k8s.io/klog/v2"
	"strings"
	"time"
)

// Output abstracts a straightforward way to write
type Output interface {
	WriteAdded(name string, namespace string, objectKind string)
	WriteDeleted(name string, namespace string, objectKind string)
	WriteUpdated(name string, namespace string, objectKind string, diffs []string)
}

// OutputFormat encodes
type OutputFormat int

const (
	// Text outputs the diffs in a simple text based format
	Text OutputFormat = iota
	// JSON outputs the diffs in json
	JSON
)

type output struct {
	format     OutputFormat
	logAdded   bool
	logDeleted bool
}

type jsonformat struct {
	Timestamp  string `json:"timestamp"`
	Verb       string `json:"verb"`
	ObjectKind string `json:"kind"`
	Notes      string `json:"notes"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
}

// NewOutput constructs a new outputter
func NewOutput(fmt OutputFormat, logAdded bool, logDeleted bool) Output {
	return &output{
		format:     fmt,
		logAdded:   logAdded,
		logDeleted: logDeleted,
	}
}

func (f *output) WriteAdded(name string, namespace string, objectKind string) {
	if !f.logAdded {
		return
	}

	f.write(name, namespace, "added", objectKind, nil)
}

func (f *output) WriteDeleted(name string, namespace string, objectKind string) {
	if !f.logDeleted {
		return
	}

	f.write(name, namespace, "deleted", objectKind, nil)
}

func (f *output) WriteUpdated(name string, namespace string, objectKind string, diffs []string) {
	f.write(name, namespace, "updated", objectKind, diffs)
}

func (f *output) write(name string, namespace string, verb string, objectKind string, etc []string) {
	diffString := strings.Join(etc, ", ")

	switch f.format {
	case Text:
		fmt.Printf("%s %s : %s %s (%s) %v\n", time.Now().UTC().Format(time.RFC3339), verb, namespace, name, objectKind, diffString)
	case JSON:
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(jsonformat{
			Timestamp:  time.Now().UTC().Format(time.RFC3339),
			Name:       name,
			Namespace:  namespace,
			Verb:       verb,
			ObjectKind: objectKind,
			Notes:      diffString,
		})

		if err != nil {
			klog.Errorf("Failed to convert to json: %v", err)
			return
		}

		fmt.Print(buf.String()) // 末尾の改行はjson.Encode()で追加されている
	}
}
