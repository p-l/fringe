BINDIR := bin
BINFILE := mellon-server
GOMAIN = main.go

all: test build run

clean:
	go clean
	rm -rf $(BINDIR)
	rm *.db

dep:
	go mod download

lint:
	golangci-lint run --enable-all --fix

test: lint
	go test $(GOMAIN)

build: dep
	mkdir -p $(BINDIR)
	go build -o $(BINDIR)/$(BINFILE) $(GOMAIN)

run: build
	$(BINDIR)/$(BINFILE)
