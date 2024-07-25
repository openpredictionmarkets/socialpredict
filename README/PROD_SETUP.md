### About

The following is a general tutorial geared at getting [SocialPredict](https://github.com/openpredictionmarkets/socialpredict) set up on Digital Ocean. Other virtual machine servers may work similarly, this serves as an example and is designed for more novice users.

### Setting Up Your Domain

First, you should purchase a domain, which you can do at any number of domain registrars. We are using Namecheap for this article.


### Set Up Digital Ocean Account

We are using Digital Ocean as a server, since it's predictable, easy and cheap. SocialPredict is designed to be performant, meaning it can perform well on a small server.

After logging into Digital Ocean, you can use the following marketplace app to instantly spin up a Droplet with all of the necessary docker tools pre-installed.

https://marketplace.digitalocean.com/apps/docker

Here is a more detailed tutorial on the topic:

https://www.digitalocean.com/community/tutorials/how-to-use-the-docker-1-click-install-on-digitalocean


### Logging In via SSH

Generally you should follow this tutorial to get connected via ssh to your Docker-based Digital Ocean droplet.

https://docs.digitalocean.com/products/droplets/how-to/connect-with-ssh/

If you use ssh-keygen to start up a new key file,

```
ssh-keygen
id_ed25519-do
```

Adding the `-do` at the end allows concerns to be kept separate, e.g. if other keyfiles are created, they can have that suffix at the end to allow for different keys for different purposes.  After this has been created do the following to log in:

```
ssh -i ~/.ssh/id_ed25519-do root@{IP_ADDRESS}
```
Where `{IP_ADDRESS}` is your server ip address.

#### Recovery

If anything bad happens, the server can be re-built from scratch on Digital Ocean, if you sign up for backups. However after the server is rebuilt, there will be a, "this server has changed," message.

```
ssh-keygen -R {IP_ADDRESS}
```

If you accidentally delete your key on your local machine or over-write it with a new key, you have to log in through the console on Digital Ocean for that droplet and add it.

First, copy the new key on your local machine by doing the following and copying the key:

```
cat ~/.ssh/id_ed25519-do.pub
```

Then copy and paste that, and within the [Digital Ocean droplet console](https://docs.digitalocean.com/products/droplets/how-to/connect-with-console/), add it to your authoirzed_key file.

```
nano ~/.ssh/authorized_keys
```

### Update and Upgrade

After signing in following the instructions above, you should run the following;

```
sudo apt update && sudo apt upgrade
```

#### Reboot After Updating and Upgrading

```
sudo reboot
```

### Check Docker and Run Hello World

The following will help verify that docker is up and running.

```
sudo systemctl status docker        # Check Docker service status
sudo docker run hello-world         # Run a test Docker container
```

### Verifying Docker


https://marketplace.digitalocean.com/apps/docker?#getting-started

```
user@server:~# docker version
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
```

### Verifying Networking and Domain

To verify that the networking works and the domain is connected to Digital Ocean properly and serving the application as expected, a sample app can be installed first.

Digital Ocean provides a sample, "hello world," app [here](https://github.com/digitalocean/sample-dockerfile/tree/main).

Navigate to your `/home` directory or wherever this should be organized on your machine and clone the repo with:

```
git clone https://github.com/digitalocean/sample-dockerfile.git
```

To deploy this, after logging in via SSH, run the folowing, within the directory:

```
docker build .
```
This will build the docker image. After having done this, the image ID can be verified with:

```
user@breirfoxforecast-alpha:/home/sample-dockerfile# docker images
REPOSITORY   TAG       IMAGE ID       CREATED         SIZE
<none>       <none>    71dea4f9eb20   2 minutes ago   15.7MB
```
Having verified the image exists, launch the app with:

```
root@breirfoxforecast-alpha:/home/sample-dockerfile# docker run -ti -p 80:80 71dea4f9eb20
```
There will be an ASCII art image of a shark as well as `Server listening at :80 🚀`

Then, navigate to the `{IP_ADDRESS}` on your browser for the droplet and the following message should be displayed.

```
Hello! you've requested /
```

This means that an app has been successfully deployed and is serving on 80.

### Link Domain Directly to Digital Ocean

Once you have a domain name, reference your domain registrar's documentation for [connecting up your domain to a routing service such as aws Route53](https://www.namecheap.com/support/knowledgebase/article.aspx/10371/2208/how-do-i-link-my-domain-to-amazon-web-services/).


https://www.namecheap.com/support/knowledgebase/article.aspx/10375/2208/how-do-i-link-a-domain-to-my-digitalocean-account/

https://docs.digitalocean.com/products/networking/dns/getting-started/dns-registrars/

### Add a Domain

Once the Domain has been pointed to Digital Ocean, add a Domain within the [Digital Ocean Networking Dashboard](https://cloud.digitalocean.com/networking/domains).


### Create A Records in The Control Panel

Once this application has been set up, as mentioned above, create A records for `www` and `@` for your droplet. This can be found by navigating to https://cloud.digitalocean.com/networking/domains/`{yourdomain}`

https://docs.digitalocean.com/products/networking/dns/how-to/manage-records/#create-update-and-delete-records-using-the-control-panel


https://docs.digitalocean.com/products/networking/dns/how-to/manage-records/#a-records

Ensure that both `www` and `@` are created.  After all of this has been updated, it may take 48 hours to take effect.

* Status of your domain can be checked at https://www.whatsmydns.net/
* The `dig` tool on the local command line can be used to check a domain's status (for all domains):

```
dig www.yourdomain.com
dig yourdomain.com
```

### Create a Firewall

* Firewalls are ways to restrict port accesses to only a narrow number of ports, or one port for a given purpose.

https://docs.digitalocean.com/products/networking/firewalls/how-to/create/#create-a-firewall-using-the-control-panel

#### HTTP and HTTPS:

* Use port 80 for HTTP.
* HTTPS (HTTP Secure) is the secure version of HTTP, encrypted using TLS (SSL). Use port 443 for HTTPS.


### Clone the SocialPredict Repo

Now that the domain has been pointed in the right direction and is working on propogating, the SocialPredict repo could be downloaded and run in the meantime.

```
root@breirfoxforecast-alpha:# cd /home
root@breirfoxforecast-alpha:/home# git version
git version 2.34.1
```
then...

```
root@breirfoxforecast-alpha:/home# git clone https://github.com/openpredictionmarkets/socialpredict.git
```

### Running `./SocialPredict install`


### Running `./SocialPredict up`



### Prod Tools

#### Entering a Container

##### Database

* Log into the psql service directly:

```
docker exec -it -e PGPASSWORD=${POSTGRES_PASSWORD} socialpredict-postgres-container psql -U ${POSTGRES_USER} -d socialpredict_db
```

