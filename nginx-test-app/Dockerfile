FROM nginx:1.13.5

EXPOSE 8080

COPY nginx.conf /etc/nginx/nginx.conf
COPY default.conf /etc/nginx/conf.d/default.conf

RUN rm -rf /var/log/nginx/access.log