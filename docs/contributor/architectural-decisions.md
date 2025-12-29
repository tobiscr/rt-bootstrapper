## Architectural Decisions

Several architectural decisions were made during the Kyma architecture meeting and the implementation phase. These decisions were primarily driven by technical constraints and the need for timely solutions.

### Manipulation is Limited to Pods
The webhook only manipulates Pod resources. Other resources, such as StatefulSets, DaemonSets, and Deployments, are ignored. This is required to avoid conflicts between Kyma Lifecycle Manager (KLM) and Kyma Infrastructure Manager (KIM). KLM regularly processes the resources it deployed (for example, Deployments of operators). If the webhook were to modify these deployments, the KLM would revert the modifications regularly, and both processes would "fight" against each other. To avoid such a situation, we agreed that KLM will never deploy Pods, but high-level resources like Deployments, DaemonSets, StatefulSets, etc. The drawback of this decision is that the deployed Pod can include different values compared to its definition within a Deployment, StatefulSet, DaemonSet, etc., which may be confusing for engineers or developers reviewing a Pod definition in Kubernetes who are unaware of the webhook's existence and its adjustments.

### Non-Blocking Webhook
The admission webhook must be configured as a non-blocking processing step for API-server requests. This means that the API server continues processing the request when the webhook cannot be invoked. This decision ensures that the API server continues to process requests even when the webhook is temporarily unavailable. The decision introduces the risk that Pods get scheduled without being manipulated.

### Detection of Non-Manipulated Resources Is Not Part of the Webhook
We agreed that the webhook is exclusively responsible for manipulating the manifest of Pods during their creation phase. If a Pod gets scheduled without being processed by the webhook (for example, when the webhook is temporarily down), the Pod might miss critical adjustments and, in the worst case, may not start up properly. To address this issue, a housekeeping process implemented outside of the webhook regularly scans all Pods for any missing manipulations. If such Pods are identified, the housekeeping process restarts them (during the re-creation, the webhook is invoked, and the manipulations are applied).

### The Webhook Scans All Pods
We agreed that all Pods are processed and manipulated by the webhook, including both Kyma and customer workloads. It's not required to annotate or label a Pod or a namespace to activate this behaviour. However, customers have the option to disable this mechanism using annotations (see [Webhook Configuration](#webhook-configuration)).

### Webhook Configuration
The webhook retrieves a default configuration (provided by KIM) that defines the number of manipulations it must apply to each Pod. Customers or other workloads cannot modify this configuration. However, it is possible to disable webhook manipulations and use the opt-out approach by setting an annotation on either a Pod or a namespace. Setting the annotation on a Pod disables the manipulation for this particular Pod, and setting the annotation on a namespace disables the manipulation of all Pods within the namespace. The webhook applies the manipulation using the following precedence, where the first finding takes priority:

1. Check for the annotation on the Pod level.
2. Check for the annotation on the namespace level.
3. Apply the default configuration.

### Applied Manipulations
The webhook supports multiple manipulations. The default configuration, managed by KIM, determines which manipulation is used. The following manipulations are supported:
* Image registry adjustment: Renders the image name to inject a different container registry. Private registries are supported, and if needed, an imagePullSecret is also injected into the Pod manifest.
* CA Bundle volume: Mounts the CA Bundle in a Pod (needed primarily for NS2).

### Pull Secret Synchronisation
Private container registries rtequire a pull secret. KIM ensures that the latest pull secret becomes available within the `kyma-system` namespace. The name of the pull secret is static and does not change over time, allowing other components to use it as a unique identifier. This pull secret must be replicated across all namespaces. This is required because Kyma workloads, such as Istio sidecars or serverless, can be deployed in any namespace, and pull secrets are namespace-scoped. The webhook includes a dedicated controller that ensures the secret is available and synchronized in all namespaces.
