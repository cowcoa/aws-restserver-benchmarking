FROM public.ecr.aws/docker/library/node:latest

WORKDIR /usr/src/app

COPY package*.json ./
COPY server.js ./

RUN npm install

EXPOSE 8387
CMD [ "node", "server.js" ]
