# Golang Backend Conventions

## 1. Code Style

Refer to the [Golang Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) for the standard coding style. This document provides detailed guidelines on best practices for writing Go code.

## 2. Test Structure & Conventions

All tests should be placed in the same directory as the code they are testing. Each test file should follow the convention of ending with `_test.go`.

Example:

```
handlers/

    user_handler.go
    user_handler_test.go
```

Tests don't have to be written for private functions, but they MUST be written for all public functions.

### Organization of Tests
Tests in Go are typically placed in the same package as the code they test. This practice ensures that tests have appropriate access to internal variables and functions necessary for thorough testing. 

### Exception for Test Utilities
An exception exists for test utilities intended to be shared across multiple packages. For example, a custom fake database designed for testing should be placed in a separate package. This allows it to be imported and reused in various test scenarios throughout the application. This approach is beneficial when the utility (e.g., a fake database) is not part of the main functionality but is essential for testing components that interact with the database.


## 3. Data Conventions

* The entire application should be as stateless as possible, meaning we have a one-way writeable database of users, markets, bets and calculate all relevant states of the application from as few possible columns within those models.
* We should separate the data logic from the business logic as much as possible with a Domain Driven Design (DDD), meaning we have a repository/ directory which is designed to be the central location to keep functions that extract data from the databases. This should ideally help slow the growth of the codebase over time and keep data extraction more testable, which should make our stateless architecture more reliable.

## 4. Development Environment

We recommend using Visual Studio Code (VsCode) with the official Golang extension for development. This extension includes linting, debugging, and other useful features. Ensure that the main Golang linter is enabled for consistent code quality.

More about this in [DEVELOPMENT](REAMDE/DEVELOPMENT/DEVELOPMENT.md)