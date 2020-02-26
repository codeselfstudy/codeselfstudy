FROM node:12-alpine
WORKDIR "/app"
ENV NODE_ENV='production'
COPY ./package.json ./
RUN npm install
COPY . .
CMD ["npm", "run", "production"]
