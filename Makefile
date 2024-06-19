CGO_ENABLED:=1

.EXPORT:
	export CGO_ENABLED

all : deps prepStaticTor

prepStaticTor : .EXPORT
	go run setup.go

fixProcessGo :
	go run setup.go -skip-compile

deps :
	go mod tidy -v

clean :
	rm -vrf process/embedded/tor-static
