# build service
FROM golang:buster as builder
WORKDIR /app
COPY . .

WORKDIR /app/cmd/taskrunner/ffmpeg
RUN go get
RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o ffmpeg-service.out

# deploy
FROM alpine:latest
RUN apk add --no-cache ffmpeg
COPY --from=builder /app/cmd/taskrunner/ffmpeg/ffmpeg-service.out /usr/local/bin/ffmpeg-service

# add entrypoint script
WORKDIR /app
ADD build/taskrunner/ffmpeg/docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

ENTRYPOINT [ "./docker-entrypoint.sh" ]
CMD ["ffmpeg-service"]

ENV NVIDIA_DRIVER_CAPABILITIES all
