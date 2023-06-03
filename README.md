# CIFuzzer-Server

## Run with Host

### Prerequisites

```bash
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
