# CIFuzzer-Server

## Run with Host

### Prerequisites

```bash
# Need to use "screen" to display each task of fuzzing
sudo apt update
sudo apt install screen

golang ver 1.18
AFL++
```

### You can build AFL++ using the shell script

```bash
chmod +x build_afl.sh
./build_afl.sh
```

### Run API Server

```bash
go mod tidy
go run CIServer.go
```

### Test API Server

```bash
# Need to prepare your c file,
# See CIServer_test.go and change path
go test
```

## Run with Docker

### Prerequisites

```bash
docker
docker compose
```

## Build Docker

```bash
docker build -t ci_fuzzer .
```

## Run & Stop Docker

```bash
# Run Docker, default port 8080
docker compose up

# Stop Docker
docker compose down
```
