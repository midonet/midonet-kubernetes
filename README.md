# MidoNet Kubernetes Integration

## Overview

This software provides a way to use [MidoNet][MidoNet] as a backend
for Kubernetes networking.  Namely, it provides the following
Kubernetes networking functionalitites.

* Basic cluster network, that is, connectivity among Pods, Nodes, and the apiserver
* Services with ClusterIP type (Note: externalIPs are ignored)

[MidoNet]: https://github.com/midonet/midonet

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
	% kubectl apply -f midonet-kube-config.yaml
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

3. Submit a patch to update docker image tags in our kubernetes manifests.

4. Review and merge the patch.
