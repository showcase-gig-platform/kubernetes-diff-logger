package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	memory "k8s.io/client-go/discovery/cached"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/grafana/kubernetes-diff-logger/pkg/differ"
	"github.com/grafana/kubernetes-diff-logger/pkg/signals"
	"github.com/grafana/kubernetes-diff-logger/pkg/wrapper"
	"github.com/pkg/errors"
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
	flag.Parse()

	// build k8s client
	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	// client, err := kubernetes.NewForConfig(config)
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatalf("kubernetes.NewForConfig failed: %v", err)
	}

	/*
		// build shared informer
		var informerFactory informers.SharedInformerFactory
		if namespace == "" {
			informerFactory = informers.NewSharedInformerFactory(client, resyncPeriod)
		} else {
			informerFactory = informers.NewSharedInformerFactoryWithOptions(client, resyncPeriod, informers.WithNamespace(namespace))
		}
	*/

	// build shared informer
	var informerFactory dynamicinformer.DynamicSharedInformerFactory
	if namespace == "" {
		informerFactory = dynamicinformer.NewDynamicSharedInformerFactory(client, resyncPeriod)
	} else {
		informerFactory = dynamicinformer.NewFilteredDynamicSharedInformerFactory(client, resyncPeriod, namespace, func(*metav1.ListOptions) {})
	}

	stopCh := signals.SetupSignalHandler()

	// load config
	cfg := DefaultConfig()
	err = loadConfig(configFile, &cfg)
	if err != nil {
		log.Fatalf("loadConfig failed: %v", err)
	}

	/*
		// build differs
		var wg sync.WaitGroup
		for _, cfgDiffer := range cfg.Differs {
			informer, wrap, err := informerForName(cfgDiffer.Type, informerFactory)
			if err != nil {
				log.Fatalf("informerForName failed: %v", err)
			}

			output := differ.NewOutput(differ.JSON, logAdded, logDeleted)
			d := differ.NewDiffer(cfgDiffer.NameFilter, wrap, informer, output)

			wg.Add(1)
			go func(differ *differ.Differ) {
				defer wg.Done()

				if err := d.Run(stopCh); err != nil {
					log.Fatalf("Error running differ %v", err)
				}

			}(d)
		}
	*/

	// build differs
	var wg sync.WaitGroup
	for _, gk := range cfg.GroupKinds {
		gvk, err := searchResource(config, schema.GroupKind{
			Group: gk.Group,
			Kind:  gk.Kind,
		})
		if err != nil {
			// todo: エラーログで出す
			log.Printf("failed to find GroupVersionResouces: %v", err.Error())
			continue
		}

		informer := informerFactory.ForResource(gvk).Informer()
		if err != nil {
			log.Fatalf("informerForName failed: %v", err)
		}

		output := differ.NewOutput(differ.JSON, logAdded, logDeleted)
		d := differ.NewDiffer(gk.NameFilter, wrapper.WrapUnstructured, informer, output)

		wg.Add(1)
		go func(differ *differ.Differ) {
			defer wg.Done()

			if err := d.Run(stopCh); err != nil {
				log.Fatalf("Error running differ %v", err)
			}

		}(d)
	}

	informerFactory.Start(stopCh)
	wg.Wait()
}

func informerForName(name string, i informers.SharedInformerFactory) (cache.SharedInformer, wrapper.Wrap, error) {

	switch name {
	case "deployment":
		return i.Apps().V1().Deployments().Informer(), wrapper.WrapDeployment, nil
	case "statefulset":
		return i.Apps().V1().StatefulSets().Informer(), wrapper.WrapStatefulSet, nil
	case "daemonset":
		return i.Apps().V1().DaemonSets().Informer(), wrapper.WrapDaemonSet, nil
	case "cronjob":
		return i.Batch().V1().CronJobs().Informer(), wrapper.WrapCronJob, nil
	}

	return nil, nil, fmt.Errorf("Unsupported informer name %s", name)
}

func loadConfig(filename string, cfg *Config) error {
	buf, err := ioutil.ReadFile(filename)
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
