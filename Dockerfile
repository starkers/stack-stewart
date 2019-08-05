# backend
FROM golang:1.12-buster AS build-back
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -i -v -o agent  github.com/starkers/stack-stewart/cmd/agent
RUN CGO_ENABLED=0 GOOS=linux go build -i -v -o server github.com/starkers/stack-stewart/cmd/server

RUN ls -lash agent server ; pwd


# frontend
FROM node:lts-alpine AS build-front
WORKDIR /build
COPY . .
# will create a "dist" directory under /build/frontend
RUN cd frontend ; \
      npm install -g @vue/cli ; \
      npm run build


####
FROM alpine
# RUN apk add --no-cache tini
RUN addgroup -g 1000 app && adduser -D -G app -u 1000 -h /app app
WORKDIR /app
USER app

COPY --from=build-back  /build/agent .
COPY --from=build-back  /build/server .
COPY --from=build-front /build/frontend/dist public

# ENTRYPOINT ["/sbin/tini", "--"]
