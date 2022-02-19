#!/bin/bash
go install github.com/swaggo/swag/cmd/swag@latest
cd cmd/manager && go generate
