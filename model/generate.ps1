# --go_out和--go_opt指定生成dto文件
protoc  -I ./ --go_out=./ --go_opt=paths=source_relative ./*.proto

# --go-grpc_out和--go-grpc_opt指定生成服务文件
protoc -I ./ --go-grpc_out=require_unimplemented_servers=false:./ --go-grpc_opt=paths=source_relative ./*.proto