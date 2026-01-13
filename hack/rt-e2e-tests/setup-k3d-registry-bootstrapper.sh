#!/bin/bash
set -e
set -o pipefail

docker_registry_name_secured="registry-test-secured"
docker_registry_port_secured="5001" # If you change this, also change it in registry-config.yml
docker_registry_name_open="registry-test-open"
docker_registry_port_open="5002" # If you change this, also change it in registry-config.yml
cluster_name="rt-bootstrapper-test-e2e"

#Prepare a Docker registry
docker run --entrypoint htpasswd httpd:2.4.66-alpine -Bbn admin password123 > htpasswd

docker run -d \
--name "${docker_registry_name_secured}" \
--restart=always \
-p "${docker_registry_port_secured}":5000 \
-v "$(pwd)"/htpasswd:/auth/htpasswd \
-e "REGISTRY_AUTH=htpasswd" \
-e "REGISTRY_AUTH_HTPASSWD_REALM=Registry" \
-e "REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd" \
registry:3

docker run -d \
--name "${docker_registry_name_open}" \
--restart=always \
-p "${docker_registry_port_open}":5000 \
registry:3

docker login localhost:${docker_registry_port_secured} -u admin -p password123
echo "Password secured Docker registry '${docker_registry_name_secured}' is running on port ${docker_registry_port_secured}."

if [ "$(curl -o /dev/null -s -w "%{http_code}\n" http://localhost:"${docker_registry_port_open}"/v2/_catalog)" -ne 200 ]; then
  echo "Failed to connect to open Docker registry '${docker_registry_name_open}' on port ${docker_registry_port_open}."
  exit 1
fi
echo "Open Docker registry '${docker_registry_name_open}' is running on port ${docker_registry_port_open}."

docker pull alpine:latest
docker tag alpine:latest localhost:${docker_registry_port_secured}/test-alpine-image:v1
docker push localhost:${docker_registry_port_secured}/test-alpine-image:v1

docker pull busybox:latest
docker tag busybox:latest localhost:${docker_registry_port_open}/test-busybox:v1
docker push localhost:${docker_registry_port_open}/test-busybox:v1

docker network connect k3d-${cluster_name} ${docker_registry_name_secured}
docker network connect k3d-${cluster_name} ${docker_registry_name_open}