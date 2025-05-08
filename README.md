# Simple Contents Service

A Go-based service for managing and storing content with support for multiple storage backends.

## Features

- RESTful API built with Chi router
- Support for multiple storage backends:
  - Google Cloud Storage
  - AWS S3
  - MinIO
- UUID-based content identification
- SQL database integration

## Prerequisites

- Go 1.23.8 or higher
- Docker (optional, for containerized deployment)
- One of the supported storage backends:
  - Google Cloud Storage credentials
  - AWS S3 credentials
  - MinIO configuration

## Installation

1. Clone the repository:
```bash
git clone https://github.com/livefire2015/simple-contents.git
cd simple-contents
```

2. Install dependencies:
```bash
make dep
```

## Building

To build the application:
```bash
make
```

This will create executable binaries in the `dist` directory:
- `dist/cmd/server`: Server binary

## Running

To run the application:
```bash
make run
```

## Docker

To build and run the application in a Docker container:

```bash
make docker-build
docker run -p 8080:8080 simple-contents
```

## Project Structure

```
simple-contents/
├── cmd/              # Command-line applications
│   └── server/       # Main server application
├── model/           # Data models
├── repository/      # Data access layer
├── service/         # Business logic
├── storage/         # Storage backends
└── transport/       # API transport layer
```

## Cleaning

To clean up build artifacts:
```bash
make clean
```
