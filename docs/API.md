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

## History API

The History API allows you to record, retrieve, and manage the history of requests made to the echo server and the responses it generated.

### Start Recording

*   **Endpoint:** `POST /history/start`
*   **Description:** Starts recording request and response history. Optionally, you can set the maximum number of history entries to store by providing `maxSize` in the request body.
*   **Request Body (Optional JSON):**
    ```json
    {
        "maxSize": 100
    }
    ```
*   **Response Body (JSON):**
    Indicates the current recording status and maximum size.
    ```json
    {
        "recordingActive": true,
        "maxSize": 100
    }
    ```

### Stop Recording

*   **Endpoint:** `POST /history/stop`
*   **Description:** Stops recording request and response history. Entries already recorded are retained until cleared or evicted.
*   **Response Body (JSON):**
    Indicates the current recording status and maximum size.
    ```json
    {
        "recordingActive": false,
        "maxSize": 100
    }
    ```

### Configure History

*   **Endpoint:** `PUT /history/config`
*   **Description:** Configures the history settings, primarily the maximum number of entries to store. Setting `maxSize` to 0 will stop new entries from being stored (if recording is active) and clear existing entries.
*   **Request Body (JSON):**
    ```json
    {
        "maxSize": 200
    }
    ```
*   **Response Body (JSON):**
    Indicates the current recording status and updated maximum size.
    ```json
    {
        "recordingActive": true,
        "maxSize": 200
    }
    ```

### Retrieve History

*   **Endpoint:** `GET /history`
*   **Description:** Retrieves all currently recorded history entries.
*   **Response Body (JSON Array of HistoryEntry):**
    Each entry includes a timestamp, the original request data, a summary of the response, and the response size.
    ```json
    [
        {
            "timestamp": "2023-10-27T10:30:00Z",
            "request": {
                "method": "GET",
                "path": "/api/data",
                "queryParams": {"id": ["123"]},
                "headers": {"User-Agent": "TestClient"},
                "body": "",
                "remoteAddr": "127.0.0.1:12345"
            },
            "response": {
                "statusCode": 200,
                "headers": {"Content-Type": "application/json"},
                "body": "{\"message\": \"success\"}"
            },
            "responseSize": 22
        }
    ]
    ```
    *Note: The `request` object structure shown is illustrative; the actual fields are defined by `model.RequestData` and `response` by `model.ResponseSummary`.*

### Clear History

*   **Endpoint:** `DELETE /history`
*   **Description:** Deletes all currently recorded history entries.
*   **Response:**
    *   `204 No Content` on successful clearing.

## Error Codes

- 200: Success
- 201: Created (new configuration)
- 400: Bad Request
- 404: Not Found
- 405: Method Not Allowed
- 500: Internal Server Error
```

You can now mark "Create usage documentation" and "Document API endpoints" as completed in `TASKS.md` and move on to "Create example configurations".