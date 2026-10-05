# 📱 SMSGate Core

[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stars][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![License][license-shield]][license-url]

Shared Go library for SMSGate backend services: configuration, logging, Redis, HTTP, and validation as composable Uber Fx modules.

## 📖 About

SMSGate Core is the shared Go foundation for SMSGate services. It provides environment-driven configuration, structured logging, Redis clients, Fiber HTTP servers, and request validation — each packaged as a self-contained Uber Fx module. Services import only the modules they need. Full ecosystem docs: https://docs.sms-gate.app/.

## 📚 Table of Contents

- [📖 About](#-about)
- [📚 Table of Contents](#-table-of-contents)
- [⭐ Features](#-features)
- [📦 Prerequisites](#-prerequisites)
- [🚀 Quickstart](#-quickstart)
- [⚙️ Configuration](#️-configuration)
- [🔌 API Overview](#-api-overview)
- [📚 Documentation](#-documentation)
- [🤝 Contributing](#-contributing)
- [⚖️ License](#️-license)

## ⭐ Features

- Uber Fx modules: `logger`, `redis`, `validator`, `http`
- envconfig + godotenv configuration loader
- Zap structured logging (development mode via `DEBUG`)
- Fiber v2 HTTP server with request logging, recovery, trusted proxies, error handlers
- go-redis/v9 client from a single URL
- go-playground/validator/v10 with required-struct validation
- Reusable JSON/HTML error handlers and response `jsonify` middleware

## 📦 Prerequisites

- Go 1.25+ (see [go.mod](go.mod))
- Redis 6+ when using the redis module
- No other runtime dependencies

## 🚀 Quickstart

```bash
go get github.com/android-sms-gateway/core
```

Wire the modules into an Fx application:

```go
package main

import (
	"github.com/android-sms-gateway/core/http"
	"github.com/android-sms-gateway/core/logger"
	"github.com/android-sms-gateway/core/redis"
	"github.com/android-sms-gateway/core/validator"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		logger.Module,
		redis.Module,
		validator.Module,
		http.Module,
	).Run()
}
```

## ⚙️ Configuration

[config.Load](config/config.go) reads environment variables via envconfig and an optional `.env` file via godotenv. Field names map to env vars (no prefix):

```bash
export ADDRESS=":8080"                # http.Config.Address
export PROXYHEADER="X-Forwarded-For"  # http.Config.ProxyHeader
export PROXIES="10.0.0.0/8"           # http.Config.Proxies
export URL="redis://localhost:6379/0" # redis.Config.URL
```

Logger config: set `DEBUG=1` for development mode. See [logger/logger.go](logger/logger.go).

## 🔌 API Overview

| Package | Provides |
| ------- | -------- |
| [config](config/config.go) | `Load[T]` env/.env loader |
| [handler](handler/handler.go) | `Base` body parser + validator helper |
| [http](http/http.go) | `*fiber.App` module, error handlers, jsonify |
| [logger](logger/logger.go) | `*zap.Logger` module |
| [redis](redis/redis.go) | `*redis.Client` module |
| [validator](validator/validator.go) | `*validator.Validate` module |

Each package exposes an `fx.Module`; import paths mirror the table above.

## 📚 Documentation

- [SMSGate docs](https://docs.sms-gate.app/)
- [API integration](https://docs.sms-gate.app/integration/api/)
- [Client libraries](https://docs.sms-gate.app/integration/client-libraries/)

## 🤝 Contributing

Open an issue first, then submit a PR. Follow the [README style guide](https://docs.sms-gate.app/). Run `make lint` and `make test` locally.

## ⚖️ License

Apache-2.0. See [LICENSE](LICENSE).

[contributors-shield]: https://img.shields.io/github/contributors/android-sms-gateway/core?style=for-the-badge
[contributors-url]: https://github.com/android-sms-gateway/core/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/android-sms-gateway/core?style=for-the-badge
[forks-url]: https://github.com/android-sms-gateway/core/network/members
[stars-shield]: https://img.shields.io/github/stars/android-sms-gateway/core?style=for-the-badge
[stars-url]: https://github.com/android-sms-gateway/core/stargazers
[issues-shield]: https://img.shields.io/github/issues/android-sms-gateway/core?style=for-the-badge
[issues-url]: https://github.com/android-sms-gateway/core/issues
[license-shield]: https://img.shields.io/github/license/android-sms-gateway/core?style=for-the-badge
[license-url]: https://github.com/android-sms-gateway/core/blob/master/LICENSE
