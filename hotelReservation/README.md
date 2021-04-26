# Hotel Reservation

The application implements a hotel reservation service, build with Go and gRPC, and starting from the open-source project https://github.com/harlow/go-micro-services. The initial project is extended in several ways, including adding back-end in-memory and persistent databases, adding a recommender system for obtaining hotel recommendations, and adding the functionality to place a hotel reservation. 

<!-- ## Application Structure -->

<!-- ![Social Network Architecture](socialNet_arch.png) -->

Supported actions: 
* Get profile and rates of nearby hotels available during given time periods
* Recommend hotels based on user provided metrics
* Place reservations

## Pre-requirements
### Runing dependencies
- Docker
- Docker-compose
- luarocks (apt-get install luarocks)
- luasocket (luarocks install luasocket)

### Developing dependencies
If changes are made to the Protocol Buffer files, the dependencies are:

- proto buffer: >= 3.0.0
- protoc-gen-go: 1.0.0

How to install protoc-gen-go==1.0.0:
```bash
GIT_TAG="v1.0.0"
go get -d -u github.com/golang/protobuf/protoc-gen-go
git -C "$(go env GOPATH)"/src/github.com/golang/protobuf checkout $GIT_TAG
go install github.com/golang/protobuf/protoc-gen-go

```


## Running the social network application
### Before you start
- Install Docker and Docker Compose.
- Make sure exposed ports in docker-compose files are available 
- Replace x.x.x.x in config.json with ip address of your servers

### Start docker containers
Start docker containers by running `docker-compose up -d`. All images will be 
pulled from Docker Hub.

#### workload generation
```bash
$WRK_DIR/wrk -D exp -t <num-threads> -c <num-conns> -d <duration> -L -s ./wrk2_lua_scripts/mixed-workload_type_1.lua http://x.x.x.x:5000 -R <reqs-per-sec>
```

### New proto and compile
If changes are made to the Protocol Buffer files use the Makefile to regenerate:

```bash
For all files at root directory:
make proto

For a certain file:
protoc user.proto --go_out=plugins=grpc:.
```

### Questions and contact

You are welcome to submit a pull request if you find a bug or have extended the application in an interesting way. For any questions please contact us at: <microservices-bench-L@list.cornell.edu>
