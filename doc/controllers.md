# Controllers

midonet-kube-controllers executable contains several controllers.

You can choose which controllers to enable by the ENABLED_CONTROLLER
environment variable.  By default all controllers are enabled.

<pre>
K8S resources

+-----------+ <-----+    +-------------+
|Node       |       |    |nodeannotator|
+-----------+ +--+  +--> |controller   | +--------------+
                 |       +-------------+                |
+-----------+    |                                      |
|Pod        | +--+                                      v
+-----------+    |      +--------------------+
                 |      |pod controller      |      +-----------+
+-----------+    |      +--------------------+      |           |
|Service    | +--+      |node controller     |      |MidoNet API|
+-----------+    +----> +--------------------+ +-+  |           |
                 |      |service controller  |   |  +-----------+
+-----------+    |      +--------------------+   |
|Endpoints  | +--+      |endpoints controller|   |      ^
+-----------+           +--------------------+   |      |
                                                 |      |
                   +-----------------------------+      |
                   |                                    |
+-----------+ <----+                                    |
|Translation|           +------------------+            |
+-----------+ +-------> |pusher controller | +----------+
                        +------------------+
</pre>

## pod, node, service, endpoints

These controllers watch the corresponding Kubernetes resources
and create/update/delete Translation custom resources accordingly.

## pusher

This controller watches Translation custom resources and
create/update/delete MidoNet API resources accordingly.

## nodeannotator

This controller adds "midonet.org/host-id" annotation to Kubernetes
Node resources, by querying MidoNet API with the assumption that
MidoNet Host name and Kubernetes Node name on a node match.

This controller also adds "midonet.org/tunnel-zone-id" and
"midonet.org/tunnel-endpoint-ip" annotations.

The annotation is used by pod and node controllers.
