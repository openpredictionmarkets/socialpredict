# nginx/Dockerfile
FROM nginx:latest

COPY nginx-start.sh /nginx-start.sh
CMD ["/nginx-start.sh"]
