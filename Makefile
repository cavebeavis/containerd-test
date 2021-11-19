.PHONY: all test go_test build clean

#COUNT:=0

all: go_test build clean

test: go_test clean

go_test:
#	COUNT:=$(COUNT)+1
#	@echo "$$COUNT" 
	@echo "Performing go test..."
#	Do testing which can return other than exit(0)

	@echo "Testing OK"
	@echo ""

build:
	@echo "Building docker images..."
#	docker build stuff

	@echo "Finished building docker images!"
	@echo ""

clean:
	@echo "Cleaning up..."
#	clean up any temp or other stuff like the ctr/ip networking, namespaces, and links.

	@echo "Done cleaning up!"
	@echo ""