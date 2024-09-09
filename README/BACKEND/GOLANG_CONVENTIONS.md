# Golang Backend Conventions

## 1. Code Style

Refer to the [Golang Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) for the standard coding style. This document provides detailed guidelines on best practices for writing Go code.

## 2. Test Structure

All tests should be placed in the same directory as the code they are testing. Each test file should follow the convention of ending with `_test.go`.

Example:

```
handlers/

    user_handler.go
    user_handler_test.go
```

## 3. Data Conventions

* The entire application should be as stateless as possible, meaning we have a one-way writeable database of users, markets, bets and calculate all relevant states of the application from as few possible columns within those models.
* We should separate the data logic from the business logic as much as possible with a Domain Driven Design (DDD), meaning we have a repository/ directory which is designed to be the central location to keep functions that extract data from the databases. This should ideally help slow the growth of the codebase over time and keep data extraction more testable, which should make our startless architecture more reliable.

## 4. Development Environment

We recommend using Visual Studio Code (VsCode) with the official Golang extension for development. This extension includes linting, debugging, and other useful features. Ensure that the main Golang linter is enabled for consistent code quality.

More about this in [DEVELOPMENT](REAMDE/DEVELOPMENT/DEVELOPMENT.md)