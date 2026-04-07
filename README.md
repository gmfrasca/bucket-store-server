# Simple Object Storage Service

A lightweight RESTful HTTP service for storing, retrieving, and deleting objects organized by buckets. Objects are persisted to disk.

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

| Variable   | Default  | Description                     |
|------------|----------|---------------------------------|
| `PORT`     | `8080`   | Port the server listens on      |
| `DATA_DIR` | `./data` | Directory for object storage    |

## API Reference

### PUT /objects/{bucket}/{objectID}

Store an object. Overwrites any existing object with the same ID in the bucket.

```bash
curl -X PUT -d "hello world" http://localhost:8080/objects/mybucket/greeting
```

**Response:** `201 Created`
```json
{"id": "greeting"}
```

### GET /objects/{bucket}/{objectID}

Retrieve an object.

```bash
curl http://localhost:8080/objects/mybucket/greeting
```

**Response (found):** `200 OK` with object data in the body

**Response (not found):** `404 Not Found`

### DELETE /objects/{bucket}/{objectID}

Delete an object.

```bash
curl -X DELETE http://localhost:8080/objects/mybucket/greeting
```

**Response (found):** `200 OK`

**Response (not found):** `404 Not Found`

## Development

### Build

```bash
make build
```

### Run Tests

```bash
make test
```

### Container

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
