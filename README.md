<a name="readme-top"></a>

<!-- PROJECT SHIELDS -->
[![Go Reference][gorefs-shield]][gorefs-url]
[![Go Report Card][goreport-shield]][goreport-url]
[![GitHub Workflow Status][build-shield]][build-url]
[![Codecov][codecov-shield]][codecov-url]
[![License][license-shield]][license-url]

<br />

<div align="center">
    <h3 align="center">SMSGate Core Library</h3>
    <p align="center">
        Foundational components for SMSGate services with Fx dependency injection
    </p>
</div>

## 📖 Table of Contents

- [📖 Table of Contents](#-table-of-contents)
- [🚀 About The Project](#-about-the-project)
  - [🔧 Built With](#-built-with)
- [🛠️ Getting Started](#️-getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
- [💻 Usage](#-usage)
  - [Fx Dependency Injection](#fx-dependency-injection)
    - [Available Modules](#available-modules)
  - [Configuration Management](#configuration-management)
  - [HTTP Server Setup](#http-server-setup)
  - [Logger Integration](#logger-integration)
  - [Redis Client](#redis-client)
  - [Data Validation](#data-validation)
- [🧩 Core Components](#-core-components)
  - [🧩 Configuration Module](#-configuration-module)
  - [🌐 HTTP Server](#-http-server)
  - [📝 Logger](#-logger)
  - [🔴 Redis Client](#-redis-client)
  - [✅ Validator](#-validator)
- [🤝 Contributing](#-contributing)
- [📄 License](#-license)

## 🚀 About The Project

The Core Library provides essential infrastructure components for the SMSGate ecosystem. It serves as the foundational layer that handles critical cross-cutting concerns across all services using Uber's Fx for dependency injection and lifecycle management.

Key value propositions:

- **Unified Configuration** - Centralized configuration management for all components
- **Consistent Logging** - Standardized logging interface with structured output
- **Redis Integration** - Optimized Redis client with connection pooling
- **Validation Framework** - Comprehensive data validation utilities
- **Fx Ready** - Built-in Fx modules for seamless dependency injection

This library enables consistent implementation of critical infrastructure patterns across the entire ecosystem.

### 🔧 Built With

- [![Go][go-shield]][go-url]
- [![Fiber][fiber-shield]][fiber-url]
- [![Redis][redis-shield]][redis-url]
- [![Zap Logger][zap-shield]][zap-url]
- [![FX][fx-shield]][fx-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## 🛠️ Getting Started

### Prerequisites

- Go 1.24.1+
- Redis 6.0+

### Installation

1. Add to your project:
```bash
go get github.com/android-sms-gateway/core
```

2. Basic Fx application setup:
```go
package main

import (
    "github.com/android-sms-gateway/core"
    "github.com/android-sms-gateway/core/config"
    "github.com/android-sms-gateway/core/http"
    "github.com/android-sms-gateway/core/logger"
    "github.com/android-sms-gateway/core/redis"
    "github.com/android-sms-gateway/core/validator"
    "go.uber.org/fx"
)

func main() {
    app := fx.New(
        // Core modules
        logger.Module,
        config.Module,
        redis.Module,
        validator.Module,
        http.Module,
        
        // Provide your configuration
        fx.Provide(func() (redis.Config, http.Config) {
            return redis.Config{}, http.Config{}
        }),
        
        // Application entrypoint
        fx.Invoke(func(lc fx.Lifecycle, log *zap.Logger) {
            lc.Append(fx.Hook{
                OnStart: func(ctx context.Context) error {
                    log.Info("SMSGate core application started")
                    return nil
                },
                OnStop: func(ctx context.Context) error {
                    log.Info("SMSGate core application stopped")
                    return nil
                },
            })
        }),
    )

    app.Run()
}
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## 💻 Usage

### Fx Dependency Injection

The library uses Uber's Fx for dependency injection. Each component provides a pre-configured Fx module that handles:

- Dependency provisioning
- Lifecycle management
- Configuration decoration

#### Available Modules

```go
import (
    "github.com/android-sms-gateway/core/http"
    "github.com/android-sms-gateway/core/logger"
    "github.com/android-sms-gateway/core/redis"
    "github.com/android-sms-gateway/core/validator"
    "go.uber.org/fx"
)

// Use the pre-built modules
fx.New(
    logger.Module,     // Provides *zap.Logger with lifecycle management
    redis.Module,      // Provides *redis.Client with connection pooling
    validator.Module,  // Provides *validator.Validate with required struct validation
    http.Module,       // Provides *fiber.App with middleware and lifecycle management
)
```

### Configuration Management

Configuration is loaded from environment variables using the `envconfig` library with Fx integration. Define your configuration struct with appropriate tags:

```go
package config

import (
    "github.com/android-sms-gateway/core/redis"
    "github.com/android-sms-gateway/core/http"
    "go.uber.org/fx"
)

// Config holds all application configuration
type Config struct {
    // Redis configuration
    Redis struct {
        URL string `envconfig:"REDIS__URL"`
    }
    
    // HTTP server configuration  
    HTTP struct {
        Address     string `envconfig:"HTTP__ADDRESS"`
        ProxyHeader string `envconfig:"HTTP__PROXY_HEADER"`
        Proxies     []string `envconfig:"HTTP__PROXIES"`
    }
    
    // Custom application configuration
    App struct {
        Name    string `envconfig:"APP__NAME" default:"SMSGate"`
        Version string `envconfig:"APP__VERSION" default:"1.0.0"`
    }
}

// Module provides configuration loading via Fx
var Module = fx.Module(
    "config",
    fx.Provide(func() (Config, error) {
        var cfg Config
        if err := Load(&cfg); err != nil {
            return nil, err
        }
        return cfg, nil
    }),
)

// Load loads configuration from environment variables
func Load[T any](c *T) error {
    // Implementation from config/config.go
}
```

Set environment variables:
```bash
# Redis configuration
export REDIS__URL="redis://localhost:6379/0"

# HTTP server configuration
export HTTP__ADDRESS=":8080"
export HTTP__PROXY_HEADER="X-Forwarded-For"
export HTTP__PROXIES="192.168.1.0/24"

# Application configuration
export APP__NAME="SMSGate"
export APP__VERSION="1.0.0"
```

### HTTP Server Setup

The HTTP server uses the Fiber framework with built-in logging, recovery middleware, and Fx lifecycle management:

```go
package main

import (
    "github.com/android-sms-gateway/core/http"
    "github.com/gofiber/fiber/v2"
    "go.uber.org/fx"
    "go.uber.org/zap"
)

// httpOptionsProvider provides HTTP server options
func httpOptionsProvider() http.Options {
    return http.Options{
        ErrorHandler: func(c *fiber.Ctx, err error) error {
            return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
                "error": err.Error(),
            })
        },
    }
}

func main() {
    app := fx.New(
        http.Module,
        
        // Provide custom HTTP options
        fx.Provide(httpOptionsProvider),
        
        // Add your routes
        fx.Invoke(func(app *fiber.App, logger *zap.Logger) {
            // Health check endpoint
            app.Get("/api/health", func(c *fiber.Ctx) error {
                return c.JSON(fiber.Map{
                    "status":  "healthy",
                    "service": "sms-gateway-core",
                })
            })

            logger.Info("Routes registered")
        }),
    )

    app.Run()
}
```

### Logger Integration

The logger module provides a Zap-based structured logger with Fx lifecycle management:

```go
package main

import (
    "errors"

    "github.com/android-sms-gateway/core/logger"
    "go.uber.org/fx"
    "go.uber.org/zap"
)

func main() {
    app := fx.New(
        logger.Module,
        
        fx.Invoke(func(logger *zap.Logger) {
            logger.Info("Application starting", 
                zap.String("env", "production"),
                zap.String("version", "1.0.0"))
            
            logger.Debug("Debug message", 
                zap.Int("count", 42),
                zap.Bool("active", true))
            
            logger.Error("Error occurred", 
                zap.Error(errors.New("demo error")),
                zap.String("operation", "database-query"))
        }),
    )

    app.Run()
}
```

### Redis Client

The Redis module provides a go-redis/v9 client with URL-based configuration and connection pooling:

```go
package main

import (
    "context"

    "github.com/android-sms-gateway/core/redis"
    "go.uber.org/fx"
    "go.uber.org/zap"
)

func main() {
    app := fx.New(
        redis.Module,
        
        fx.Invoke(func(redisClient *redis.Client, logger *zap.Logger) {
            ctx := context.Background()

            // Example Redis operations
            err := redisClient.Set(ctx, "key", "value", 0).Err()
            if err != nil {
                logger.Error("Failed to set key", zap.Error(err))
                return
            }

            val, err := redisClient.Get(ctx, "key").Result()
            if err != nil {
                logger.Error("Failed to get key", zap.Error(err))
                return
            }

            logger.Info("Redis operation successful", zap.String("value", val))
        }),
    )

    app.Run()
}
```

### Data Validation

The validator module provides go-playground/validator/v10 with required struct validation enabled:

```go
package main

import (
    "github.com/android-sms-gateway/core/validator"
    "go.uber.org/fx"
    "go.uber.org/zap"
)

// User represents a user model with validation tags
type User struct {
    Name     string `validate:"required,min=2,max=30"`
    Email    string `validate:"required,email"`
    Age      uint8  `validate:"required,gte=18"`
    Password string `validate:"required,min=8"`
}

func main() {
    app := fx.New(
        validator.Module,
        
        fx.Invoke(func(validate *validator.Validate, logger *zap.Logger) {
            // Create user instance
            user := User{
                Name:     "John",
                Email:    "john@example.com",
                Age:      25,
                Password: "secret123",
            }

            // Validate user
            if err := validate.Struct(user); err != nil {
                logger.Error("User validation failed", zap.Error(err))
                return
            }

            logger.Info("User validation successful")
        }),
    )

    app.Run()
}
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## 🧩 Core Components

### 🧩 Configuration Module

- **Package**: `github.com/android-sms-gateway/core/config`
- **Function**: `Load[T any](c *T) error` - Loads configuration from environment variables with `.env` support
- **Fx Integration**: Provides pre-configured configuration structs via Fx modules
- **Features**: Environment variable loading, default values, validation support

### 🌐 HTTP Server

- **Package**: `github.com/android-sms-gateway/core/http`
- **Function**: `New(config Config, option Options, logger *zap.Logger) (*fiber.App, error)` - Creates a new Fiber HTTP server
- **Fx Integration**: Provides `*fiber.App` instance with lifecycle management
- **Features**: Built-in logging, recovery middleware, graceful shutdown, configurable options

### 📝 Logger

- **Package**: `github.com/android-sms-gateway/core/logger`
- **Function**: `New() (*zap.Logger, error)` - Creates a new Zap logger instance
- **Fx Integration**: Provides `*zap.Logger` with lifecycle management (sync on shutdown)
- **Features**: Structured logging, development/production modes, configurable output

### 🔴 Redis Client

- **Package**: `github.com/android-sms-gateway/core/redis`
- **Function**: `New(config Config) (*redis.Client, error)` - Creates a new Redis client
- **Fx Integration**: Provides `*redis.Client` with connection pooling
- **Features**: URL-based configuration, connection pooling, built-in health checks

### ✅ Validator

- **Package**: `github.com/android-sms-gateway/core/validator`
- **Function**: `New() *validator.Validate` - Creates a new validator instance
- **Fx Integration**: Provides `*validator.Validate` with required struct validation
- **Features**: Comprehensive validation rules, cross-field validation, custom validators

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## 🤝 Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## 📄 License

Distributed under the Apache 2.0 License. See [LICENSE](LICENSE) for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- MARKDOWN LINKS & IMAGES -->
[gorefs-shield]: https://img.shields.io/badge/Go_Reference-007d9c?style=for-the-badge&logo=go&logoColor=white
[gorefs-url]: https://pkg.go.dev/github.com/android-sms-gateway/core
[goreport-shield]: https://goreportcard.com/badge/github.com/android-sms-gateway/core?style=for-the-badge
[goreport-url]: https://goreportcard.com/report/github.com/android-sms-gateway/core
[build-shield]:    https://img.shields.io/github/actions/workflow/status/android-sms-gateway/core/go.yml?branch=master&style=for-the-badge
[build-url]:       https://github.com/android-sms-gateway/core/actions/workflows/go.yml
[codecov-shield]:  https://img.shields.io/codecov/c/github/android-sms-gateway/core/master?flag=core&style=for-the-badge
[codecov-url]:     https://codecov.io/gh/android-sms-gateway/core
[license-shield]: https://img.shields.io/badge/License-Apache_2.0-blue.svg?style=for-the-badge
[license-url]: https://www.apache.org/licenses/LICENSE-2.0
[go-shield]: https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white
[go-url]: https://golang.org
[redis-shield]: https://img.shields.io/badge/Redis-DC382D?style=for-the-badge&logo=redis&logoColor=white
[redis-url]: https://redis.io
[zap-shield]: https://img.shields.io/badge/Zap-Log-000000?style=for-the-badge&logo=go&logoColor=white
[zap-url]: https://github.com/uber-go/zap
[fiber-shield]: https://img.shields.io/badge/Fiber-000000?style=for-the-badge&logo=go&logoColor=white
[fiber-url]: https://github.com/gofiber/fiber
[fx-shield]: https://img.shields.io/badge/FX-000000?style=for-the-badge&logo=go&logoColor=white
[fx-url]: https://github.com/uber-go/fx