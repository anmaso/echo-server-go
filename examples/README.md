# Echo Server Example Configurations

This directory contains example configurations demonstrating various features of the echo server.

## Usage

1. Copy the configuration files to your server's config directory:
```bash
cp -r config/* /path/to/echo-server/config/
```

2. Test the endpoints:

### Simple Response
```bash
curl http://localhost:8080/simple
```

### Delayed Response
```bash
curl http://localhost:8080/delayed/test
```

### Error Response
```bash
# Call multiple times to see error every 3rd request
curl http://localhost:8080/error-test/example
```

### Template Response
```bash
curl -H "X-Custom: test" http://localhost:8080/template/example
```

## Configuration Features Demonstrated

- Basic request/response
- Response delays
- Error injection
- Request counters
- Response templating
- Custom headers
- Multiple HTTP methods
- Regex path matching

Each configuration file includes comments explaining the features being demonstrated.