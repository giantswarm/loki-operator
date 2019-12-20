[![CircleCI](https://circleci.com/gh/giantswarm/loki-operator.svg?&style=shield)](https://circleci.com/gh/giantswarm/loki-operator) [![Docker Repository on Quay](https://quay.io/repository/giantswarm/loki-operator/status "Docker Repository on Quay")](https://quay.io/repository/giantswarm/loki-operator)

# Loki operator

This operator is expected to gather partial `promtail` configurations relevant to specific applications
and combine them into the single promtail config file in its ConfigMap.

*Warning: This is proof-of-concept quality software! Don't run in production until you really YOLO!*

## How it works

Each application can deliver its own promtail log parsing config by creating a ConfigMap in the same namespace.
The name of the ConfigMap is arbitrary. The ConfigMap needs to have just 1 key `promtail.yaml`, which includes
promtail config to include in the actual promtail ConfigMap on behalf of this application.

The application informs `loki-operator` to register its config by including the Label in Pod's template yaml
(can be single Pod or any Pod created by Deployment or any other controller). The Label looks like this

```yaml
giantswarm.io/loki-promtail-config: apiserver-promtail-config
```

If there are more than 1 container in the Pod (like because you're running some kind of service mesh), you have to
explicitly point to the container you want to get logs from:

```yaml
giantswarm.io/loki-promtail-container: apiserver
```

## What's missing

- any tests
- almost all validation of input files
- currently the application needs to provide both the actual configuration pipeline but also filters config
  to apply the pipeline only to valid pods; this second part (where to apply the config - the filter) should
  be auto-generated
- restarts of promtail pods
- ability to get logs from many containers of the same Pod applying different configs to them.
