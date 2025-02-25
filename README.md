# ORBAT - Order of Battle Management System

ORBAT is a web application for managing military units, their equipment, and organizational structure.

## Project Structure

The project follows a standard Go project layout:

```
ORBAT/
├── cmd/                  # Command-line applications
│   └── orbat/            # Main application entry point
├── internal/             # Private application code
│   ├── database/         # Database operations
│   ├── handlers/         # HTTP handlers
│   ├── models/           # Data models
│   └── storage/          # Storage operations (Google Cloud Storage)
├── templates/            # HTML templates
├── SQL/                  # SQL scripts and migrations
├── .env                  # Environment variables
├── go.mod                # Go module definition
└── go.sum                # Go module checksums
```

## Features

- Manage military groups, teams, and members
- Track weapons and their usage across different units
- Manage vehicles and their crew
- View statistics by country
- Upload and manage images for weapons and vehicles

## Development

### Prerequisites

- Go 1.16 or higher
- SQLite or Turso database
- Google Cloud Storage account (for image storage)

### Environment Variables

Create a `.env` file with the following variables:

```
DATABASE_URL=libsql://your-database-url
GCS_BUCKET_NAME=your-bucket-name
PORT=8080
```

### Running the Application

```bash
go run cmd/orbat/main.go
```

### Building the Application

```bash
go build -o orbat cmd/orbat/main.go
```

## Deployment

The application can be deployed to Google Cloud Run or any other platform that supports Go applications.

### Docker

A Dockerfile is provided for containerized deployment:

```bash
docker build -t orbat .
docker run -p 8080:8080 orbat
```

## License

This project is licensed under the MIT License - see the LICENSE file for details. 