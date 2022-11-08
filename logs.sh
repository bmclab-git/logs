#!/bin/sh
docker rm -f logs
docker rmi logs
docker load -i ./logs.tar
docker run --name logs -d --restart always --network host -v /data/logs/web.json:/app/web.json -v /data/logs/grpc.json:/app/grpc.json logs