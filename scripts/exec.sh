#!/bin/bash

#################################################################
#                                                               #
# Script By Vasileios Ntoufoudis                                #
# info@ntoufoudis.com                                           #
#                                                               #
#################################################################

# Make sure the script can only be run via SocialPredict Script
[ -z "$CALLED_FROM_SOCIALPREDICT" ] && { echo "Not called from SocialPredict"; exit 42; }

if [ "$1" = "nginx" ]; then
	if [ -z "$2" ]; then
		docker exec -it "${NGINX_CONTAINER_NAME}" /bin/bash
	else
		docker exec "${NGINX_CONTAINER_NAME}" "$2"
	fi
elif [ "$1" = "backend" ]; then
        if [ -z "$2" ]; then
                docker exec -it "$BACKEND_CONTAINER_NAME}" /bin/bash
        else    
                docker exec "${BACKEND_CONTAINER_NAME}" "$2"
        fi      
elif [ "$1" = "frontend" ]; then
        if [ -z "$2" ]; then
                docker exec -it "${FRONTEND_CONTAINER_NAME}" /bin/bash
        else    
                docker exec "${FRONTEND_CONTAINER_NAME}" "$2"
        fi      
elif [ "$1" = "postgres" ]; then
        if [ -z "$2" ]; then
                docker exec -it "${POSTGRES_CONTAINER_NAME}" /bin/bash
        else    
                docker exec "${POSTGRES_CONTAINER_NAME}" "$2"
        fi      
else
	echo "Wrong Container Name."
fi
