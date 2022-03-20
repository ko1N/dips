# build service
FROM golang:buster as builder
WORKDIR /app
COPY . .

WORKDIR /app/cmd/jobrunner
RUN go get
RUN CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-w -extldflags "-static"' -o jobrunner.out

# deploy
FROM alpine:latest
RUN apk add --no-cache ffmpeg
COPY --from=builder /app/cmd/jobrunner/jobrunner.out /usr/local/bin/jobrunner

# add entrypoint script
WORKDIR /app
ADD build/jobrunner/docker-entrypoint.sh .
RUN chmod +x docker-entrypoint.sh

ENTRYPOINT [ "./docker-entrypoint.sh" ]
CMD ["jobrunner"]
