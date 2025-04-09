# Echo Server API Documentation

## Configuration Endpoints

### Get All Configurations
```http
GET /config/paths
```

Response:
```json
{
    "server": {
        "host": "localhost",
        "port": 8080,
        "readTimeout": "30s",
        "writeTimeout": "30s"
    },
    "paths": [
        {
            "pattern": "^/api/.*",
            "methods": ["GET", "POST"],
            "response": {
                "statusCode": 200,
                "body": "{\"status\":\"ok\"}"
            }
        }
    ]
}
```

### Add Path Configuration
```http
POST /config/paths
Content-Type: application/json

{
    "pattern": "^/test/.*",
    "methods": ["GET"],
    "response": {
        "statusCode": 200,
        "body": "{\"status\":\"ok\"}"
    }
}
```

### Update Path Configuration
```http
PUT /config/paths/test
Content-Type: application/json

{
    "methods": ["GET", "POST"],
    "response": {
        "statusCode": 200,
        "body": "{\"status\":\"updated\"}"
    }
}
```

## Counter Endpoints

### Get All Counters
```http
GET /counter
```

Response:
```json
{
    "global": 100,
    "paths": {
        "/api/test": 50,
        "/api/other": 25
    }
}
```

### Reset Path Counter
```http
DELETE /counter/api/test
```

### Reset All Counters
```http
DELETE /counter
```

## Error Codes

- 200: Success
- 201: Created (new configuration)
- 400: Bad Request
- 404: Not Found
- 405: Method Not Allowed
- 500: Internal Server Error
```

You can now mark "Create usage documentation" and "Document API endpoints" as completed in `TASKS.md` and move on to "Create example configurations".// filepath: /Users/angelluismarinsoler/work/workspaces/go/echo-server/docs/API.md
# Echo Server API Documentation

## Configuration Endpoints

### Get All Configurations
```http
GET /config/paths
```

Response:
```json
{
    "server": {
        "host": "localhost",
        "port": 8080,
        "readTimeout": "30s",
        "writeTimeout": "30s"
    },
    "paths": [
        {
            "pattern": "^/api/.*",
            "methods": ["GET", "POST"],
            "response": {
                "statusCode": 200,
                "body": "{\"status\":\"ok\"}"
            }
        }
    ]
}
```

### Add Path Configuration
```http
POST /config/paths
Content-Type: application/json

{
    "pattern": "^/test/.*",
    "methods": ["GET"],
    "response": {
        "statusCode": 200,
        "body": "{\"status\":\"ok\"}"
    }
}
```

### Update Path Configuration
```http
PUT /config/paths/test
Content-Type: application/json

{
    "methods": ["GET", "POST"],
    "response": {
        "statusCode": 200,
        "body": "{\"status\":\"updated\"}"
    }
}
```

## Counter Endpoints

### Get All Counters
```http
GET /counter
```

Response:
```json
{
    "global": 100,
    "paths": {
        "/api/test": 50,
        "/api/other": 25
    }
}
```

### Reset Path Counter
```http
DELETE /counter/api/test
```

### Reset All Counters
```http
DELETE /counter
```

## Error Codes

- 200: Success
- 201: Created (new configuration)
- 400: Bad Request
- 404: Not Found
- 405: Method Not Allowed
- 500: Internal Server Error
```

You can now mark "Create usage documentation" and "Document API endpoints" as completed in `TASKS.md` and move on to "Create example configurations".