package ss_mgr

//go:generate echo "Generating protobuf code..."
//go:generate protoc manager/protocol/shadowsocks_manager.proto --go_out=plugins=grpc:.
//go:generate protoc plugin/protocol/plugin_core.proto --go_out=plugins=grpc:.
