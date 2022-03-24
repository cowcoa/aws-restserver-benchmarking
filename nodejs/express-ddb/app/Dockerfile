FROM public.ecr.aws/docker/library/node:16.14

WORKDIR /usr/src/app

COPY package*.json ./
COPY server.js ./

RUN npm install

EXPOSE 8387
CMD [ "npm", "start" ]
