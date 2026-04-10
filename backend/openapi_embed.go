package main

import "embed"

//go:embed docs/openapi.yaml
var openAPISpec []byte

//go:embed swagger-ui/*
var swaggerUIFS embed.FS
