# Echo HTTP Server Implementation Tasks

This document outlines the tasks required to implement an echo HTTP server with extended configuration options.

## 1. Project Setup
- [x] Initialize Go module
- [x] Create project structure
- [x] Set up logging framework
- [x] Create configuration structures

## 2. Core Server Implementation
- [x] Implement basic HTTP server setup
- [x] Create handler function for incoming requests
- [x] Implement request data extraction (method, URL, params, headers, body)
- [x] Implement default response behavior (echo request data)
- [x] Set up server-side request logging

## 3. Request Counter Implementation
- [x] Create a global request counter with mutex protection
- [x] Implement counter increment for each request
- [x] Create path-specific counters with mutex protection
- [x] Implement per-path counter reset functionality
  - [x] fix: the /counter endpoint was not being used
  - [x] fix: the calls to /counter endpoit should also increase counters 
  - [x] feat: include the counter data for the path in the responses

## 4. Path Configuration System
- [x] Design configuration data structures
- [x] Implement configuration loading mechanism and include tests
- [x] Create path matcher using regex support
- [x] Implement configuration lookup for incoming requests

## 5. Response Customization Features
- [x] Implement custom status code responses
- [x] Implement response delay functionality
- [x] Implement custom response body support
- [x] Implement error response configuration. Returning an error means returning the Status code as the configured value for error
- [x] Implement "error every" functionality. When the counter for the request is divisible by errorEvery return the error configuration

## 6. Thread Safety & Concurrency
- [x] Ensure thread-safe access to counters with proper mutex usage
- [x] Handle concurrent configuration access
- [x] Ensure thread safety in path matching and lookup

## 7. Configuration Management
- [x] Implement API endpoints for adding/updating path configurations
- [x] Implement API endpoints for viewing current configurations
- [x] Implement API endpoints for viewing counter values

## 8. Testing
- [x] Write unit tests for core functionality
- [x] Create integration tests for request/response behavior
- [x] Test concurrency handling
- [x] Test regex path matching
- [x] Test error frequency functionality
- [x] Test delayed response functionality

## 9. Documentation
- [x] Create usage documentation
- [x] Document API endpoints
- [x] Create example configurations
- [x] Document project structure

## 10. Final Integration and Review
- [ ] Ensure all features are working together
- [ ] Performance review
- [ ] Code cleanup
- [ ] Final testing