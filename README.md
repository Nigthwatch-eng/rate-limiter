# Rate Limiting API with Redis and Go

This project implements a simple rate limiting mechanism using Redis and Go. The rate limiting logic is applied to a specific API endpoint, `/api/is_rate_limited/:unique_token`, which checks if a given token has exceeded the allowed number of requests in a given time window.

## Features

- **Rate Limiting**: Ensures a maximum number of requests per unique token within a specified time window.
- **Atomic Operations**: Uses Lua scripting in Redis to ensure atomic operations, making it safe for distributed systems.
- **Easy to Extend**: Built with Gin, a fast and simple web framework for Go.

## Requirements

- Go (1.16+)
- Redis (5.0+)

## Installation

### Install Go

1. **Download Go**:
    - Go to the [official Go download page](https://golang.org/dl/) and download the installer for your operating system.

2. **Install Go**:
    - Follow the instructions for your operating system to install Go.

3. **Verify installation**:
    - Open a terminal or command prompt and run:
      ```sh
      go version
      ```
    - You should see the installed Go version.

### Install Redis

1. **Run Redis locally**:

    - **On macOS**:
      ```sh
      brew install redis
      brew services start redis
      ```


### Install Project Dependencies

```
go mod tidy
```

### Usage

1. **Clone Repository**:
2. **Start Redis**:
3. Run the application
    ```
    go run server/main.go
    ```
4. CURL API Endpoint:
    ```
   curl http://localhost:8080/api/is_rate_limited/some_unique_token
    ```
    Response 
    ```
    {"rate_limited": false}
    ```
5. Running tests 
```
go test -v
```

### Screenshots from local run
1. Server

    ![Screenshot 2024-08-08 at 6 51 51 AM](https://github.com/user-attachments/assets/005cd2a5-06b6-4eed-a45e-b57f5ab25192)

2. cURL requests
   
    ![Screenshot 2024-08-08 at 6 51 40 AM](https://github.com/user-attachments/assets/0f236703-927e-43c2-864e-da91eb244d8f)
