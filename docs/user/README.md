
# Runtime Bootstrapper


## Overview
Some Kyma landscapes required individual infrastructure setups (e.g. if the landscape runs in a restricted context or market).  This can be a private container registry,  certificate based access mechanisms or other special configurations which let them differ from common Kyma landscapes.

Kyma modules are usually not aware about different landscapes and, without any adjustments, would not fully work in different landscape setups.

To solve this problem, the Runtime Bootstrapper is responsible to apply landscape specific configurations to Kyma modules, respectively to the workloads installed by the modules.

The Runtime Boostrapper is implemented as manipulated webhook which intercepts requests which create or update Pods. It extends or rewrites parts of their Pod manifests to make them compatible with the current landscape.

*Only Pods are intercepted by the webhook! Other resources (like `Deployment`, `DeamonSet` or `StatefulSet` etc.) are ignored.*

## Pod Manipulations

The Runtime Bootstrapper modifies a Pod only if one of the following conditions is met:

1. The Pod runs within a namespace which is listed in the Webhook's default configuration. All pods in these namespaces are automatically intercepted and modified. This option is usually only used for namespaces which are Kyma managed (e.g. `kyma-system`, `istio-system` etc.)
2. The namespace contains an annotation which indicates that all Pods within the namespace should be intercepted
2. The Pod itself is annotated to get intercepted by the webhook

### Applied Manipulations

The following table gives an overview of the different manipulations supported by the Runtime Bootstrapper.

THe column `Opt-In` contains the annotation which has to be added to an `Namespace` or `Pod` manifest to enable the manipulation for it (only required if the Pod is running in a Namespace which is not already per default monitored by the webhook).

|Name|Purpose|Applied Manipulation|Opt-In Annotation|
|--|--|--|--|
|Container Registry Rewrite|The webhook configuration contains a map of container registry hosts which have to be replaced by another host (e.g. if a private container registry should be used). |Rewrite container-registry host in `image` field.|`rt-cfg.kyma-project.io/alter-img-registry: "true"`<br/>|
|Image Pull Secret Injection|If the registry requires user credentials, the webhook will make sure that the secret-resource exists in the namespace and add a pull-secret entry to the manifest.|Add secret-reference to the `imagePullSecrets` field if registry requires credentials.|`rt-cfg.kyma-project.io/add-img-pull-secret: "true"`|
|FIPS mode enablement|Set an environment to the Pod which indicates to run in FIPS mode.|Add environment variable `KYMA_FIPS_MODE_ENABLED`|`rt-cfg.kyma-project.io/set-fips-mode: "true"`|
|Mount cluster trust bundle volume|Selected landscapes require a certificate (stored as `ClusterTrustBundle`) to interact with BTP backend services. The `ClusterTrustBundle` will be mounted as projected volume into the containers (includes also init-containers).|Mount a projected `volume` from `ClusterTrustBundle` to each container in the pod.|`rt-cfg.kyma-project.io/add-cluster-trust-bundle: "true"`|

*Note: if a Pod was manpulated by the webhook, the pod is annotated with `rt-bootstrapper.kyma-project.io/defaulted: "true"`*


## High Level Flow

![High Level Flow](./assets/flow.png)

## Useful Links (Optional)
* [Architectural decision](../contributor/architectural-decisions.md)
