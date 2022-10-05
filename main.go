package main

import (
	"flag"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/pkg/errors"
	"github.com/showcase-gig-platform/kubernetes-diff-logger/pkg/config"
	"github.com/showcase-gig-platform/kubernetes-diff-logger/pkg/differ"
	"github.com/showcase-gig-platform/kubernetes-diff-logger/pkg/signals"
	"github.com/showcase-gig-platform/kubernetes-diff-logger/pkg/wrapper"
)

var (
	masterURL    string
	kubeconfig   string
	resyncPeriod time.Duration
	namespace    string
	logAdded     bool
	logDeleted   bool
	configFile   string
)

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.DurationVar(&resyncPeriod, "resync", time.Second*30, "Periodic interval in which to force resync objects.")
	flag.StringVar(&namespace, "namespace", "", "Filter updates by namespace.  Leave empty to watch all.")
	flag.BoolVar(&logAdded, "log-added", false, "Log when deployments are added.")
	flag.BoolVar(&logDeleted, "log-deleted", false, "Log when deployments are deleted.")
	flag.StringVar(&configFile, "config", "", "Path to config file.  Required.")
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	// build k8s client
	kubeClinetConfig, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	client, err := dynamic.NewForConfig(kubeClinetConfig)
	if err != nil {
		klog.Fatalf("kubernetes.NewForConfig failed: %v", err)
	}

	// build shared informer
	var informerFactory dynamicinformer.DynamicSharedInformerFactory
	if namespace == "" {
		informerFactory = dynamicinformer.NewDynamicSharedInformerFactory(client, resyncPeriod)
	} else {
		informerFactory = dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, resyncPeriod, namespace, func(*metav1.ListOptions) {})
	}

	stopCh := signals.SetupSignalHandler()

	// load config
	var cfg config.Config
	err = loadConfig(configFile, &cfg)
	if err != nil {
		klog.Fatalf("loadConfig failed: %v", err)
	}

	// build differs
	var wg sync.WaitGroup
	for _, cfgDiffer := range cfg.Differs {
		gvk, err := searchResource(kubeClinetConfig, schema.GroupKind{
			Group: cfgDiffer.GroupKind.Group,
			Kind:  cfgDiffer.GroupKind.Kind,
		})
		if err != nil {
			klog.Errorf("failed to find GroupVersionResouces: %v", err.Error())
			continue
		}

		informer := informerFactory.ForResource(gvk).Informer()
		if err != nil {
			klog.Fatalf("informerForName failed: %v", err)
		}

		output := differ.NewOutput(differ.JSON, logAdded, logDeleted)
		d := differ.NewDiffer(cfgDiffer.NameFilter, wrapper.WrapUnstructured, informer, output, cfg.CommonLabelConfig, cfg.CommonAnnotationConfig)

		wg.Add(1)
		go func(differ *differ.Differ) {
			defer wg.Done()

			if err := d.Run(stopCh); err != nil {
				klog.Fatalf("Error running differ %v", err)
			}

		}(d)
	}

	informerFactory.Start(stopCh)
	wg.Wait()
}

func loadConfig(filename string, cfg *config.Config) error {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return errors.Wrap(err, "Error reading config file")
	}

	return yaml.UnmarshalStrict(buf, &cfg)
}

func searchResource(c *rest.Config, gk schema.GroupKind) (schema.GroupVersionResource, error) {
	dc, err := discovery.NewDiscoveryClientForConfig(c)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc))
	mapping, err := mapper.RESTMapping(gk, "")

	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	return mapping.Resource, nil
}
