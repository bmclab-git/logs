FROM scratch
WORKDIR /app
COPY main ./
COPY web.json ./
COPY grpc.json ./
ENTRYPOINT ["./main"]