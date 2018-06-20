#!/bin/bash

echo "Remove CDN Services"
docker service rm cdn_cache
echo "Remove FaaS functions"
faas-cli rm -f stack.yml 
