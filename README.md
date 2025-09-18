# Yulia Lingo

[![ci Status](https://github.com/Sunagatov/Iced-Latte/actions/workflows/dev-branch-pr-deployment-pipeline.yml/badge.svg)](https://github.com/Sunagatov/Iced-Latte/actions)
[![license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/danilqa/node-file-router/blob/main/LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/zufarexplainedit/yulia-lingo-backend.svg)](https://hub.docker.com/r/zufarexplainedit/yulia-lingo-backend/)
[![GitHub issues](https://img.shields.io/github/issues/Sunagatov/Yulia-Lingo)](https://github.com/Sunagatov/Yulia-Lingo/issues)
[![GitHub stars](https://img.shields.io/github/stars/Sunagatov/Yulia-Lingo)](https://github.com/Sunagatov/Yulia-Lingo/stargazers)

**Yulia-Lingo** is a modern, secure English Learning Telegram Bot built with Go.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Tech Stack](#tech-stack)
- [Architecture](#architecture)
- [Security Features](#security-features)
- [Quick Start](#quick-start)
- [Features](#features)
- [Configuration](#configuration)
- [Development](#development)
- [Contributing](#contributing)
- [Code of Conduct](#code-of-conduct)
- [License](#license)
- [Contact](#contact)

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 13+
- Docker and Docker Compose (optional)
- Telegram Bot Token

## Tech Stack

- **Language:** Go 1.21
- **Architecture:** Clean Architecture with Dependency Injection
- **Database:** PostgreSQL with connection pooling
- **Telegram Bot API:** github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
- **Logging:** Structured JSON logging with Logrus
- **Configuration:** Environment-based configuration
- **Containerization:** Docker with multi-stage builds

## Architecture

The application follows Clean Architecture principles with clear separation of concerns:

```
cmd/app/                 # Application entry point
internal/
├── config/             # Configuration management
├── database/           # Database connection and management
├── logger/             # Structured logging
├── telegram/           # Telegram bot management
│   └── handlers/       # Message and callback handlers
├── irregular_verbs/    # Irregular verbs domain
├── my_word_list/       # Word list domain
├── translate/          # Translation service
└── util/              # Shared utilities
```

## Security Features

- **Input Validation:** All user inputs are validated and sanitized
- **SQL Injection Prevention:** Parameterized queries and prepared statements
- **SSRF Protection:** URL allowlisting for external API calls
- **Log Injection Prevention:** Structured logging with input sanitization
- **Secure Configuration:** Environment-based secrets management
- **Resource Management:** Proper connection pooling and cleanup
- **Concurrency Control:** Limited goroutine spawning to prevent resource exhaustion

## Quick Start

### Using Docker Compose (Recommended)

1. Clone the repository:
```bash
git clone https://github.com/Sunagatov/Yulia-Lingo.git
cd Yulia-Lingo
```

2. Copy and configure environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Start the application:
```bash
docker-compose up -d
```

### Manual Setup

1. Install dependencies:
```bash
go mod download
```

2. Set up PostgreSQL database and configure environment variables

3. Run the application:
```bash
go run cmd/app/main.go
```

## Features

- **Irregular Verbs Learning:** Interactive study of English irregular verbs
- **Personal Word Lists:** Create and manage custom vocabulary lists
- **Translation Service:** Real-time word translation with multiple meanings
- **Pagination:** Efficient browsing through large datasets
- **Structured Logging:** Comprehensive logging for monitoring and debugging
- **Health Checks:** Database connectivity monitoring
- **Graceful Shutdown:** Proper resource cleanup on termination

## Configuration

The application uses environment variables for configuration. See `.env.example` for all available options:

### Required Variables
- `TELEGRAM_BOT_TOKEN`: Your Telegram bot token
- `POSTGRESQL_PASSWORD`: Database password

### Optional Variables
- `POSTGRESQL_HOST`: Database host (default: localhost)
- `POSTGRESQL_PORT`: Database port (default: 5432)
- `POSTGRESQL_USER`: Database user (default: postgres)
- `POSTGRESQL_DATABASE_NAME`: Database name (default: yulia_lingo)
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `IRREGULAR_VERBS_FILE_PATH`: Path to irregular verbs Excel file

## Development

### Running Tests
```bash
go test ./...
```

### Code Quality
The project includes comprehensive security scanning and follows Go best practices:
- SOLID principles implementation
- Dependency injection
- Interface-based design
- Comprehensive error handling
- Resource leak prevention

### Building
```bash
go build -o yulia-lingo cmd/app/main.go
```

## Contributing

Interested in contributing? Read our [Contributing Guide](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## Code of Conduct

Please read our [Code of Conduct](CODE_OF_CONDUCT.md) to keep our community approachable and respectable.

## License

This project is licensed under the [MIT License](LICENSE).

## Contact

Have any questions or suggestions? Feel free to [open an issue](https://github.com/Sunagatov/Yulia-Lingo/issues) or contact us directly.

## FAQ

### How do I set up the project?
Follow the instructions in the [Quick Start](#quick-start) section above.

### Where can I find API documentation?
The bot uses Telegram Bot API. Internal API documentation is available through code comments and interfaces.

### How do I report security vulnerabilities?
Please see our [Security Policy](SECURITY.md) for reporting security issues.

### What databases are supported?
Currently, the application supports PostgreSQL 13+. The database layer is abstracted and can be extended for other databases.

## Community and Support

Join our community at https://t.me/zufarexplained for support and discussions!