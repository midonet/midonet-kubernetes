# Custom Resources

This document describes the usage of Kubernetes Custom Resources
in this integration.

## Translation Custom Resource

A Translation is a Kubernetes Custom Resource, which represents
an ordered list of backend resources for the correspoinding
Kubernetes resource.

It uses the ownership mechanism to simplify the deletion handling.

It has a finanizer to ensure the deletion of the corresponding resources
on the backend.

<pre>
Pod -------- Translation
              owner: Pod
              finalizers:
                midonet.org/deleter
              resources:
                Port
                HostInterfacePort
</pre>

The Translations are mirrored to the backend by a controller.

The main purpose of having this indirection is to make deletions
of stale resources reliable without introducing the ownership tracking
mechanism in the backend.

### Multiple Translations

A controller can create multiple Translations for a Kubernetes resource.

<pre>
Service ----+------ Translation for servicePort 1
            |        owner: Service
            |
            +------ Translation for servicePort 2
            |        owner: Service
            |
            +------ Translation
                     owner: Service
</pre>

A Kubernetes resources can be updated in a way it deletes some of
Translations.
A controller can find those Translations by traversing Translations
owned by the Kubernetes resource.

### Pusher

The "pusher" controller watches the changes in Translation resources and
reflects them to the backend. (MidoNet API)

### Limitations

To keep the pusher controller simple, there are a few assumptions about
how a Translation can be updated.  The same restrictions apply in case
of an upgrade of the controller with existing Translations.

* The set of backend resources in a Translation should not change.
  Adding a backend resource to a Translation is ok. Removing them is not.
* Unless a backend resource supports PUT in the backend, it should never be
  changed in the Translation.
* A backend resource should not belong to multiple Translations.
