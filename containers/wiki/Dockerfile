# TODO: fix this file

this line will break it -- this file shoulnd't be used yet.

FROM rust:alpine3.11 as builder
# RUN apk update
#&& apk add build-base autoconf automake libtool pkgconfig nasm
WORKDIR /app
COPY ./ ./
RUN cargo install mdbook
RUN mdbook build

# Get a clean image with gatsby-cli and the pre-built node modules
FROM nginx



COPY ./default_dev.conf /etc/nginx/conf.d/default.conf
