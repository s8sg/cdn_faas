# nginx_faas
A PreConfigure stack that provides a overview of using FAAS in CDN

### Try it out
> 1. [Deploy `Openfaas`](http://docs.openfaas.com/deployment/docker-swarm/)    
    
> 2. [Install `faas-cli`](http://docs.openfaas.com/cli/install/)
     
> 3. Deploy
>```bash
>    $ ./deploy.sh
>```
    
> 4. Visit [`127.0.0.1:80`](http://127.0.0.1:80)    
>    upload a file (.png). Say `monster.png` 
     
> 5. Visit [`127.0.0.1:80/assets/monster.png`](http://127.0.0.1:80/assets/monster.png)    
>    the resized file will be available and cached for 10sec
    
> 5. Teardown
>```bash
>    $ ./teardown.sh
>```


### Concept

![https://github.com/s8sg/nginx_faas/blob/master/assets/nginx_faas.png](https://github.com/s8sg/nginx_faas/blob/master/assets/nginx_faas.png)
