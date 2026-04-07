# Simple Object Storage Service

A lightweight RESTful HTTP service for storing, retrieving, and deleting objects organized by buckets. Objects are persisted to disk with content-based deduplication.

## Prerequisites

- Go 1.22+

## Quick Start

```bash
# Build and run
make run

# Or with custom port and data directory
PORT=9090 DATA_DIR=/tmp/objstore make run
```

The service listens on port `8080` by default.

## Configuration

The server can be configured by providing the following environment variables:

| Variable   | Default  | Description                     |
|------------|----------|---------------------------------|
| `PORT`     | `8080`   | Port the server listens on      |
| `DATA_DIR` | `./data` | Directory for object storage    |

## API Reference

### PUT /objects/{bucket}/{objectID}

Store an object. Identical content is deduplicated across all buckets.

```bash
curl -X PUT -d "hello world" http://localhost:8080/objects/mybucket/greeting
```

- **Response (created/updated):** `201 Created` with body: `{"id": "greeting"}`

### GET /objects/{bucket}/{objectID}

Retrieve an object with ID `objectID` from bucket `bucket` if it exists.

```bash
curl http://localhost:8080/objects/mybucket/greeting
```

- **Response (found):** `200 OK` with object data in the body
- **Response (not found):** `404 Not Found`

### DELETE /objects/{bucket}/{objectID}

Delete an object with ID `objectID` from bucket `bucket` if it exists.

```bash
curl -X DELETE http://localhost:8080/objects/mybucket/greeting
```

- **Response (found):** `200 OK`
- **Response (not found):** `404 Not Found`

## Development

The `Makefile` provides a few utilities to assist in testing and developing the server

### Build

To build an executable for the server, simply run:

```bash
make build
```

### Run Tests

To run the server's unit tests, simply run:

```bash
make test
```

### Container

To build and run a portable container image, run:

```bash
make image
podman run -p 8080:8080 bucket-store-server
```

## AI Usage Disclosure

AI tools (Cursor with Claude) were used during development for:

- **Design discussion**: Evaluating language/framework choices, project structure, and trade-offs before writing code
- **Scaffolding**: Generating initial project skeleton, boilerplate code, and test structure
- **Code review**: Identifying dead code, godoc improvements, and status code inconsistencies in the spec
- **Iterative refinement**: Each layer was reviewed and simplified based on discussion before moving to the next

All AI-generated code was reviewed, tested, and modified by the developer. The iterative approach (stub server → store layer → server layer → packaging) was driven by the developer to maintain understanding and control at each step.
