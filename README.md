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
