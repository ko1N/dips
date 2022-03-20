# build service
FROM golang:buster as builder
WORKDIR /app
COPY . .

WORKDIR /app/cmd/manager
RUN go get
RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o manager.out

# deploy
FROM alpine:latest
RUN apk add --no-cache ffmpeg
COPY --from=builder /app/cmd/manager/manager.out /usr/local/bin/manager

# add entrypoint script
WORKDIR /app
ADD build/manager/docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

# gin flags
ENV PORT=8080
ENV GIN_MODE=release

ENTRYPOINT [ "./docker-entrypoint.sh" ]
EXPOSE 8080
CMD ["manager"]
