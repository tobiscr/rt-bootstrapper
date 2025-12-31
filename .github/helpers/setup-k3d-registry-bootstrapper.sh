#!/bin/bash
set -e
set -o pipefail

docker_registry_name="my-private-registry"
docker_registry_port="5001"
cluster_name="registry-test"
bootstrapper_image_name="localhost:5001/rt-bootstrapper:registry-test"

#Prepare a Docker registry

docker run --entrypoint htpasswd httpd:2.4.66-alpine -Bbn admin password123 > htpasswd

docker run -d \
--name "${docker_registry_name}" \
--restart=always \
-p "${docker_registry_port}":5000 \
-v "$(pwd)"/htpasswd:/auth/htpasswd \
-e "REGISTRY_AUTH=htpasswd" \
-e "REGISTRY_AUTH_HTPASSWD_REALM=Registry" \
-e "REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd" \
registry:3

docker login localhost:5001 -u admin -p password123

echo "Docker registry '${docker_registry_name}' is running on port ${docker_registry_port}."

docker pull alpine:latest
docker tag alpine:latest localhost:5001/test-alpine-image:v1
docker push localhost:5001/test-alpine-image:v1

#Preapre a k3d cluster and connect it to the registry
k3d cluster create ${cluster_name} \
  --registry-config "$(pwd)"/registry-config.yml
docker network connect k3d-${cluster_name} ${docker_registry_name}

#Build rt-bootstrapper image and push it to the private registry
make -C ../.. docker-build IMG=${bootstrapper_image_name}
make -C ../.. docker-push IMG=${bootstrapper_image_name}
make -C ../.. build-k3d-installer IMG=${bootstrapper_image_name}

#Deploy a rt-bootstrapper to k3d
export KUBECONFIG="$(k3d kubeconfig write ${cluster_name})"
kubectl create namespace kyma-system
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.18.2/cert-manager.yaml
kubectl wait --namespace cert-manager --for=condition=Available deployment --all --timeout=40s

kubectl create secret docker-registry registry-credentials -n kyma-system \
  --docker-server=localhost:5001 \
  --docker-username=admin \
  --docker-password=password123

kubectl apply -f ../../dist/k3d-install.yaml
kubectl patch deployment rt-bootstrapper-controller-manager -p '{"spec":{"template":{"spec":{"imagePullSecrets":[{"name":"registry-credentials"}]}}}}' -n kyma-system
kubectl wait --namespace kyma-system --for=condition=Available deployment/rt-bootstrapper-controller-manager --timeout=40s