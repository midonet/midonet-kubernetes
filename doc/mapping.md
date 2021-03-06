Resource mapping between Kubernetes and MidoNet
===============================================

| Kubernetes | MidoNet     |
|:-----------|:------------|
| Node       | Bridge      |
| Pod        | Bridge Port |
| Service    | Chain/Rules |
| Endpoint   | Chain/Rules |

<pre>
              +----------------+  dst X/32 gw Y port P
              | Cluster Router |  dst S    port P
              |                |
              |            P   |
              +---+--------+---+
                  |        |Z
                  |        |      S
                  |      +-+--+-------+-+ Bridge
+----+-------+----+-+         |       |  (with Chains/Rules to
     |       |                |       |   implement Services)
     |       |                |       |
     |       |                |       |
     |       |                |       |       MidoNet
 - - | - - - | - - - - - - - -|- - - -|- - - - - - - -
     |       |                |Y      |       Linux
 +---+--+  +-+-+          +---+--+  +-+-+
 |Node  |  |Pod|          |Node  |  |Pod| dst default gw Z
 |      |  |   |          |      |  |   |
 |      |  +---+          |      |  +---+
 |      |                 |      |
 +------+                 +-+----+
                            |X
                            |
                         +--+---------------+
</pre>

Global resources
----------------

For a given Kubernetes deployment, the controller (midonet-kube-controllers)
automatically creates the following MidoNet API objects.
They don't have particular Kubernetes counterparts.

- A tunnel zone
- A deployment global Router (we call this the cluster router)
- Chains and Rules shared among all Bridges.

Kubernetes Node
---------------

When the controller noticed a Kubernetes Node, it would create
the following MidoNet REST API objects.

- A Bridge
- A Bridge Port on the bridge
- A Router Port on the cluster router
- A Port Link to link the above two ports
- A local Route on the cluster router for the subnet (PodCIDR of the node)
- Another Bridge Port on the bridge for Node connectivity
- HostInterfacePort to bound the interface to the port
  (The interface itself is asynchronously created by midonet-kube-node.)
- MACPort and IPv4MACPair for the port

Besides, it would create MidoNet Route objects on the cluster router,
to every addresses on the Node, either ExternalIP or InternalIP.

Also, it adds the node to the MidoNet tunnel zone, using Node's
first InternalIP as the tunnel endpoint.
(This can be overridden with midonet.org/tunnel-endpoint-ip Node annotation.)

Kubernetes Pod
--------------

- A Bridge Port on the Node Bridge
- HostInterfacePort to bound the interface to the port
  (The interface itself is asynchronously created by midonet-kube-cni.)
- MACPort and IPv4MACPair for the port

Kubernetes Service
------------------

- Chains for the ServicePort ("Service Chains")
- In the global "SERVICES" Chain which is shared by all Node bridges:
	- Rules to redirect the service traffic to the above per-Service Chains

Kubernetes Endpoint
-------------------

- Chains for each endpoints in EndpointSubsets
- In the corresponding Service Chains:
	- Jump rules to the Endpoint Chain
	  (In future, we might implement probability match for this rule.
	  Right now, the rule which happened to be the first in the list
	  is always used.  [MNA-1264][MNA-1264])
- In the Endpoint Chain:
	- A Rule to SNAT if the source IP matches the Endpoint IP
	- A Rule to DNAT to the endpoint IP

The corresponding REV_SNAT and REV_DNAT are created as a part of
a startup process.  See "Global resources" section above.

[MNA-1264]: https://midonet.atlassian.net/browse/MNA-1264
