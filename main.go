package main

import (
	"flag"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

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
	flag.DurationVar(&resyncPeriod, "resync", time.Minute*1, "Periodic interval in which to force resync objects.")
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
	err = config.LoadConfig(configFile, &cfg)
	if err != nil {
		klog.Fatalf("LoadConfig failed: %v", err)
	}
	klog.Infof("Loaded config: %+v\n", cfg)

	rm, err := restMapper(kubeClinetConfig)
	if err != nil {
		klog.Fatalf("Failed to get rest mapper: %v", err)
	}

	// build differs
	var wg sync.WaitGroup
	for _, cfgDiffer := range cfg.Differs {
		gvr, err := searchResource(rm, cfgDiffer.Resource)
		if err != nil {
			klog.Errorf("failed to find GroupVersionResouces: %v", err.Error())
			continue
		}

		informer := informerFactory.ForResource(gvr).Informer()
		if err != nil {
			klog.Fatalf("informerForName failed: %v", err)
		}

		output := differ.NewOutput(differ.JSON, logAdded, logDeleted)
		d := differ.NewDiffer(wrapper.WrapUnstructured, informer, output, cfg.CommonLabelConfig, cfg.CommonAnnotationConfig, cfgDiffer.MatchRegexp, cfgDiffer.IgnoreRegexp)

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

func restMapper(c *rest.Config) (meta.RESTMapper, error) {
	dc, err := discovery.NewDiscoveryClientForConfig(c)
	if err != nil {
		return nil, err
	}
	gr, err := restmapper.GetAPIGroupResources(dc)
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDiscoveryRESTMapper(gr)

	return mapper, nil
}

func searchResource(rm meta.RESTMapper, resource string) (schema.GroupVersionResource, error) {
	gvr, err := rm.ResourceFor(schema.GroupVersionResource{
		Group:    "",
		Version:  "",
		Resource: resource,
	})

	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	return gvr, nil
}
