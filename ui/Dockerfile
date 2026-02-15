FROM node:22-alpine AS build

RUN apk add --no-cache git

WORKDIR /app
COPY .git ./.git
COPY ui/package.json ui/package-lock.json ./
RUN npm ci
COPY ui/ .
RUN npm run build

FROM nginx:alpine

RUN apk add --no-cache openssl apache2-utils yq

COPY --from=build /app/dist /usr/share/nginx/html
COPY assets/icons/ /usr/share/nginx/html/
COPY ui/nginx/nginx.conf /etc/nginx/nginx.conf
COPY ui/nginx/default.conf /etc/nginx/conf.d/default.conf
COPY ui/entrypoint.sh /entrypoint.sh

EXPOSE 80 443

ENTRYPOINT ["/entrypoint.sh"]
