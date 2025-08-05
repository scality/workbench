BINARY_NAME := workbench

build:
	go build -o $(BINARY_NAME) ./cmd

run:
	go build -o $(BINARY_NAME) ./cmd
	./$(BINARY_NAM)

lint:
	golangci-lint run

clean:
	go clean
	rm ${BINARY_NAME}
