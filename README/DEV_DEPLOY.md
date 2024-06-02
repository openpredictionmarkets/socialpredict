## Hypothetical Deploying from Dev

### Background

As of 31 May 2024, SocialPredict does not have a production version launched. That is to say, SocialPredict is designed for development, meaning out of the box, everything it comes with can be worked on within a local machine, but it's not set up to be served securely as a production instance on the web that can't be easily hacked.

That being said, while not recommended, it is possible to adapt the development version of SocialPredict to run in a production environment in its current state, for small and experimental deployments.

This document is geared at helping those who might want to convert the development version of socialpredict into a web deployment. While not all-encompassing and user caution is encouraged, this document can serve as a general guideline to get going.

### Contents

#### Setting Up Your Domain

First, you should purchase a domain, which you can do at any number of domain registrars. We are using Namecheap for this article.


### Set Up Digital Ocean Account

We are using Digital Ocean as a server, since it's predictable, easy and cheap. SocialPredict is designed to be performant, meaning it can perform well on a small server.

Setup Droplet with Docker
https://www.digitalocean.com/community/tutorials/how-to-use-the-docker-1-click-install-on-digitalocean

Getting Started After Deploying Docker
https://marketplace.digitalocean.com/apps/docker


### Logging In

If you use ssh-keygen to start up a new key file, then do:

```
ssh -i ~/.ssh/id_ed25519-do root@IPADDRESS
```

### Update and Upgrade

```
sudo apt update && sudo apt upgrade
```

### Reboot

```
sudo reboot
```

### Check Docker and Run Hello World

```
sudo systemctl status docker        # Check Docker service status
sudo docker run hello-world         # Run a test Docker container
```

### Verifying Docker


https://marketplace.digitalocean.com/apps/docker?ipAddress=143.198.177.112#getting-started

root@breirfoxforecast-alpha:~# docker version
Client: Docker Engine - Community
 Version:           26.1.3
 API version:       1.45
 Go version:        go1.21.10
 Git commit:        b72abbb
 Built:             Thu May 16 08:33:29 2024
 OS/Arch:           linux/amd64
 Context:           default

Server: Docker Engine - Community
 Engine:
  Version:          26.1.3
  API version:      1.45 (minimum version 1.24)
  Go version:       go1.21.10
  Git commit:       8e96db1
  Built:            Thu May 16 08:33:29 2024
  OS/Arch:          linux/amd64
  Experimental:     false
 containerd:
  Version:          1.6.32
  GitCommit:        8b3b7ca2e5ce38e8f31a34f35b2b68ceb8470d89
 runc:
  Version:          1.1.12
  GitCommit:        v1.1.12-0-g51d5e94
 docker-init:
  Version:          0.19.0
  GitCommit:        de40ad0
root@breirfoxforecast-alpha:~# docker-compose version
Command 'docker-compose' not found, but can be installed with:
snap install docker          # version 24.0.5, or
apt  install docker-compose  # version 1.29.2-1
See 'snap info docker' for additional versions.
root@breirfoxforecast-alpha:~# docker compose version
Docker Compose version v2.17.2
root@breirfoxforecast-alpha:~# docker buildx version
github.com/docker/buildx v0.14.0 171fcbe


### Clone the Repo

```
root@breirfoxforecast-alpha:# cd /home
root@breirfoxforecast-alpha:/home# git version
git version 2.34.1
```


```
root@breirfoxforecast-alpha:/home# git clone https://github.com/openpredictionmarkets/socialpredict.git
```


### Set Up Environmental Variables for Ports




### Setting Up NGINX Production System



### Link Domain Directly to Digital Ocean

Once you have a domain name, reference your domain registrar's documentation for [connecting up your domain to a routing service such as aws Route53](https://www.namecheap.com/support/knowledgebase/article.aspx/10371/2208/how-do-i-link-my-domain-to-amazon-web-services/).


https://www.namecheap.com/support/knowledgebase/article.aspx/10375/2208/how-do-i-link-a-domain-to-my-digitalocean-account/

https://docs.digitalocean.com/products/networking/dns/getting-started/dns-registrars/


### Create a Record in The Control Panel

https://docs.digitalocean.com/products/networking/dns/how-to/manage-records/#create-update-and-delete-records-using-the-control-panel


### Create an A Record to Point to Droplet

https://docs.digitalocean.com/products/networking/dns/how-to/manage-records/#a-records

### Secure NGinx with SSL

https://www.digitalocean.com/community/tutorials/how-to-secure-nginx-with-let-s-encrypt-on-ubuntu-22-04


