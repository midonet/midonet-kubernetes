# Well-Known Labels, Annotations, etc

## Labels

| Label                 | Resource    | Description                         |
|:----------------------|:------------|:------------------------------------|
| midonet.org/owner-uid | Translation | UID of the k8s resource to which this Translation belongs |
| midonet.org/global    | Translation | Translations not owned by k8s resources |

## Annotations

| Annotation                     | Resource    | Description                         |
|:-------------------------------|:------------|:------------------------------------|
| midonet.org/host-id            | Node        | The corresponding MidoNet Host ID   |
| midonet.org/tunnel-zone-id     | Node        | The MidoNet Tunnel Zone to add this Node (An empty string means the default Tunnel Zone) |
| midonet.org/tunnel-endpoint-ip | Node        | The MidoNet tunnel endpoint IP for this Node |
| midonet.org/mac-address        | Pod, Node   | The MAC address for the pod/node    |

## Finalizers

| Finalizer             | Resource    | Description                         |
|:----------------------|:------------|:------------------------------------|
| midonet.org/deleter   | Translation | Postpone deletion for MidoNet API sync |
