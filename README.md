# Eventexporter

[![Build Status](https://travis-ci.org/sapcc/kubernetes-eventexporter.svg?branch=master)](https://travis-ci.org/sapcc/kubernetes-eventexporter)
[![Contributions](https://img.shields.io/badge/contributions-welcome-brightgreen.svg)](https://travis-ci.org/sapcc/kubernetes-eventexporter.svg?branch=master)
[![License](https://img.shields.io/badge/license-Apache-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0.txt)

----

Eventexporter filters events in a Kubernetes cluster by a custom definition and exposes them in a configurable metric.

Configuration example:

```yaml
metrics:
- name: metric_1
  event_matcher:
  - key: InvolvedObject.Kind
    expr: Pod
  - key: Message
    expr: .*Created container.*
  labels:
    node: Source.Host
- name: metric_2
  event_matcher:
  - key: Message
    expr: .*Created container.*
  labels:
    node: Source.Host
    type: Type
- name: metric_3
  event_matcher:
  - key: Type
    expr: Warning
  - key: Reason
    expr: PodOOMKilling
  labels:
    node: Source.Host
```

See [yaml/eventexporter.yaml](yaml/eventexporter.yaml) for an actual configuration and deployment of eventexporter.

## License
This project is licensed under the Apache2 License - see the [LICENSE](LICENSE) file for details
