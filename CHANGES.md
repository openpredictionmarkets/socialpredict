# Deploy To Production

## Changes made:

### backend directory
* removed .env.dev file
* removed builddockerfile
* removed buildproddockerfile
* renamed backend.Dockerfile to Dockerfile

### frontend directory
* removed builddockerfile
* removed buildproddockerfile
* renamed frontend.Dockerfile to Dockerfile
* added file src/config.js.template to change the variables in the src/config.js file

### nginx directory
* removed entire directory

### data directory
* created data directory that will hold all the relevant files - Done
* data/postgres directory will hold the database data in order to mainain data when containers are restarted.
* data/certbot directory will hold all the required files, including the SSL certificate for the domain.
* data/nginx directory will hold all the required files for nginx ( Templates & Config Files ).

### scripts directory
* created scripts directory that will hold all the relevant scripts for the deployment.
* docker-compose-prod.yaml is the main file for docker compose.
* build.sh is the script that builds the required images.
* ssl.sh is the script that generates the SSL certificate for the domain.
* env_writer.sh is the script that will generate the .env file with the data from user input.
* exec.sh is the script that will perform exec command on docker in order to connect to the container.
* compose.sh is the script that will perform docker compose up and down.

### .env files
* removed all .env files and created a .env.example file that will hold all the required environment variables with generic values
* added .env file to .gitignore

### root directory
* removed file builddockerfiles - It will be handled by the ./scripts/build.sh script.
* removed all docker-compose.yaml files. They will be stored in the scripts directory.
* removed all compose scripts. They will be handled by the socialpredict script.
* removed all exec scripts. They will be handled by ./scripts/exec.sh script.
* removed start file.
* added script socialpredict that will handle all logic.
* removed env_writer file. It will be handled by the ./scripts/env_writer.sh script.
* removed entrypoint - Empty File.

## How to use:

1. Clone repository:
```
git clone https://github.com/openpredictionmarkets/socialpredict.git
```

2. Enter repository directory
```
cd socialpredict
```

3. Make sure SocialPredict file is executable
```
chmod +x SocialPredict
```

4. At this point, before proceeding any further make sure you already have a domain/subdomain and it is pointing to your server.
```
nslookup -type=A domain
```
The Address returned should match the address of your server

5. Run SocialPredict Script
```
sudo ./SocialPredict install
```

6. The script will initialize the application, create the necessary .env file and ask for user input on some variables:
* DOMAIN: the domain name you want to use - Required
* EMAIL: a valid email address to be used by Let's Encrypt - Required
* Database User: the name for the database user - Optional (defaults to user)
* Database User Password: the password for the database user - Optional (defaults to password)
* Database Name: the name for the database - Optional (defaults to socialpredict_db)
* Admin Password: the password for the admin user - Optional (defaults to password)

7. Then the script will proceed to build the required images and issue the SSL Certificate for the Domain.

8. After the script is completed, you can start the application with:
```
sudo ./SocialPredict up
```

9. And stop it with:
```
sudo ./SocialPredict down
```

## Logic Behind The Script:
* Each time you run the script (SocialPredict) first of all it checks if docker compose is installed or not. It only supports docker compose and not docker-compose since docker-compose has stopped receiving security updates for about a year now.
* During the first time the script is run, it create a .env file by copying the .env.example file and asks the user for input on some specific variables. In the subsequent runs it doesn't perform this step.
* Then it checks for a valid .env file and sources it.
* Then it checks if the docker images are present on the server or not. If they are not present, it builds them using the .env file. If they are present, it asks the user if they want to re-build them or not.
* Then it proceeds to issue an SSL Certificate for the domain specified in the .env. If it finds an SSL present in the directory, it asks for confirmation to delete it and issue a new one. If it doesn't get confirmation, it exits.
* After that everything is ready to run with ./SocialPredict up

## Issues:
* We need to fix the issue with the frontend building. Since this is a production envrionment, the frontend should build the files and serve them directly from Nginx. This way there is no need for a frontend container to be running. It should only run one time to build and then exit. Current issue when building is "Uncaught ReferenceError: require is not defined https://brierfoxforecast.com/assets/index-Yldixx8y.js:54".
* The Backend Server seems to take too much time to start. We need to figure out a way to debug it and improve the performance.When you start the backend container you get these logs:
Server started with PID 8
Setting up watches.  Beware: since -r was given, this may take a while!
Watches established.

But the server is not up and running until you get these:
Successfully connected to the database.
Starting server on :8080

It seems like the backend takes too much time to connect to the db. It can take up to 2 minutes, sometimes even more.

## Todo:
* Improve current scripts, specifically add more options to the compose.sh script which is responsible for running docker compose command and the exec.sh which is responsible for executing commands directly to the containers.
* Add more options to the SocialPredict script.
* Add a help section to the SocialPredict script to offer help messages regarding its usage.
* Figure out a way to remove ADMIN_PASSWORD from the .env file after the admin user has been created.
