# build service
FROM golang:buster as builder
WORKDIR /app
COPY . .

WORKDIR /app/cmd/taskrunner/shell
RUN go get
RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o shell-service.out

# deploy
FROM alpine:latest
COPY --from=builder /app/cmd/taskrunner/shell/shell-service.out /usr/local/bin/shell-service

# add entrypoint script
WORKDIR /app
ADD build/taskrunner/shell/docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

ENTRYPOINT [ "./docker-entrypoint.sh" ]
CMD ["shell-service"]
