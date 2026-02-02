
# Runtime Configuration Synchronization via Controller Loop

The Runtime Boostrapper synchronises several resources between KCP and the Kyma runtimes. The resources are either required by some webhook features (e.g. pull-secret is needed for accessing a private container registry, `ClusterTrustBundle` is mandatory to interact with BTP backend services etc.)

The following described behavior represents an interim solution. A long‑term architecture is planned in which the controller loop will directly synchronize configuration data to the runtimes without relying on indirect signaling.


## Components

### Controller Loop

A custom Kubernetes controller responsible for observing changes in selected cluster resources on KCP and initiating downstream actions. The watched resources are copied to  Kyma runtimes. Each resource change has to be synchronized from KCP to Kyma runtimes.

### Watched Resources

The controller loop monitors the following Kubernetes objects:

| Resource Type | Purpose | 
| --- | --- | 
| **Pull Secret** | Provides authentication credentials for pulling container images form private registries. | 
| **ClusterTrustBundle** | Supplies trust anchors (e.g., CA certificates) required by runtimes which interact with the BTP backend services. | 
| **Webhook ConfigMap** | Contains configuration for the Runtime Bootstrapper webhook. | 


### Runtime Custom Resource (RuntimeCR)

A custom resource representing a managed runtime instance.

The infrastructure manager component reacts on `Runtime` CR labels to determine if a runtime requires a reconciliation.


## Current Behavior (Interim Solution)

### Change Detection

The controller loop continuously watches the pull‑secret, ClusterTrustBundle, and webhook ConfigMap.

Whenever one of these resources is created, updated, or deleted, the controller receives an event.


### Triggering Reconciliation via Labeling

Upon detecting a change, the controller loop performs the following steps:

1. Identifies all affected `Runtime` CR objects.
2. Applies or updates a specific label (e.g., [`operator.kyma-project.io/force-patch-reconciliation=true`](https://github.com/kyma-project/kyma-infrastructure-manager/blob/c1d2f48a9b446b3374528278b46ea9be23ff622a/pkg/reconciler/annotations_utils.go#L4C32-L4C83)) on each `Runtime` CR.
3. The infrastructure manager component observes this label change.
4. The infrastructure manager reconciles the corresponding runtimes to ensure they receive the updated configuration.

This mechanism uses the `Runtime` CR label as a signaling channel between the controller loop and the infrastructure manager.


### Rationale for the Interim Approach

The labeling strategy provides a lightweight and low‑risk integration path:

* No direct modification of runtime resources is required.
* Existing reconciliation logic in the infrastructure manager remains unchanged.
* The controller loop only signals intent rather than performing the full synchronization.

This allows incremental rollout and testing of the controller loop without impacting runtime stability.


## Long‑Term Architecture

In the future design, the controller loop will no longer rely on labeling `Runtime` CR objects to trigger reconciliation. Instead, it will directly propagate configuration changes to the runtimes without any Kyma Infrastructure involvement.

![Runtime Bootstrapper Architecture](./assets/new-arch-rt-boostrapper.drawio.svg)

### Planned Behavior

When a watched resource changes:

1. The controller loop will use a dedicated Custom Resource to manage the lifecycle of the Runtime Bootstrapper:
    1. A new created Custom Resource will install the Runtime Boostrapper on a Kyma runtime.
    2. Status applied to the Runtime Bootstrapper will be reflected in the Custom Resource status.
    3. Deletion of the Custom Resource will trigger undeploy the Runtime Bootstrapper from the Kyma runtime.
2. The controller loop retrieves the updated value (e.g., new pull secret, updated trust bundle, modified webhook configuration).
3. The controller loop writes the updated data directly into the runtime’s configuration or associated Kubernetes objects.
4. The Kyma infrastructure manager no longer needs to detect labels or perform reconciliation for these specific updates.

### Benefits of the Long‑Term Approach

* **Reduced latency:** Changes are applied immediately without waiting for external reconciliation loops.
* **Lower complexity:** Reduces complexity in KIM and removes the indirection of using labels as signals.
* **Improved consistency:** Ensures runtimes always reflect the latest cluster configuration.
* **Clearer separation of responsibilities:**
  * Controller loop handles configuration propagation.
  * Controller manages lifecycle of the Runtime Boostrapper
  * Kyma Infrastructure manager focuses on lifecycle management of SKR runtimes.
  * Kyma Infrastructure Manager being unaware about the Runtime Bootstrapper.



## Migration Considerations

Transitioning from the interim to the long‑term solution requires:

* Clear separation and decoupling between KIM and the Runtime Bootstrapper components.
* Existing code which resists in KIM can be refactored and moved to Runtime Bootstrapper which simplifies the KIM codebase and reduces the amount of reconciliations steps.
* Updating the controller loop to implement direct synchronization logic.
* Ensuring the infrastructure manager no longer depends on RuntimeCR label changes for these configuration types.
* Introducing versioning or compatibility checks to avoid partial updates.
* Validating that runtimes can safely accept live updates to trust bundles, pull secrets, or webhook configuration.



## Summary

The current system uses a Kubernetes controller loop to detect changes in key configuration resources and signals the infrastructure manager by labeling `Runtime` CR objects. This approach serves as a temporary mechanism to ensure runtimes are reconciled when configuration changes occur.

The long‑term solution will 

1. be responsible to deploy the Runtime Bootstrapper webhook on SKR runtimes
2. uses it's own Customer Resource for managing the Runtome Boostrapper lifecycle and status
3. eliminate the signaling step to KIM and instead allow the controller loop to directly propagate configuration updates to the runtimes, improving efficiency, reducing complexity, and strengthening consistency across the system

