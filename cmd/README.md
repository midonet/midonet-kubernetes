# Executables

This directory contains a few executables.
Users are not expected to kick them directly.

## midonet-kube-controllers

This command contains Kubernetes controllers for MidoNet integration.
Those controllers monitor relevant changes on the apiserver, and
translate and apply them to the MidoNet API.

<pre>
MIDONETKUBE_LOG_LEVEL=debug MIDONETKUBE_MIDONET_API=http://localhost:8181/midonet-api MIDONETKUBE_KUBECONFIG=~/.kube/config ./midonet-kube-controllers
</pre>

## midonet-kube-node

This command connects the node to the cluster network.

<pre>
MIDONETKUBE_CLUSTERCIDR=10.1.0.0/16 MIDONETKUBE_SERVICECIDR=10.96.0.0/12 MIDONETKUBE_KUBECONFIG=~/.kube/config MIDONETKUBE_NODENAME=k sudo -E ./midonet-kube-node
</pre>

## midonet-kube-cni

This command is a CNI plugin for MidoNet Kubernetes integration.
Note that it's very specific to this Kubernetes integration.  It isn't
expected to work for other container environments.

https://kubernetes.io/docs/concepts/cluster-administration/network-plugins/#cni
