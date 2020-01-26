# Nginx

The entry point is an nginx Docker container. It proxies all the requests to other containers.

Nginx internally serves on port 80, but port 80 is mapped to the host's machine (your laptop's) port 4444. To visit the site in development, visit [localhost:4444](http://localhost:4444/).

See the `containers/nginx/default*.conf` files for details on how the requests are sent to the other containers.

```conf
# Defining the upstream applications. Gatsby, for example, is running on port 8000 in development.
upstream gatsby {
    server gatsby:8000;
}

# Express is running on port 5000
upstream express_api {
    server express_api:5000;
}

# Nginx listens on port 80 and proxies requests to the upstream servers.
server {
    listen 80;

    location /api {
        # Express doesn't know about the `/api` prefix, so this rewrites
        # the requests for Express.
        rewrite /api/(.*) /$1 break;
        proxy_pass http://express_api;
    }

    # Routes not matched above will go to Gatsby.
    location / {
        proxy_pass http://gatsby;
    }
}
```
