FROM nginx:1.27-alpine
COPY webapp/nginx.conf /etc/nginx/conf.d/default.conf
COPY webapp/ /usr/share/nginx/html
