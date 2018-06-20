#!/bin/bash

# Check if docker is installed
if ! [ -x "$(command -v docker)" ]; then
          echo 'Unable to find docker command, please install Docker (https://www.docker.com/) and retry' >&2
            exit 1
fi

echo "Building FaaS for File handling"
#docker build -t s8sg/file-handler file_handler/
faas-cli build -f stack.yml

echo "Creating Docker Network func_functions if not exist"
[ ! "$(docker network ls | grep func_functions)" ] && docker network create -d overlay --attachable func_functions

echo "Deploying CDN Services"
# Deploy the docker stack with device name
docker stack deploy --compose-file docker-compose.yml cdn
echo "Service Stack successfully Deployed"


echo "Deploy File Handler Function as a FaaS"
faas-cli deploy -f stack.yml
