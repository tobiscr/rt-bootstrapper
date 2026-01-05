#!/bin/bash
set -e
set -o pipefail

docker_registry_secured="localhost:5001" # If you change this, also change in other test files
docker_registry_port_open="localhost:5002" # If you change this, also change it in registry-config.yml
cluster_name="registry-test"

export KUBECONFIG="$(k3d kubeconfig write ${cluster_name})"

check_annotation() {
  local label_selector=$1
  local annotation_key=$2
  local expected_value=$3

  annotations_output=$(kubectl annotate pod -l "$label_selector" --list)

  if echo "$annotations_output" | grep -q "$annotation_key=$expected_value"; then
    echo "Annotation $annotation_key has the expected value: $expected_value"
  else
    echo "Annotation check failed"
    exit 1
  fi
}

check_image() {
  local label_selector=$1
  local expected_image=$2

  actual_image=$(kubectl get pod -l "$label_selector" -o jsonpath="{.items[0].spec.containers[0].image}")

  if [ "$actual_image" == "$expected_image" ]; then
    echo "Image has the expected value: $expected_image"
  else
    echo "Image has unexpected value: $actual_image (expected: $expected_image)"
    exit 1
  fi
}

check_pull_secret() {
    local label_selector=$1
    local expected_pull_secret=$2

    actual_pull_secret=$(kubectl get pod -l "$label_selector" -o jsonpath="{.items[0].spec.imagePullSecrets[0].name}")
    if [ "$actual_pull_secret" == "$expected_pull_secret" ]; then
      if [ -z "$expected_pull_secret" ]; then
        echo "The rt-bootstrapper correctly omits ImagePullSecrets for this deployment"
        return
      fi
      echo "ImagePullSecrets has the expected value: $expected_pull_secret"
    else
      echo "ImagePullSecrets has unexpected value: $actual_pull_secret (expected: $expected_pull_secret)"
      exit 1
    fi

}

# Test 1:
# Acceptance criteria:
# - Image is pulled from the secure private registry
# - Image name is rewritten and Pod is in Ready state
# - Annotation rt-bootstrapper.kyma-project.io/defaulted is set to "true"
echo "Starting Test 1"
kubectl apply -f testdata/test-pod-1.yaml
kubectl wait --for=condition=Ready pod -l app=test-case1 --timeout=20s
check_annotation "app=test-case1" "rt-bootstrapper.kyma-project.io/defaulted" "true"
check_pull_secret "app=test-case1" "registry-credentials"
check_image "app=test-case1" ${docker_registry_secured}"/test-alpine-image:v1"
echo "Test 1 passed"; echo "========="

# Test 2:
# Acceptance criteria:
# - Image is pulled from the passwordless private registry
# - Image name is rewritten and Pod is in Running state
# - Annotation rt-cfg.kyma-project.io/add-img-pull-secret is set to "false"
# - Annotation rt-bootstrapper.kyma-project.io/defaulted is set to "true" by rt-bootstrapper controller
# - No imagePullSecrets are used
echo "Starting Test 2"
kubectl apply -f testdata/test-pod-2.yaml
kubectl wait --for=condition=Ready pod -l app=test-case2 --timeout=20s
check_annotation "app=test-case2" "rt-cfg.kyma-project.io/add-img-pull-secret" "false"
check_annotation "app=test-case2" "rt-bootstrapper.kyma-project.io/defaulted" "true"
check_pull_secret "app=test-case2" ""
check_image "app=test-case2" ${docker_registry_port_open}"/test-busybox:v1"
echo "Test 2 passed"; echo "========="

# Test 2:
# Acceptance criteria:
# - Image is pulled from the public Docker Hub registry

echo "All tests passed successfully"