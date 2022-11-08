#!/bin/sh
docker rm -f ecom_logs
docker rmi ecom_logs
docker load -i ./ecom_logs.tar
docker run --name ecom_logs -d --restart always --network host -v /data/ecom_logs/web.json:/app/web.json -v /data/ecom_logs/grpc.json:/app/grpc.json ecom_logs