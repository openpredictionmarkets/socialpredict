## Hypothetical Deploying from Dev

### Background

As of 31 May 2024, SocialPredict does not have a production version launched. That is to say, SocialPredict is designed for development, meaning out of the box, everything it comes with can be worked on within a local machine, but it's not set up to be served securely as a production instance on the web that can't be easily hacked.

That being said, while not recommended, it is possible to adapt the development version of SocialPredict to run in a production environment in its current state, for small and experimental deployments.

This document is geared at helping those who might want to convert the development version of socialpredict into a web deployment. While not all-encompassing and user caution is encouraged, this document can serve as a general guideline to get going.

### Contents

### Setting Up Your Domain

First, you should purchase a domain, which you can do at any number of domain registrars. We are using Namecheap for this article.


### Set Up Digital Ocean Account

We are using Digital Ocean as a server, since it's predictable, easy and cheap. SocialPredict is designed to be performant, meaning it can perform well on a small server.

Setup Droplet with Docker
https://www.digitalocean.com/community/tutorials/how-to-use-the-docker-1-click-install-on-digitalocean

Getting Started After Deploying Docker
https://marketplace.digitalocean.com/apps/docker


### Logging In via SSH

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

If anything bad happens, the server can be re-built from scratch on Digital Ocean. However after the server is rebuilt, there will be a, "this server has changed," message.

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

### Set Up Environmental Variables for Ports


* We have a script designed to inject environmental variables into your environment after having manually written them into a file.
* It is important to not ever leak environmental variable values into your socialpredict repo itself, to anywhere that has git, so that you can't accidentally push them to the internet somewhere.
* This is not the most secure, ideal practice, a better practice would be to use a secrets manager such as [Infisical](https://infisical.com/) or AWS Secrets Manager, but as long as you only manually change your environmental variable file locally on the Digital Ocean machine you're using for deployment, it should be fine.

#### Elements of Automatically Injecting Environmental Variables

* The env file should have the following values set:

```
BACKEND_IMAGE_NAME=socialpredict-backend-prod
BACKEND_CONTAINER_NAME=socialpredict-backend-container-prod
BACKEND_PORT=8080
BACKEND_HOSTPORT=8086
POSTGRES_CONTAINER_NAME=db-postgres-container-prod
DB_PORT=5433
DB_HOST=db
POSTGRES_PORT=5432
POSTGRES_USER={set custom value}
POSTGRES_PASSWORD={set custom value}
POSTGRES_DATABASE=proddb
FRONTEND_IMAGE_NAME=socialpredict-frontend-prod
FRONTEND_CONTAINER_NAME=socialpredict-frontend-container-prod
REACT_PORT=5173
REACT_HOSTPORT=5173
NGINX_IMAGE_NAME=socialpredict-nginx-prod
NGINX_CONTAINER_NAME=socialpredict-nginx-container-prod
NGINX_PORT=80
NGINX_HOSTPORT=80
SSLPORT=443
ADMIN_PASSWORD={set custom value}
```

* Then an executable env-writer file could look like this, and be activated with `./env-writer` to push all of the environment variables into the environment.

```
#!/bin/bash

echo "Make sure to run 'source env_writer rather than ./env_writer to properly source variables."

ENV_FILE="envfile"

# Read each line from the environment file
while IFS='=' read -r key value
do
  # Prepare the export command
  echo "export $key=\"$value\"" >> ~/.bashrc

  # Log the export to the terminal
  echo "Appended 'export $key=\"$value\"' to ~/.bashrc"
done < "$ENV_FILE"

echo "Run 'source ~/.bashrc' to load the new environment variables."
```


### Setting Up NGINX Production System

* Rather than creating an NGNX dockerfile with a nginx.conf file that is already hard-written, it's easier to just mount a nginx.conf file to the container, so that it can be tweaked as necessary to make it work.
* So within the docker-compose. file we mount our `nginx/nginx.conf` file to `/etc/nginx/nginx.conf` within the nginx docker container like so:

```
  nginx:
    ports:
      - target: 80
        published: 80
        protocol: tcp
        mode: host
    environment:
      - ENVIRONMENT=production
    volumes:
      - "${PWD}/nginx/nginx.conf:/etc/nginx/nginx.conf"
      # we may mount static asssets here in the future
    networks:
      - database_network
      - frontend_network
```

Then, the nginx.conf file itself should look like the following, in order to get HTTP working:

```
events {
    worker_connections 1024;
}

http {
    server {
        listen 80;
        listen [::]:80;

        server_name brierfoxforecast.com www.brierfoxforecast.com;

        # the following would redirect all http to https
        # return 301 https://$host$request_uri;

        # serve the frontend
        location / {
            # Docker DNS
            resolver 127.0.0.11 valid=30s;
            proxy_pass http://frontend:5173/;
            proxy_redirect default;
        }

        # serve the backend
        location /api/ {
            # Docker DNS
            resolver 127.0.0.11 valid=30s;
            proxy_pass http://backend:8080/;
            proxy_redirect default;
        }
    }
}
```

### Secure NGinx with SSL

Follow the following guide, ensuring that you install certbot, etc.

https://www.digitalocean.com/community/tutorials/how-to-secure-nginx-with-let-s-encrypt-on-ubuntu-22-04


However, for the ufw status, the firewall, whereas in the DigitalOcean documentation you typically have:

```
Output
Status: active

To                         Action      From
--                         ------      ----
OpenSSH                    ALLOW       Anywhere
Nginx HTTP                 ALLOW       Anywhere
OpenSSH (v6)               ALLOW       Anywhere (v6)
Nginx HTTP (v6)            ALLOW       Anywhere (v6)
```

For us, we typically should have `ufw status`

```
sudo ufw status
Status: active

To                         Action      From
--                         ------      ----
22/tcp                     LIMIT       Anywhere
2375/tcp                   ALLOW       Anywhere
2376/tcp                   ALLOW       Anywhere
22/tcp (v6)                LIMIT       Anywhere (v6)
2375/tcp (v6)              ALLOW       Anywhere (v6)
2376/tcp (v6)              ALLOW       Anywhere (v6)
```

To interpret these:

* 22/tcp LIMIT Anywhere: Limits incoming connections to port 22 (typically used for SSH) from any IP address. The LIMIT action is often used to reduce the risk of brute-force attacks by limiting the number of connections allowed over a period of time.
* 2375/tcp ALLOW Anywhere: Allows incoming connections to port 2375 (often used by Docker) from any IP address.
* 2376/tcp ALLOW Anywhere: Allows incoming connections to port 2376 (also often used by Docker, typically for encrypted connections) from any IP address.
* the v6 indicates ipv6 as opposed to ipv4.




### Configure config.js Domains

Make sure to mention config.js Domains to configure.

### Restarting SocialPredict Upon Upgrade



### Troubleshooting

#### Cannot Program Address

While attempting to re-build using `./compose-prod` you might find the following form of error:

```
Error response from daemon: failed to add interface veth17bdc95 to sandbox: error setting interface "veth17bdc95" IP to 172.20.0.4/16: cannot program address 172.20.0.4/16 in sandbox interface because it conflicts with existing route {Ifindex: 163 Dst: 172.20.0.0/16 Src: 172.20.0.1 Gw: <nil> Flags: [] Table: 254 Realm: 0}
```

