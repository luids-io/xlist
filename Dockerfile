FROM golang:alpine as build-env
ARG arch=amd64

# Install git and certificates
RUN apk update && apk add --no-cache git make ca-certificates && update-ca-certificates

# create user for service
RUN adduser -D -g 'luids' luxlist

WORKDIR /app
## dependences
COPY go.mod .
COPY go.sum .
RUN go mod download

## build
COPY . .
RUN make binaries SYSTEM="GOOS=linux GOARCH=${arch}"

## create docker
FROM scratch

LABEL maintainer="Luis Guillén Civera <luisguillenc@gmail.com>"

# Import the user and group files from the builder.
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-env /etc/passwd /etc/passwd

COPY --from=build-env /app/bin/xlist? /bin/
COPY --from=build-env /app/configs/docker/xlistd/*  /etc/luids/xlist/

USER luxlist

EXPOSE 5801
VOLUME [ "/etc/luids", "/var/lib/luids/xlist" ]
CMD [ "/bin/xlistd", "--config", "/etc/luids/xlist/xlistd.toml" ]
