# nginx/Dockerfile
FROM nginx:latest
COPY nginx.conf.template /etc/nginx/nginx.conf.template
COPY nginx-start.sh /nginx-start.sh
CMD ["/nginx-start.sh"]
