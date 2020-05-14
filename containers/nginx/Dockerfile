FROM nginx
EXPOSE 80
EXPOSE 443

RUN mkdir -p /etc/letsencrypt/live/codeselfstudy.com/

COPY ./redirects.conf /etc/nginx/redirects.conf
COPY ./default.conf /etc/nginx/conf.d/default.conf
