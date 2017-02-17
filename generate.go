package ssmgr

//go:generate echo "Generating protobuf code..."
//go:generate protoc protocol/master_slave.proto --go_out=plugins=grpc:.
