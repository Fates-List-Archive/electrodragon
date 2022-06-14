vps:
	go build -v
mac:
	CGO_CFLAGS=-I/opt/homebrew/include CGO_LDFLAGS=-L/opt/homebrew/lib make
dev:
	LIST_ID=123 SECRET_KEY=123 ./wv2
