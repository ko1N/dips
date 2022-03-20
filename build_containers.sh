#!/bin/bash

# taskrunners
docker build -t ffmpeg-dips-service -f build/taskrunner/ffmpeg/ffmpeg-service.dockerfile .
docker build -t file_copy-dips-service -f build/taskrunner/file_copy/file_copy-service.dockerfile .
docker build -t shell-dips-service -f build/taskrunner/shell/shell-service.dockerfile .
