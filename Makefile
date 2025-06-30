BINARY_NAME := workbench

build:
	go build -o $(BINARY_NAME) ./cmd

run:
	go build -o $(BINARY_NAME) ./cmd
	./$(BINARY_NAM)

clean:
	go clean
	rm ${BINARY_NAME}
