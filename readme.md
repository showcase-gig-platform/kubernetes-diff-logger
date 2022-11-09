# kubernetes-diff-logger

This simple application is designed to watch Kubernetes objects and log diffs when they occur.  It is designed to log changes to Kubernetes objects in a clean way for storage and processing in [Loki](https://github.com/grafana/loki/).

## Example output

```
{"timestamp":"2022-10-06T08:34:30Z","verb":"updated","kind":"Deployment","notes":".spec.replicas: 1 -> 2","name":"podinfo","namespace":"default"}
{"timestamp":"2022-10-06T08:35:47Z","verb":"updated","kind":"Deployment","notes":".spec.template.spec.containers[0].env[1]: map[name:NEW_ENV_KEY value:newEnvValue] (added)","name":"podinfo","namespace":"default"}
```

See [Deployment](./deployment) for example yaml to deploy to Kubernetes.  The example will monitor and log information about changes in all namespaces.

## Usage

```
Usage of ./kubernetes-diff-logger:
  -config string
    	Path to config file.  Required.
  -kubeconfig string
    	Path to a kubeconfig. Only required if out-of-cluster.
  -log-added
    	Log when deployments are added.
  -log-deleted
    	Log when deployments are deleted.
  -master string
    	The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.
  -namespace string
    	Filter updates by namespace.  Leave empty to watch all.
  -resync duration
    	Periodic interval in which to force resync objects. (default 30s)
```

## Config File

```yaml
differs:
  - resource: "deployment"
    matchRegex: "want-to-log"
    ignoreRegex: "want-to-(ignore|exclude)"
  - resource: "configmap"
  - resource: "mycustomresource"
commonLabelConfig:
  enable: true
  ignoreKeys:
    - untracked-label-key
commonAnnotationConfig:
  enable: true
  ignoreKeys:
    - untracked-annotation-key
```

| Field                             | Type     | Description                                                      |
|-----------------------------------|----------|------------------------------------------------------------------|
| differs.resource                  | string   | Name of Kubernetes resource type that you want to log diff.      |
| differs.matchRegexp               | string   | Regexp for resource name to log. (If blank, log all)             |
| differs.ignoreRegexp              | string   | Regexp for resource name to not log. (Priority over matchRegexp) |
| commonLabelConfig.enable          | boolean  | Whether to log metadata.labels diff.                             |
| commonLabelConfig.ignoreKeys      | []string | Key of labels that do not log diff. (exact match)                |
| commonAnnotationConfig.enable     | boolean  | Whether to log metadata.annotations diff.                        |
| commonAnnotationConfig.ignoreKeys | []string | Key of annotations that do not log diff. (exact match)           |
