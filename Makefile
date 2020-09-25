APPLICATION_NAME = ministub

all:
	@$(MAKE) create-dir deps test build success || $(MAKE) failure

deps: 
	go mod download
	go mod verify

build:
	CGO_ENABLED=0 GOOS=linux go build -o bin/${APPLICATION_NAME} -v cmd/${APPLICATION_NAME}.go

ide-build:
	@$(MAKE) build success || $(MAKE) failure

install:
	cp ./bin/ministub /usr/local/bin

clean:
	go clean
	if [ -f ./bin/${APPLICATION_NAME} ]; then rm ./bin/${APPLICATION_NAME}; fi;
	if [ -f ./coverage.html ]; then rm ./coverage.html; fi;
	if [ -f ./coverage.out ]; then rm ./coverage.out; fi;

test:
	go test ./... -coverprofile=coverage.out -bench . -count=1
	go tool cover -html=coverage.out -o coverage.html

success:
	printf "\n\e[1;32mBuild Successful\e[0m\n"

failure:
	printf "\n\e[1;31mBuild Failure\e[0m\n"
	exit 1

docker-build:
	docker-compose build

create-dir:
	if ! [ -d ./bin ]; then mkdir bin; fi;