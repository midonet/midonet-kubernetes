[![Build Status](https://travis-ci.org/midonet/midonet-kubernetes.svg?branch=master)](https://travis-ci.org/midonet/midonet-kubernetes)
[![Coverage Status](https://coveralls.io/repos/github/midonet/midonet-kubernetes/badge.svg?branch=master)](https://coveralls.io/github/midonet/midonet-kubernetes?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/midonet/midonet-kubernetes)](https://goreportcard.com/report/github.com/midonet/midonet-kubernetes)
[![GoDoc](https://godoc.org/github.com/midonet/midonet-kubernetes?status.svg)](https://godoc.org/github.com/midonet/midonet-kubernetes)

# MidoNet Kubernetes Integration

## Overview

This software provides a way to use [MidoNet][MidoNet] as a backend
for Kubernetes networking.  Namely, it provides the following
Kubernetes networking functionalitites.

* Basic cluster network, that is, connectivity among Pods, Nodes, and the apiserver
* Services with ClusterIP type (Note: externalIPs are ignored)

[MidoNet]: https://github.com/midonet/midonet

### Supported versions

This software is tested with:

* Kubernetes v1.10
* MidoNet 5.6

### Limitations

* Even if a Service has multiple Endpoints, only one endpoint which happens
  to be first is always used.  I.e. no load-balancing. [MNA-1264][MNA-1264]

[MNA-1264]: https://midonet.atlassian.net/browse/MNA-1264

### References

* The [doc][doc] directry contains internal documentations

* The [design doc][design] might have more details

[doc]: ./doc
[design]: https://docs.google.com/document/d/1dYwz26I6NXO0MnbUf_pnC2Ihoz1Kdp0Pdm0DmEmGn4I/edit

## How to build

### Prequisite

- docker

### Build docker images

<pre>
	% TAG=1.1
	% ./build.sh ${TAG}
</pre>

### Upload docker images

See the output of the above build.sh script.

### Create docker manifest lists

<pre>
	% TAG=1.1
	% ./createmanifest.sh ${TAG}
</pre>

Note: This uses ["docker manifest"][docker-manifest] command,
which is experimental.
You might need to enable the experimental feature on your docker
environment. And/or you might need to upgrade your docker.

[docker-manifest]: https://docs.docker.com/edge/engine/reference/commandline/manifest/

### Upload docker manifest lists

See the output of the above createmanifest.sh script.

## How to deploy

The following procedure assumes [kubeadm][kubeadm] based deployment.
But something similar should apply to other deployment methods as well.

[kubeadm]: https://kubernetes.io/docs/setup/independent/create-cluster-kubeadm/

0. Deploy MidoNet as usual.

   * Every Kubernetes nodes including the master node should run MidoNet agent.
   * The following instruction assumes that MidoNet Host names and
     Kubernetes Node names are same for each nodes. It's usually the case
     because both of them are inferred from the hostname.
   * No need to create a tunnel zone.  The integration will create its own
	 automatically.
   * Some of these assumptions can be overridden with manual [annotations][annotations].

1. "kubeadm init" with Node IPAM enabled.
   (This integration relies on Node's spec.PodCIDR.)
<pre>
	% kubeadm init --pod-network-cidr=10.1.0.0/16
</pre>
2. Remove kube-proxy.
   (It isn't necessary or compatible with this integration.
   Unfortunately, kubeadm unconditionally sets it up.
   cf. [kubeadm issue 776][kubeadm-776])
<pre>
	% kubectl -n kube-system delete ds kube-proxy
</pre>
3. After stopping kube-proxy, you might need to remove iptables rules
   installed by kube-proxy manually.
   Note: the following commands would remove many of relevant rules but
   leave some of rules and chains installed by kube-proxy. The simplest
   way to get a more clean state is to reboot the system.
<pre>
	% sudo iptables -t nat -F KUBE-SERVICES
	% sudo iptables -F KUBE-SERVICES
</pre>
4. Look at [manifests][manifests] directory in this repository.
   Copy and edit midonet-kube-config.template.yaml to match your deployment.
   The modified file will be called midonet-kube-config.yaml hereafter.
5. Apply manifests.
<pre>
	% kubectl apply -f midonet-kube-config.yaml
	% kubectl apply -f midonet-kube.yaml
</pre>
6. "Untaint" the master node if you want.
7. If you have workers, do "kubeadm join" as usual.

[annotations]: ./doc/labels-annotations.md#annotations
[kubeadm-776]: https://github.com/kubernetes/kubeadm/issues/776
[manifests]: ./manifests

## Trouble shooting

### Translation custom resources

midonet-kube-controllers use a CRD "Translation" to store the
intermediate objects converted from Kubernetes resources like
Pods and Services.
You can investigate them with kubectl command.
"tr" is a short name for the CRD.

#### Examples

- Get a list of Translation resources for a Service.
<pre>
% kubectl -n kube-system get tr -l midonet.org/owner-uid=$(kubectl -n kube-system get svc kube-dns -o jsonpath='{.metadata.uid}')
NAME                                                                                                   AGE
service-port.3.kube-system-kube-dns-dns-10.96.0.10-17-53-a3982c897cb53bcf4ec5dd4895d5e191e28628a5      9d
service-port.3.kube-system-kube-dns-dns-tcp-10.96.0.10-6-53-4b9da0761bde315827a2f3a49560b52b3c7a53a5   9d
service.3.kube-dns                                                                                     9d
</pre>

- The same for an Endpoints.
<pre>
% kubectl -n kube-system get tr -l midonet.org/owner-uid=$(kubectl -n kube-system get ep kube-dns -o jsonpath='{.metadata.uid}')
NAME                                                                                                      AGE
endpoints-port.3.kube-dns-dns-10.96.0.10-10.1.0.21-53-udp-54b27350879608cd686eb61cedb7273386a367c2        11s
endpoints-port.3.kube-dns-dns-10.96.0.10-10.1.0.22-53-udp-3adaf57ddc83531c51e037c7489c27ab6f671afa        10s
endpoints-port.3.kube-dns-dns-10.96.0.10-10.1.0.23-53-udp-1277c2fba4ef5db3a8a406311d1c1a534b09e171        10s
endpoints-port.3.kube-dns-dns-10.96.0.10-10.1.0.24-53-udp-9562d46816ee2c652d4d0f8a880489a13ec1e073        8s
endpoints-port.3.kube-dns-dns-10.96.0.10-10.1.1.145-53-udp-8c74d17e9a85139c771b0eb8321c000c88907c94       1h
endpoints-port.3.kube-dns-dns-10.96.0.10-10.1.1.177-53-udp-7afaaf3a2ba96d55bf2c9192a0cb4a902c8271f6       22s
endpoints-port.3.kube-dns-dns-tcp-10.96.0.10-10.1.0.21-53-tcp-6245126dfb5cd4878ccce4b468dd229e73ee7f6c    10s
endpoints-port.3.kube-dns-dns-tcp-10.96.0.10-10.1.0.22-53-tcp-1400a542aaf5e799316f80cb6daa0dc141f0902f    9s
endpoints-port.3.kube-dns-dns-tcp-10.96.0.10-10.1.0.23-53-tcp-7ebfb0755608d0cefb2939d92b18edd654fcb1ca    9s
endpoints-port.3.kube-dns-dns-tcp-10.96.0.10-10.1.0.24-53-tcp-f624318b925446dfa7016c1d2e1adfb4917f3fa9    12s
endpoints-port.3.kube-dns-dns-tcp-10.96.0.10-10.1.1.145-53-tcp-3302c499a8532147a976c74a96165bea98dc7034   1h
endpoints-port.3.kube-dns-dns-tcp-10.96.0.10-10.1.1.177-53-tcp-3ae2c8ecc12fb3c76ac73e6260b16bb35d8795ed   20s
</pre>

- Investigate one of those Translations.
<pre>
% kubectl -n kube-system describe tr endpoints-port.3.kube-dns-dns-10.96.0.10-10.1.0.24-53-udp-9562d46816ee2c652d4d0f8a880489a1
3ec1e073
Name:         endpoints-port.3.kube-dns-dns-10.96.0.10-10.1.0.24-53-udp-9562d46816ee2c652d4d0f8a880489a13ec1e073
Namespace:    kube-system
Labels:       midonet.org/owner-uid=a55ca29d-2330-11e8-8d8e-fa163ec6ef35
Annotations:  <none>
API Version:  midonet.org/v1
Kind:         Translation
Metadata:
  Cluster Name:
  Creation Timestamp:  2018-06-28T00:44:03Z
  Finalizers:
    midonet.org/deleter  Generation:  0
  Owner References:
    API Version:     v1
    Kind:            Endpoints    Name:            kube-dns
    UID:             a55ca29d-2330-11e8-8d8e-fa163ec6ef35
  Resource Version:  12040595
  Self Link:         /apis/midonet.org/v1/namespaces/kube-system/translations/endpoints-port.3.kube-dns-dns-10.96.0.10-10.1.0.24-53-udp-9562d46816ee2c652d4d0f8a880489a13ec1e073
  UID:               5adf0dce-7a6c-11e8-ba60-fa163ec6ef35
Resources:
  Body:    {"id":"5cfc9435-b3f4-5a79-8093-3c6a8c0f5fff","tenantId":"midonetkube","name":"KUBE-SEP-kube-dns/dns/10.96.0.10/10.1.0.24/53/UDP"}
  Kind:    Chain
  Parent:
  Body:    {"id":"e75a66ad-a7e6-569f-83bc-543d4a01107a","type":"jump","jumpChainId":"5cfc9435-b3f4-5a79-8093-3c6a8c0f5fff"}
  Kind:    Rule
  Parent:  2ba08097-dcb1-548e-817e-d259736a286c
  Body:    {"id":"0f0dbbff-40d9-537a-ae8f-3fdc533cb9fd","type":"dnat","flowAction":"accept","natTargets":[{"addressFrom":"10.1.1.177","addressTo":"10.1.1.177","portFrom":53,"portTo":53}]}
  Kind:    Rule
  Parent:  5cfc9435-b3f4-5a79-8093-3c6a8c0f5fff
  Body:    {"id":"2399be50-d5bd-5691-a6ff-8ffbc14df8a8","type":"snat","dlType":2048,"nwSrcAddress":"10.1.1.177","nwSrcLength":32,"flowAction":"continue","natTargets":[{"addressFrom":"10.96.0.10","addressTo":"10.96.0.10","portFrom":30000,"portTo":60000}]}
  Kind:    Rule
  Parent:  5cfc9435-b3f4-5a79-8093-3c6a8c0f5fff
Events:
  Type    Reason                   Age               From                      Message
  ----    ------                   ----              ----                      -------
  Normal  TranslationUpdatePushed  52s (x2 over 1m)  midonet-kube-controllers  Translation Update pushed to the backend
</pre>

### Prometheus

midonet-kube-controllers provides a few metrics for Prometheus.

<pre>
% curl -s http://localhost:9453/metrics|grep -E "^# (HELP|TYPE) midonet_"
# HELP midonet_kube_controllers_midonet_client_request_duration_seconds Latency of MidoNet API call
# TYPE midonet_kube_controllers_midonet_client_request_duration_seconds histogram
# HELP midonet_kube_controllers_midonet_client_requests_total Number of MidoNet API calls
# TYPE midonet_kube_controllers_midonet_client_requests_total counter
</pre>

#### Examples queries

See [Prometheus Querying documentation][prometheus-query] for details.

- Number of successful MidoNet API calls per seconds.
<pre>
sum(rate(midonet_kube_controllers_midonet_client_requests_total{code=~"2.*"}[5m])) by (resource,method)
</pre>

- Latency of successful MidoNet API calls.
<pre>
histogram_quantile(0.9, sum(rate(midonet_kube_controllers_midonet_client_request_duration_seconds_bucket{code=~"2.*"}[5m])) by (resource,method,le))
</pre>

[prometheus-query]: https://prometheus.io/docs/prometheus/latest/querying/basics/

### Go net/http/pprof

midonet-kube-controllers provides Go net/http/pprof on
the same port as the Prometheus metrics.

<pre>
% go tool pprof http://localhost:9453/debug/pprof/heap
</pre>

## Contribution

### Submitting patches

We use [GerritHub][gerrithub] to submit patches.

[gerrithub]: https://review.gerrithub.io/q/project:midonet%252Fmidonet-kubernetes

We don't use GitHub pull requests.

### Reviewing patches

Everyone is enouraged to review [patches for this repository][patches to review].

[patches to review]: https://review.gerrithub.io/q/project:midonet%252Fmidonet-kubernetes+status:open

If you want to be notified of patches, you can add this repository to
["Watched Projects"][watched projects] in your GerritHub settings.

[watched projects]: https://review.gerrithub.io/#/settings/projects

We have a voting CI named "Midokura Bot".
Unfortunately, its test logs are not publicly available.
If it voted -1 on your patch, please ask one of Midokura employees
to investigate the log.

### Merging patches

Unless it's urgent, a patch should be reviewed by at least one person
other than the submitter of the patch before being merged.

Right now, members of [GerritHub midonet group][midonet group] have the permission to merge patches.
If you are interested in being a member, please reach out the existing members.

[midonet group]: https://review.gerrithub.io/#/admin/groups/80,members

### Issue tracker

Bugs and Tasks are tracked in [MidoNet jira][jira].
We might consider alternatives if the traffic goes up.

[jira]: https://midonet.atlassian.net/

We don't use GitHub issues.

## Release process

Right now, our releases are tags on master branch.

1. Create and push a git tag for the release.

2. Build and push the docker images. (See the above sections about docker images)

3. Build and push the manifest list.

4. Submit a patch to update docker image tags in our kubernetes manifests.

6. Review and merge the patch.
