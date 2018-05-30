# MidoNet Kubernetes Integration

## Overview

This software provides a way to use MidoNet as a backend for
Kubernetes networking. Namely, it provides the following Kubernetes
networking functionalitites.

* Basic cluster network
* Services (only ClusterIP type)

The [design doc][design] might have more details.

[design]: https://docs.google.com/document/d/1dYwz26I6NXO0MnbUf_pnC2Ihoz1Kdp0Pdm0DmEmGn4I/edit

## How to build

### Prequisite

- docker
- dep

### Build docker images

<pre>
	% dep ensure
	% docker build -f Dockerfile -t midonet/midonet-kube-controllers .
	% docker build -f Dockerfile-node -t midonet/midonet-kube-node .
</pre>

### Upload docker images

<pre>
	% TAG=1.1
	% docker tag midonet/midonet-kube-controllers midonet/midonet-kube-controllers:${TAG}
	% docker tag midonet/midonet-kube-node midonet/midonet-kube-node:${TAG}
	% docker push midonet/midonet-kube-controllers:${TAG}
	% docker push midonet/midonet-kube-node:${TAG}
</pre>

## How to deploy

The following procedure assumes [kubeadm][kubeadm] based deployment.
But something similar should apply to other deployment methods as well.

[kubeadm]: https://kubernetes.io/docs/setup/independent/create-cluster-kubeadm/

0. Deploy MidoNet as usual.

   * Every Kubernetes nodes including the master node should run MidoNet agent.
   * The following instruction assumes that MidoNet Host names and
     Kubernetes Node names are same for each nodes. It's usually the case
     because both of them are inferred from the hostname.

1. Create a MidoNet logical router.
   See the "Cluster router" section below.
   Record its UUID for the later use.
2. "kubeadm init" with Node IPAM enabled.
   (This integration relies on Node's spec.PodCIDR.)
<pre>
	% kubeadm init --pod-network-cidr=10.1.0.0/16
</pre>
3. Remove kube-proxy.
   (It isn't necessary or compatible with this integration.
   Unfortunately, kubeadm unconditionally sets it up.)
<pre>
	% kubectl -n kube-system delete ds kube-proxy
</pre>
4. After stopping kube-proxy, you might need to remove iptables rules
   installed by kube-proxy manually.
   Note: the following commands would remove many of relevant rules but
   leave some of rules and chains installed by kube-proxy. The simplest
   way to get a more clean state is to reboot the system.
<pre>
	% sudo iptables -t nat -F KUBE-SERVICES
	% sudo iptables -F KUBE-SERVICES
</pre>
5. Look at "manifests" directory in this repository.
   Copy and edit midonet-kube-config.template.yaml to match your deployment.
   Use the above mentioned MidoNet router UUID here.
   The modified file will be called midonet-kube-config.yaml hereafter.
6. Apply manifests.
<pre>
	% kubectl apply -f midonet-kube-crd.yaml
	% kubectl apply -f midonet-kube.yaml
</pre>
7. "Untaint" the master node if you want.
8. If you have workers, do "kubeadm join" as usual.

## Cluster router

This integration uses a deployment global MidoNet logical router.
We call it the cluster router.
A deployer should create it manually.

### External connectivity

The cluster router is used as the default gateway for every Pods
in the deployment. You can manually configure extra routes and ports
on the router to provide external connectivity to Pods.

## Limitations

* Only ClusterIP Service Type is implemented.
* Even if a Service has multiple Endpoints, only one endpoint which happens
  to be first is always used.  I.e. no load-balancing. [MNA-1264][MNA-1264]

[MNA-1264]: https://midonet.atlassian.net/browse/MNA-1264

## Contribution

### Submitting patches

We use [GerritHub][gerrithub] to submit patches.

[gerrithub]: https://review.gerrithub.io/#/q/project:midonet/midonet-kubernetes

We don't use GitHub pull requests.

### Issue tracker

Bugs and Tasks are tracked in [MidoNet jira][jira].
We might consider alternatives if the traffic goes up.

[jira]: https://midonet.atlassian.net/

We don't use GitHub issues.
