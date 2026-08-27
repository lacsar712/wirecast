# wirecast

wirecast 是一套连铸结晶器冷却工业过程控制系统，用于冷却回路监测、调节与过程联锁。

## Requirements

- Go 1.22+ (container image uses golang:1.22)
- `GOTOOLCHAIN=local` recommended on host when using a pinned toolchain

## Build

```bash
export GOTOOLCHAIN=local
go build ./...
```

## Run

```bash
export GOTOOLCHAIN=local
go run ./cmd/graindry
```

## Test

```bash
export GOTOOLCHAIN=local
go test ./... -count=1
```

## Docker (benzhi)

Must build **linux/amd64** and **linux/arm64**:

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh wirecast linux/amd64
./build_benzhi_docker.sh wirecast linux/arm64
docker run -it wirecast:latest
# inside container:
export GOTOOLCHAIN=local
go version
go build ./...
go test ./... -count=1
```
