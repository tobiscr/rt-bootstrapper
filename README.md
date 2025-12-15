[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/rt-bootstrapper)](https://api.reuse.software/info/github.com/kyma-project/rt-bootstrapper)
[![Go Report Card](https://goreportcard.com/badge/github.com/kyma-project/rt-bootstrapper)](https://goreportcard.com/report/github.com/kyma-project/rt-bootstrapper)
[![unit tests](https://badgers.space/github/checks/kyma-project/rt-bootstrapper/main/unit-tests)](https://github.com/kyma-project/rt-bootstrapper/actions/workflows/unit-tests.yaml)
[![Coverage Status](https://coveralls.io/repos/github/kyma-project/rt-bootstrapper/badge.svg?branch=main)](https://coveralls.io/github/kyma-project/rt-bootstrapper?branch=main)
[![golangci lint](https://badgers.space/github/checks/kyma-project/rt-bootstrapper/main/golangci-lint)](https://github.com/kyma-project/rt-bootstrapper/actions/workflows/lint.yaml)
[![latest release](https://badgers.space/github/release/kyma-project/rt-bootstrapper)](https://github.com/kyma-project/rt-bootstrapper/releases/latest)

# RT Bootstrapper

This repository contains the source code for the `RT-bootstrapper` Kyma component used to configure Kyma Runtime components on restricted markets infrastructure.

## Overview

The `RT-bootstrapper` component contains two functional parts:

- Kubernetes admission webhook that intercepts the creation of pods and namespaces labeled for restricted markets. 
  It modifies the pod specifications to include necessary configurations, modify image paths to use configured remote registry, and provides pull secrets with credentials.

- Kubernetes Controller that watches for namespaces labeled for restricted markets and ensures that the required credentials secrets are present and synchronised in those namespaces.

For information how to use and configure `RT-bootstrapper`, see the [user documentation](./docs/README.md).


**Note:**
> This component is implemented as part of Kyma Runtime delivery for restricted markets.  
> Installing `RT-bootstrapper` on standard BTP managed Kyma Runtime or in self-managed Kyma Runtime cluster, may cause harm to your workloads.

## Prerequisites

- A managed Kyma Runtime instance running on BTP platform.
- Access to Kyma Runtime cluster with kubeconfig.

## Installation with Kyma Control Plane

In restricted market environment, the `RT-bootstrapper` is installed and configured automatically by Kyma Control Plane on all provisioned Kyma Runtimes.

## Installation with kubectl

Enable the `RT-bootstrapper` component in your Kyma cluster with kubectl by applying a release manifest.  

```bash
kubectl apply -f https://github.com/kyma-project/re-bootstrapper/releases/latest/download/rt-bootstrapper.yaml
```
## Architectural Decisions

Several architectural decision were made within the Kyma architecture meeting and also during the implementation phase (primarily caused by technical constraints which were not visible from the beginning and their solution required an in-time decision).

### Manipulation is limited to pods
The webhook will only manipulate pod resources. Other resources (e.g. StatefulSets, DaemonSets, Deployments etc.) will be ignored. This is required to avoid conflicts between the KLM and KIM: KLM is regularly processing the resources it was deploying (e.g. deployments of operators). If the webhook would modify these deployments, the KLM would revert the modifications regularly and both processes would "fight" against each other. To avoid such a situation, we agreed that KLM will never deploy pods, but high-level resources like deployments, DaemonSets, StatefulSets etc. The drawback of this decision is that the deployed pod can include different vlaues compared to its definition within a Deployment, StatefulSet, DeamonSet etc. which can be confusing for engineers/developers who are reviewing a pod definition in K8s and are not aware about the webhook existence and his adjustments.

### Non-blocking webhook
The admission webhook must be configured as non-blocking processing step for API-server requests. This means, the API server will continue the request processing when the webhook couldn't be invoked. This decision ensures that the API server will not stop the requests processing when the webhook is (temporarily) not available. The decision introduces the risk that pods get scheduled without being manipulated.

### Detection of non-manipulated resources is not part of the webhook
We agreed that the webhook is exclusively responsible for manipulating the manifest of pods during their creation phase. If a pod gets scheduled without being processed by the webhook (e.g. webhook was temporarily down), the pod could miss critical adjustments and in worst case not come up. To overcome this scenario, a house-keeping process (outside of the webhook) will be implemented which regularly scans all pods for missing manipulations. Such pods will be restarted by the housekeeping process (during the re-creation, the webhook will be invoked and the manipulations are getting applied).

### All pods will be scanned by the webhook
We agreed that all pods will be processed and manipulated by the webhook. This includes Kyma and customer workloads. It's also not required to annotate / label the pod or namespace to activate this behaviour. But customer have the option to disable this mechanism by annotations (see next point).

### Webhook configuration
The webhook will retrieve a default configuration (provided by KIM) which defines the amount of manipulations it has to apply for each pod. This configuration cannot be modified by workloads or customers. But it will be possible to disable the webhook manipulations ("Opt-Out" approach) by setting an annotation either on a pod (to disable the manipulation for this particular pod) or on a namespace (this disables the manipulation of all pods within the namespace). The webhook decides whether a manipulation will be applied in following precedence (first finding wins):

   1. Check for annotation on pod level
   2. Check  annotation on namespace (disables the manipulation of all pods within this namespace)
   3. Default configuration

### Applied manipulations
The webhook supports multiple manipulations. Which manipulation will be used can be configured by its default configuration (which is managed by KIM). Follow manipulations are supported:
   * Image registry adjustment: render the image name to inject a different container-registry. Private registries are supported and, if needed, an image-pull secret will also be injected into the pod manifest.
   * CA Bundle volume: Mounting the CA Bundle in a pod (needed for NS2 primarily)

### Pull-Secret Synchronisation
Private container registries require a pull secret.  KIM will ensure the latest pull-secret will become available within the `kyma-system` namespace. The name of the pull-secret will be static and not change over time, so that other compoents can use the name as unique identifier. This pull-secret has to be replicated into all namespaces (needed because Kyma has workloads which can be deployed in any namespace like Istio-sidecar or serverless and pull-secrets are namespace scoped). The webhook will include a dedicated controller which ensures that the secret will be available and synchronised in all namespaces.
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

2. Create a new k3d cluster and run `RT-bootstrapper` from the main branch:

    ```bash
    k3d cluster create test-cluster
    make deploy
    ```
## Usage

To use the `RT-bootstrapper`, you need to label your Kubernetes namespaces and pods accordingly.   
The admission webhook will intercept the creation of these resources and apply the necessary configurations.    
For more information, see the [user documentation](./docs/user/README.md).

## Contributing

<!--- mandatory section - do not change this! --->

See the [Contributing Rules](CONTRIBUTING.md).

## Code of Conduct
<!--- mandatory section - do not change this! --->

See the [Code of Conduct](CODE_OF_CONDUCT.md) document.

## License

<!--- mandatory section - do not change this! --->

See the [license](./LICENSE) file.
