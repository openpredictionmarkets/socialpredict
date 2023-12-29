# syntax=docker/dockerfile:1.3-labs
FROM node:16.0.0

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm install
COPY . .

EXPOSE 3000
CMD ["npm", "start"]