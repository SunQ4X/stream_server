@echo off
protoc -I . --go_out=plugins=grpc:. ./playback.proto