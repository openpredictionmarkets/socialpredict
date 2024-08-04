### About

The following is a general tutorial geared at getting [SocialPredict](https://github.com/openpredictionmarkets/socialpredict) set up on Digital Ocean. Other virtual machine servers may work similarly, this serves as an example and is designed for more novice users.

### Setting Up Your Domain

First, you should purchase a domain, which you can do at any number of domain registrars. We are using [Namecheap](https://www.namecheap.com/) for this article.

### Set Up Digital Ocean Account

We are using Digital Ocean as a server, since it's predictable, easy and cheap. SocialPredict is designed to be performant, meaning it can perform well on a small server.

Once you have bought your domain, set up an account on [Digital Ocean](https://www.digitalocean.com/). You will want to set up a Droplet with Docker. Instructions for how to do this are here:

[Tutorial on Setting Up Droplet with Docker](https://www.digitalocean.com/community/tutorials/how-to-use-the-docker-1-click-install-on-digitalocean)

If you want to skip that tutorial and are fairly sure how to use Digital Ocean already, you can simply click this [Digital Ocean Marketplace Button here](https://marketplace.digitalocean.com/apps/docker) to just get going.

### SSH

Generally you should follow this tutorial to get connected via ssh to your Docker-based Digital Ocean droplet.

You can follow these instructions to [get started with SSH](https://docs.digitalocean.com/products/droplets/how-to/connect-with-ssh/).

#### Logging In via SSH

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

After deploying Docker, you can follow [this tutorial to verify that Docker is working](https://marketplace.digitalocean.com/apps/docker?#getting-started).

Alternatively, you can run the following commands to verify that the variosu tools are installed. Below shows some sample results which should be relatively like what should come up.

* `docker version`
* `docker compose version`
* `docker buildx version`

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

...

user@server:~# docker compose version
Docker Compose version v2.17.2

...

user@server:~# docker buildx version
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

Once you have a domain name, reference your domain registrar's documentation for [connecting up your domain to Digital Ocean]
(https://www.namecheap.com/support/knowledgebase/article.aspx/10375/2208/how-do-i-link-a-domain-to-my-digitalocean-account/).

For more documentation from Digital Ocean's perspective, [check out this tutorial here](https://docs.digitalocean.com/products/networking/dns/getting-started/dns-registrars/).

### Add a Domain

Once the Domain has been pointed to Digital Ocean, add a Domain within the [Digital Ocean Networking Dashboard](https://cloud.digitalocean.com/networking/domains).

### Create A Records in The Control Panel

Once this application has been set up, as mentioned above, create A records for `www` and `@` for your droplet. This can be found by navigating to https://cloud.digitalocean.com/networking/domains/`{yourdomain}`

There is a Digital Ocean tutorial on [setting up and managing DNS records here](https://docs.digitalocean.com/products/networking/dns/how-to/manage-records/#create-update-and-delete-records-using-the-control-panel), [as well as this section specifically about a records](https://docs.digitaloc.ean.com/products/networking/dns/how-to/manage-records/#a-records)

Ensure that both `www` and `@` are created.  After all of this has been updated, it may take 48 hours to take effect.

* Status of your domain can be checked at https://www.whatsmydns.net/
* The `dig` tool on the local command line can be used to check a domain's status (for all domains):

```
dig www.yourdomain.com
dig yourdomain.com
```

### Create a Firewall

* Firewalls are ways to restrict port accesses to only a narrow number of ports, or one port for a given purpose. This tutorial goes through [the process of setting up a firewall for your application](https://docs.digitalocean.com/products/networking/firewalls/how-to/create/#create-a-firewall-using-the-control-panel).

#### HTTP and HTTPS:

* Use port 80 for HTTP.
* HTTPS (HTTP Secure) is the secure version of HTTP, encrypted using TLS (SSL). Use port 443 for HTTPS.

### Clone the SocialPredict Repo

Now that the domain has been pointed in the right direction, you can clone the SocialPredict repo. First, check what version of `git` you are running by typing the following in the console:

```
root@DROPLET_NAME:# cd /home
root@DROPLET_NAME:/home# git version
git version 2.34.1
```
Now it's time to clone the repo:

```
root@DROPLET_NAME:/home# git clone https://github.com/openpredictionmarkets/socialpredict.git
```
From here, navigate to your `socialpredict` folder:

```
root@DROPLET_NAME: cd /home/socialpredict
```
And run `./SocialPredict install`:

```
root@DROPLET_NAME:/home/socialpredict ./SocialPredict install
```

SocialPredict will prompt you to type `1` for development, `2` for production, or `3` to quit. Hit `2` on your keyboard to start.

![Screenshot 2024-08-03 111737](https://github.com/user-attachments/assets/3423a047-213b-48cc-baf3-4eadb326c111)

Next, SocialPredict will prompt you for the name of your domain.

![Screenshot 2024-08-03 111838](https://github.com/user-attachments/assets/5c32618c-17be-4e0c-a74d-f5d546b0b8bf)

It will also prompt you for the email address linked to your SSL certificate.

![Screenshot 2024-08-03 111905](https://github.com/user-attachments/assets/d35a4ce8-991f-4389-96aa-e9920e87ccca)

Type in the default username.

![Screenshot 2024-08-03 111941](https://github.com/user-attachments/assets/1ad1062e-439e-40aa-ad2c-430b9c491801)

Specify a default password.

![Screenshot 2024-08-03 112119](https://github.com/user-attachments/assets/381dc8ac-3c43-4695-a477-6656c98be933)

Specify a name for the database.

![Screenshot 2024-08-03 112155](https://github.com/user-attachments/assets/36571a06-ae6b-464b-b974-bec11361a2e8)

Lastly, choose an admin password.

![Screenshot 2024-08-03 112236](https://github.com/user-attachments/assets/9bef37ff-8299-46f5-babd-6a2d54aef44d)

Once you're done, type `./SocialPredict up` to spin up a SocialPredict instance ready to deploy to the web, and navigate to your domain to see if it works.

If you want to spin down your SocialPredict instance, just type `./SocialPredict down` in your console.

### Prod Tools

* There are several debugging tools which can be used on prod to help figure out what might be going wrong in different instances.

#### Logs

* Logs can be useful for finding out what is going wrong under the hood.
* To access logs, `ssh` in to your droplet and use the section on "Getting Logs from Different Containers" within the LOCAL_SETUP.md [here](/README/LOCAL_SETUP.md#getting-logs-from-different-containers).

#### Entering a Container

##### Database

* Log into the psql service directly (use with CAUTION, manipulating the database can lead to serious errors):

```
docker exec -it -e PGPASSWORD=${POSTGRES_PASSWORD} socialpredict-postgres-container psql -U ${POSTGRES_USER} -d socialpredict_db
```

