# syntax=docker/dockerfile:1.3-labs
FROM node:20.0.0

RUN echo "hi"

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm install
COPY . .

EXPOSE 5173
CMD ["npm", "run", "start"]