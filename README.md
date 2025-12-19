# **Market Engine (Go)**

This project is a gRPC-based application that mimics a realtime market engine for educational purposes. It demonstrates how to structure a Go service with Protobuf/gRPC and includes a hot-reload development setup.

## **Prerequisites**

Ensure you have the following tools installed on your system:

-   **Go**: [Download Go](https://go.dev) (version 1.24 or higher recommended)
-   **Buf**: [Install Buf](https://buf.build/docs/installation) (for Protobuf code generation)
-   **Air**: [Install Air](https://github.com/air-verse/air#installation) (for live reloading)

    ```bash
    # Install Air via Go
    go install github.com/air-verse/air@latest
    ```

## **Development Setup**

Follow these steps to set up the project and start the development server.

### 1. Install Dependencies

Download the required Go modules.

```bash
go mod download
```

### 2. Generate Code

Generate the Go code from the Protobuf definitions using `buf`.

```bash
buf generate
```

### 3. Run Development Server

Start the server using `air` to enable hot reloading. The server will restart automatically when you modify code.

```bash
air
```


## **Project Structure**
-   `cmd/market-engine`: Application entry point.
-   `proto`: Protocol Buffer definitions.
-   `internal`: Core business logic and implementation.
-   `gen`: Generated Go code from Protobufs.
