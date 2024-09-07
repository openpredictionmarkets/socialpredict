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


## 3. Development Environment

We recommend using Visual Studio Code (VsCode) with the official Golang extension for development. This extension includes linting, debugging, and other useful features. Ensure that the main Golang linter is enabled for consistent code quality.

More about this in [DEVELOPMENT](REAMDE/DEVELOPMENT/DEVELOPMENT.md)