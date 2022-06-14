vps:
	go build -v
mac:
	CGO_CFLAGS=-I/opt/homebrew/include CGO_LDFLAGS=-L/opt/homebrew/lib make
