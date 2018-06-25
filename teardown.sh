#!/bin/bash

echo "Remove CDN Services"
docker stack rm cdn
echo "Remove FaaS functions"
faas-cli rm -f stack.yml 
