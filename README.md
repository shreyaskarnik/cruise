# Heptio Cruise

Stress free HTTP monitoring for your Web Applications.

**Maintainers:** [Heptio][0]

[![Build Status][1]][2]

## Overview
Cruise automatically configures HTTP monitoring of Kubernetes Ingress resources.

## Prerequisites

Contour is tested with Kubernetes clusters running version 1.7 and later, but should work with earlier versions.

## Get started

You can try out Contour by creating a deployment from a hosted manifest -- no clone or local install necessary.

What you do need:

- A Kubernetes cluster that supports Service objects of `type: LoadBalancer` ([AWS Quickstart cluster][9] or Minikube, for example)
- `kubectl` configured with admin access to your cluster

See the [deployment documentation][10] for more deployment options if you don't meet these requirements.

### Add Cruise to your cluster

Run:

```
$ git clone https://github.com/heptio/cruise
$ kubectl apply -f cruise/deploy
```

[0]: https://github.com/heptio
[1]: https://travis-ci.org/heptio/cruise.svg?branch=master
[2]: https://travis-ci.org/heptio/cruise
[3]: /docs
[4]: https://github.com/heptio/cruise/issues
[5]: /CONTRIBUTING.md
[6]: https://github.com/heptio/cruise/releases
[8]: /CODE_OF_CONDUCT.md
[9]: https://aws.amazon.com/quickstart/architecture/heptio-kubernetes/
[11]: https://kubernetes.io/docs/concepts/services-networking/service/
[12]: https://kubernetes.io/docs/concepts/services-networking/ingress/
[14]: https://github.com/kubernetes-up-and-running/kuard
[16]: https://github.com/envoyproxy/envoy/issues/95
[18]: /FAQ.md
