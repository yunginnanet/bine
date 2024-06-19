module github.com/yunginnanet/bine/examples

go 1.22

toolchain go1.22.4

require (
	github.com/golang/protobuf v1.5.3
	github.com/yunginnanet/bine v0.2.0
	golang.org/x/net v0.26.0
	google.golang.org/grpc v1.38.0
)

replace github.com/yunginnanet/bine => ../
