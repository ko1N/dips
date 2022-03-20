# build service
FROM golang:buster as builder
WORKDIR /app
COPY . .

WORKDIR /app/cmd/taskrunner/file_copy
RUN go get
RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o file_copy-service.out

# deploy
FROM alpine:latest
COPY --from=builder /app/cmd/taskrunner/file_copy/file_copy-service.out /usr/local/bin/file_copy-service

# add entrypoint script
WORKDIR /app
ADD build/taskrunner/file_copy/docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

ENTRYPOINT [ "./docker-entrypoint.sh" ]
CMD ["file_copy-service"]
