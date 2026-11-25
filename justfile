app := "kothay"
pkg := "main"

version := `git describe --tags --always --dirty 2>/dev/null || echo dev`
commit := `git rev-parse --short HEAD 2>/dev/null || echo none`
date := `date -u +%Y-%m-%dT%H:%M:%SZ`
ldflags := "-s -w -X " + pkg + ".Version=" + version + " -X " + pkg + ".Commit=" + commit + " -X " + pkg + ".Date=" + date

default:
    @just --list

# Build binary for local machine
build:
    go build -ldflags="{{ ldflags }}" -o bin/{{ app }}-{{ version }} .

# Clean generated binaries
clean:
    rm -rf bin/
