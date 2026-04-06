# Simple Object Storage Service

A lightweight RESTful API service for storing, fetching, and managing objects in buckets stored on disk.
De-duplicates objects on a per-bucket basis.

## Installation
To run and develop the service locally (Golang version):

1. **Clone the repository:**
   ```bash
   git clone https://github.com/your-username/rhobs-challenge.git
   cd rhobs-challenge
   ```

2. **Install Go (if not installed):**
   Follow the instructions at https://golang.org/dl/ to install Go.

3. **Build the application:**
   ```bash
   go build -o rhobs-server .
   ```

4. **Run the application:**
   ```bash
   ./rhobs-server
   ```
   The service will be available at `http://localhost:8000` (or your configured port).

5. **Running tests:**
   ```bash
   go test ./...
   ```

You may need to adjust filenames or commands if your entry point or dependencies differ.


## API Reference

### Object Operations

#### `PUT /objects/{bucket}/{objectID}`
Store or overwrite an object in `{bucket}` with the given `{key}`.

- **Request Body:** Raw content of the object (binary or text)
- **Response:** 
  - `201 Created` if stored successfully

#### GET /objects/{bucket}/{key}
Retrieve the object content from `{bucket}` identified by `{key}`.

- **Response:**
  - `200 OK` with raw object content in response body
  - `400 Not Found` if the object does not exist

#### DELETE /objects/{bucket}/{key}
Delete the object identified by `{key}` from `{bucket}`

- **Response:**
  - `200 OK` if object deleted successfully
  - `400 Not Found` if the object does not exist

