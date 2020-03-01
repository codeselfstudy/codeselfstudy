# This is a multi-stage build process that builds the Gatsby site
# and copies the production output into an nginx image. That nginx
# server runs on port 8000, which is upstream of port 80 on the main
# user-facing nginx server.
FROM node:12-alpine as builder
RUN apk update && apk add --no-cache --virtual .gyp python make g++
WORKDIR /app
ENV GATSBY_TELEMETRY_DISABLED=1
RUN mkdir -p ~/.config/gatsby/
COPY ./misc/config.json ~/.config/gatsby/
ENV NODE_ENV=production
COPY ./package.json ./
RUN npm install
COPY . .
COPY ./src/content/files ./static/
RUN npm run build

FROM nginx:alpine
EXPOSE 8000
COPY ./nginx/default.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/public /usr/share/nginx/html
