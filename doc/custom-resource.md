# Custom Resources

This document describes the usage of Kubernetes Custom Resources
in this integration.

## Translation Custom Resource

A Translation is a Kubernetes Custom Resource, which represents
an ordered list of backend resources for the correspoinding
Kubernetes resource.

<pre>
Pod -------- Translation
              owner: Pod
              finalizers:
                midonet.org/deleter
              resources:
                Port
                HostInterfacePort
</pre>

### Ownership

It uses the ownership mechanism to simplify the deletion handling.
The Kubernetes resource for which a Translation is created is the owner
of the Translation.
Every Translation have a single owner, except Translations for global
resources, which don't have any owner.

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

A controller can find those Translations by traversing Translations
owned by the Kubernetes resource.  When a Kubernetes resource is
updated in a way it deletes some of Translations, the contoller uses
the ownership info to delete stale Translations.

### Pusher

The "pusher" controller watches the changes in Translation resources and
reflects them to the backend. (MidoNet API)

This controller is the only entity in this integration to
request modifications of the backend resources.

### Finalizer

Translations are always created with "midonet.org/deleter" finalizer to
ensure the deletion of the corresponding resources on the backend.
When a Translation is deleted, it's the responsibility of the pusher
controller (see below section) to remove the finalizer after applying
the deletion to the backend.

### Limitations

To keep the pusher controller simple, there are a few assumptions about
how a Translation can be updated.  The same restrictions apply in case
of an upgrade of the controller with existing Translations.

* The set of backend resources in a Translation should not change.
  Adding a backend resource to a Translation is ok. Removing them is not.
* Unless a backend resource supports PUT in the backend, it should never be
  changed in the Translation.
* A backend resource should not belong to multiple Translations.

### Translation version

midonet-kube-controllers have an internal constant called
Translation version.
It has been introduced to allow an automatic upgrade of
midonet-kube-controllers which involves incompatible changes of
Translations. (See the above "Limitations" section)

Developers should avoid unnecessary version bumps.

midonet-kube-controllers treats Translations with a different version
stale and removes them.  It means, after an upgrade with the version bump,
all Translations and backend resources will be removed and re-created.
The process likely involves cluster network interruptions.  Depending on
the size of a deployment, it can take very long.
