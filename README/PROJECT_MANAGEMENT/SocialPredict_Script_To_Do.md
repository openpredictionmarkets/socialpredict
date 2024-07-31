# To-Do List for the SocialPredict script:

### Essential

The following were moved from Nice to Have below into here, the essential category.

4. Add database backup functionality:
   * Add a command like './SocialPredict db backup' that will backup current postgres database to the location specified by the user. Also add option to compress the backup.

5. Add database restore functionality:
   * Add a command like './SocialPredict db restore' that will restore the postgres database from the file specified by the user. Make sure it works with compressed files as well.

6. Add option to import database on first run:
   * When running './SocialPredict install' ask user if they want to import a database backup or use a clean database.
     * The logic will be like:
       1. Ask user if they want to use a backup for database.
       2. If they answer yes, create a postgres container.
       3. Restore the database with the user provided backup file.
       4. Remove postgres container.
       5. Continue with the rest of the installation.


### Nice to Have

1. Fix certbot docker network on SSL initialization: - Only for Production.
   * Currently when certbot is run for the first time to generate a new certificate, it generate a random-named network in docker. Change this network name and make sure it is deleted when the certificate is issued.

2. Add commands for SSL renew: - Only for Production
   * Add necessery commands in docker-compose-prod.yaml to force certbot to renew SSL Certificate and Nginx to reload the certificate.

3. Migrate changes from build_dev.sh to build_prod.sh:
   * Make sure the changes we made in build_dev.sh are present in the build_prod.sh as well.

7. Add option to remove Docker Images:
   * Add a command like './SocialPredict images remove' that will remove all the images built from docker. It will only remove the images built by our script.

8. Add option to re-build images:
   * Add a command like './SocialPredict images rebuild' that will rebuild our images. Add option to rebuild a specific image or all of them.

9. Add option to restart containers:
   * Add a command like './SocialPredict restart' that will restart all docker containers.

10. Add restart: always to containers:
    * Add 'restart: always' in docker-compose files so that the application automatically starts when the server is rebooted. Also make sure docker.service is enabled.

11. Figure out a way to implement a service like [bunkerweb](https://github.com/bunkerity/bunkerweb): - Only for Production
    * Bunkerweb is a service that offers security features like firewall, antibot etc. for an application.

12. Add option to update .env: - Only for Production
    * Add a command like './SocialPredict env update' that will update the .env file. It will ask for user input on all variables like the installation process, or update a specific variable based on user input like './SocialPredict env update --domain example.com'. Make sure it re-builds the images if necessary and warn user about that.

13. Add help information on script usage:
    * Add a command like './SocialPredict help' that will display information on script usage. Also it will provide information on specific commands.

14. Add option to add script to user's path:
    * Either add a command or ask user on installation to add SocialPredict script on user's path so it can be run as "SocialPredict" from anywhere instead of using "./SocialPredict" inside the socialpredict folder.

15. Add option to backup database in S3:
    * Add option and necessary config files to allow users to backup database to an S3 type storage.

16. Verify domain input:
    * Validate user input as a domain name like domain.extension

17. Option for www or non-www domain:
    * Ask user if they want to use www or non-www version of their domain and modify config files accordingly.

18. Update nginx configuration:
    * Update nginx config files to use backend and frontend ports based on .env values.

19. Add option to install docker and docker compose plugin if not present on the server.
    * If docker and docker compose plugin are not present, ask user if they want the script to install them. Get the distribution name and proceed with the installation accordingly.

20. Add option to utilize an already existing nginx service to serve the application.
    * If the user already has nginx on their server, give them the option to use that nginx service to host the app. Provide necessary configuration and instructions.

21. Add options start/stop:
    * Add commands './SocialPredict start' & './SocialPredict stop' that will have same functionality as 'up' and 'down'.

22. Add option to clean up:
    * Add command to remove docker images/networks/volumes and also the posibility to remove all files like a 'unistall' process.

