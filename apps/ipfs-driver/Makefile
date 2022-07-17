all:
	GOOS=linux GOARCH=amd64 go build -o bin/amd64-linux/ipfs-driver github.com/eagraf/habitat/apps/ipfs-driver
	GOOS=darwin GOARCH=amd64 go build -o bin/amd64-darwin/ipfs-driver github.com/eagraf/habitat/apps/ipfs-driver

clean:
	rm -rf bin
