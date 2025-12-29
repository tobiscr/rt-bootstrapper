[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/rt-bootstrapper)](https://api.reuse.software/info/github.com/kyma-project/rt-bootstrapper)
[![Go Report Card](https://goreportcard.com/badge/github.com/kyma-project/rt-bootstrapper)](https://goreportcard.com/report/github.com/kyma-project/rt-bootstrapper)
[![unit tests](https://badgers.space/github/checks/kyma-project/rt-bootstrapper/main/unit-tests)](https://github.com/kyma-project/rt-bootstrapper/actions/workflows/unit-tests.yaml)
[![Coverage Status](https://coveralls.io/repos/github/kyma-project/rt-bootstrapper/badge.svg?branch=main)](https://coveralls.io/github/kyma-project/rt-bootstrapper?branch=main)
[![golangci lint](https://badgers.space/github/checks/kyma-project/rt-bootstrapper/main/golangci-lint)](https://github.com/kyma-project/rt-bootstrapper/actions/workflows/lint.yaml)
[![latest release](https://badgers.space/github/release/kyma-project/rt-bootstrapper)](https://github.com/kyma-project/rt-bootstrapper/releases/latest)

# RT Bootstrapper

This repository contains the source code for the RT Bootstrapper Kyma component used to configure Kyma runtime components running in markets with individual infrastructure setups.

## Overview

RT Bootstrapper contains two functional parts:

- Kubernetes admission webhook that intercepts the creation of Pods.
  It modifies the Pod specifications to include necessary configurations, modifies image paths to use the configured remote registry, and provides pull secrets with credentials.

- Kubernetes Controller that watches for namespaces and ensures that the secrets with required credentials are present and synchronized in those namespaces.



> [!NOTE]
> This component is implemented as part of the Kyma runtime delivery.  
> Installing RT Bootstrapper in SAP BTP, Kyma runtime, or in a self-managed Kyma runtime cluster may negatively impact your workloads.

## Installation

### Prerequisites

- SAP BTP, Kyma runtime instance
- Access to the Kyma runtime cluster with kubeconfig

### Installation with Kyma Control Plane

In environments with individual infrastructure setups, RT Bootstrapper is installed and configured automatically by Kyma Control Plane in all provisioned Kyma runtimes.

### Installation with kubectl

To enable RT Bootstrapper in your Kyma cluster, apply the release manifest using kubectl:  

```bash
kubectl apply -f https://github.com/kyma-project/rt-bootstrapper/releases/latest/download/rt-bootstrapper.yaml
```
## Architectural Decisions

Several architectural decisions were made during the Kyma architecture meeting and the implementation phase. These decisions were primarily driven by technical constraints and the need for timely solutions.

### Manipulation is Limited to Pods
The webhook only manipulates Pod resources. Other resources, such as StatefulSets, DaemonSets, and Deployments, are ignored. This is required to avoid conflicts between Kyma Lifecycle Manager (KLM) and Kyma Infrastructure Manager (KIM). KLM regularly processes the resources it deployed (for example, Deployments of operators). If the webhook were to modify these deployments, the KLM would revert the modifications regularly, and both processes would "fight" against each other. To avoid such a situation, we agreed that KLM will never deploy Pods, but high-level resources like Deployments, DaemonSets, StatefulSets, etc. The drawback of this decision is that the deployed Pod can include different values compared to its definition within a Deployment, StatefulSet, DaemonSet, etc., which may be confusing for engineers or developers reviewing a Pod definition in Kubernetes who are unaware of the webhook's existence and its adjustments.

### Non-Blocking Webhook
The admission webhook must be configured as a non-blocking processing step for API-server requests. This means that the API server continues processing the request when the webhook cannot be invoked. This decision ensures that the API server continues to process requests even when the webhook is temporarily unavailable. The decision introduces the risk that Pods get scheduled without being manipulated.

### Detection of Non-Manipulated Resources Is Not Part of the Webhook
We agreed that the webhook is exclusively responsible for manipulating the manifest of Pods during their creation phase. If a Pod gets scheduled without being processed by the webhook (for example, when the webhook is temporarily down), the Pod might miss critical adjustments and, in the worst case, may not start up properly. To address this issue, a housekeeping process implemented outside of the webhook regularly scans all Pods for any missing manipulations. If such Pods are identified, the housekeeping process restarts them (during the re-creation, the webhook is invoked, and the manipulations are applied).

### The Webhook Scans All Pods
We agreed that all Pods are processed and manipulated by the webhook, including both Kyma and customer workloads. It's not required to annotate or label a Pod or a namespace to activate this behaviour. However, customers have the option to disable this mechanism using annotations (see [Webhook Configuration](#weebhook-configuration).

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
Private container registries require a pull secret. KIM ensures that the latest pull secret becomes available within the `kyma-system` namespace. The name of the pull secret is static and does not change over time, allowing other components to use it as a unique identifier. This pull secret must be replicated across all namespaces. This is required because Kyma workloads, such as Istio sidecars or serverless, can be deployed in any namespace, and pull secrets are namespace-scoped. The webhook includes a dedicated controller that ensures the secret is available and synchronized in all namespaces.
## Development

### Prerequisites

- Access to a Kubernetes cluster
- [Go](https://go.dev/)
- [k3d](https://k3d.io/)
- [Docker](https://www.docker.com/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [Kubebuilder](https://book.kubebuilder.io/)
- [yq](https://mikefarah.gitbook.io/yq)

### Installation in the k3d Cluster Using Make Targets

1. Clone the project.

    ```bash
    git clone https://github.com/kyma-project/rt-boostrapper.git && cd rt-boostrapper/
    ```

2. Create a new k3d cluster and run RT Bootstrapper from the main branch:

    ```bash
    k3d cluster create test-cluster
    make deploy
    ```
## Usage

To use RT Bootstrapper, label your Kubernetes namespaces and Pods accordingly.   
The admission webhook intercepts the creation of these resources and applies the necessary configurations.

## Contributing

<!--- mandatory section - do not change this! --->

See the [Contributing Rules](CONTRIBUTING.md).

## Code of Conduct
<!--- mandatory section - do not change this! --->

See the [Code of Conduct](CODE_OF_CONDUCT.md) document.

## License

<!--- mandatory section - do not change this! --->

See the [license](./LICENSE) file.
