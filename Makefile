CGO_ENABLED:=1

.EXPORT:
	export CGO_ENABLED

all : deps prepStaticTor

prepStaticTor : .EXPORT
	go run setup.go

deps :
	go mod tidy -v
