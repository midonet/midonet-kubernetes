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

For a given Kubernetes deployment, this controller automatically
creates the following MidoNet API objects.  They doesn't have
particular Kubernetes counterparts.

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

Besides, it would create MidoNet Route objects on the cluster router,
to every addresses on the Node, either ExternalIP or InternalIP,

Kubernetes Pod
--------------

- A Bridge Port on the Node Bridge
- HostInterfacePort to bound the interface to the port

Kubernetes Service
------------------

- Chains for the ServicePort
- In the global "SERVICES" Chain which is shared by all Node bridges:
	- Rules to redirect the service traffic to the above per-Service Chains

Kubernetes Endpoint
-------------------

- Chains for each endpoints in EndpointSubsets
- In the Service Chains:
	- Jump rules to the Endpoint Chain
	  (In future, we might implement probability match for this rule)
- In the Endpoint Chain:
	- A Rule to SNAT if the source IP matches the Endpoint IP
	- A Rule to DNAT to the endpoint IP

The corresponding REV_SNAT and REV_DNAT are created as a part of
a startup process.  See "Global resources" section above.
