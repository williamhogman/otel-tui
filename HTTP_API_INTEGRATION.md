# otel-tui HTTP API Integration Guide

This document provides comprehensive integration information for the otel-tui HTTP API, including endpoints, response schemas, and Zod validation schemas for TypeScript/JavaScript frontends.

## Getting Started

### Configuration

The HTTP API server is disabled by default. To enable it, use the `--http-api-port` flag:

```bash
otel-tui --http-api-port 8000
```

The default port is `8000`. Set it to `0` to disable the HTTP API server.

### Base URL

```
http://localhost:8000/api
```

### CORS

The API supports CORS and allows requests from any origin with the following headers:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type`

## API Endpoints Overview

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/traces` | GET | Get all traces with optional service filter |
| `/api/traces/{traceID}` | GET | Get all spans for a specific trace |
| `/api/traces/{traceID}/services/{service}` | GET | Get spans for a specific trace and service |
| `/api/spans/{spanID}` | GET | Get a specific span by ID |
| `/api/metrics` | GET | Get all metrics with optional filters |
| `/api/metrics/{service}` | GET | Get metrics for a specific service |
| `/api/metrics/{service}/{metricName}` | GET | Get specific metric by service and name |
| `/api/logs` | GET | Get all logs with optional filter |
| `/api/logs/trace/{traceID}` | GET | Get logs for a specific trace |
| `/api/topology` | GET | Get service dependency topology |
| `/api/services` | GET | Get list of all services |
| `/api/stats` | GET | Get store statistics |

---

## Zod Schemas

Below are Zod schemas for TypeScript/JavaScript validation. Install Zod:

```bash
npm install zod
```

### Common Schemas

```typescript
import { z } from 'zod';

// Attributes schema (key-value pairs)
const AttributesSchema = z.record(z.string(), z.any());

// Span Status
const SpanStatusSchema = z.object({
  code: z.string(), // "Unset" | "Ok" | "Error"
  message: z.string().optional(),
});

// Span Event
const SpanEventSchema = z.object({
  name: z.string(),
  timeUnixNano: z.number(),
  attributes: AttributesSchema,
  droppedAttributesCount: z.number(),
});

// Span Link
const SpanLinkSchema = z.object({
  traceId: z.string(),
  spanId: z.string(),
  traceState: z.string().optional(),
  attributes: AttributesSchema,
  droppedAttributesCount: z.number(),
});

// Span
const SpanSchema = z.object({
  traceId: z.string(),
  spanId: z.string(),
  parentSpanId: z.string().optional(),
  name: z.string(),
  kind: z.string(), // "Unspecified" | "Internal" | "Server" | "Client" | "Producer" | "Consumer"
  startTimeUnixNano: z.number(),
  endTimeUnixNano: z.number(),
  durationNano: z.number(),
  durationText: z.string(),
  attributes: AttributesSchema,
  status: SpanStatusSchema,
  events: z.array(SpanEventSchema),
  links: z.array(SpanLinkSchema),
  serviceName: z.string(),
  resourceAttributes: AttributesSchema,
  scopeName: z.string(),
  scopeVersion: z.string(),
  receivedAt: z.string().datetime(),
});

// Trace (complete trace with all spans)
const TraceSchema = z.object({
  traceId: z.string(),
  spans: z.array(SpanSchema),
  services: z.array(z.string()),
});

// Quantile (for Summary metrics)
const QuantileSchema = z.object({
  quantile: z.number(),
  value: z.number(),
});

// Data Point (generic metric data point)
const DataPointSchema = z.object({
  attributes: AttributesSchema,
  startTimeUnixNano: z.number().optional(),
  timeUnixNano: z.number(),
  // For Gauge and Sum
  value: z.number().optional(),
  // For Histogram
  count: z.number().optional(),
  sum: z.number().optional(),
  bucketCounts: z.array(z.number()).optional(),
  explicitBounds: z.array(z.number()).optional(),
  min: z.number().optional(),
  max: z.number().optional(),
  // For Summary
  quantileValues: z.array(QuantileSchema).optional(),
  // Flags
  flags: z.number(),
});

// Metric
const MetricSchema = z.object({
  name: z.string(),
  description: z.string().optional(),
  unit: z.string().optional(),
  type: z.string(), // "Gauge" | "Sum" | "Histogram" | "ExponentialHistogram" | "Summary"
  dataPoints: z.array(DataPointSchema),
  serviceName: z.string(),
  resourceAttributes: AttributesSchema,
  scopeName: z.string(),
  scopeVersion: z.string(),
  receivedAt: z.string().datetime(),
});

// Log
const LogSchema = z.object({
  timeUnixNano: z.number(),
  observedTimeUnixNano: z.number(),
  severityNumber: z.number(),
  severityText: z.string(),
  body: z.string(),
  attributes: AttributesSchema,
  traceId: z.string().optional(),
  spanId: z.string().optional(),
  flags: z.number(),
  serviceName: z.string(),
  resourceAttributes: AttributesSchema,
  scopeName: z.string(),
  scopeVersion: z.string(),
  receivedAt: z.string().datetime(),
});

// Topology Node
const TopologyNodeSchema = z.object({
  service: z.string(),
  depth: z.number(),
});

// Topology Edge
const TopologyEdgeSchema = z.object({
  source: z.string(),
  target: z.string(),
  count: z.number(),
});

// Topology
const TopologySchema = z.object({
  nodes: z.array(TopologyNodeSchema),
  edges: z.array(TopologyEdgeSchema),
});

// Stats
const StatsSchema = z.object({
  spanCount: z.number(),
  metricCount: z.number(),
  logCount: z.number(),
  traceCount: z.number(),
  serviceCount: z.number(),
  lastUpdated: z.string().datetime(),
  maxServiceSpanCount: z.number(),
  maxMetricCount: z.number(),
  maxLogCount: z.number(),
});

// Error Response
const ErrorSchema = z.object({
  error: z.string(),
});
```

---

## Endpoint Details

### 1. Get All Traces

**Endpoint:** `GET /api/traces`

**Query Parameters:**
- `service` (optional): Filter traces by service name

**Description:** Returns all spans in the store. If a service filter is provided, only spans matching that service will be returned.

**Response:** Array of Span objects

**Zod Schema:**
```typescript
const GetTracesResponseSchema = z.array(SpanSchema);
```

**Example Request:**
```bash
curl "http://localhost:8000/api/traces"
curl "http://localhost:8000/api/traces?service=frontend"
```

**Example Response:**
```json
[
  {
    "traceId": "1234567890abcdef",
    "spanId": "abcdef123456",
    "parentSpanId": "",
    "name": "GET /api/users",
    "kind": "Server",
    "startTimeUnixNano": 1699564800000000000,
    "endTimeUnixNano": 1699564800500000000,
    "durationNano": 500000000,
    "durationText": "500ms",
    "attributes": {
      "http.method": "GET",
      "http.url": "/api/users"
    },
    "status": {
      "code": "Ok"
    },
    "events": [],
    "links": [],
    "serviceName": "frontend",
    "resourceAttributes": {
      "service.name": "frontend"
    },
    "scopeName": "http",
    "scopeVersion": "1.0",
    "receivedAt": "2023-11-10T00:00:00Z"
  }
]
```

---

### 2. Get Trace by ID

**Endpoint:** `GET /api/traces/{traceID}`

**Path Parameters:**
- `traceID`: The trace ID

**Description:** Returns all spans for a specific trace, organized with service information.

**Response:** Trace object with spans and service list

**Zod Schema:**
```typescript
const GetTraceByIDResponseSchema = TraceSchema;
```

**Example Request:**
```bash
curl "http://localhost:8000/api/traces/1234567890abcdef"
```

**Example Response:**
```json
{
  "traceId": "1234567890abcdef",
  "spans": [
    {
      "traceId": "1234567890abcdef",
      "spanId": "abcdef123456",
      "name": "GET /api/users",
      "serviceName": "frontend",
      ...
    },
    {
      "traceId": "1234567890abcdef",
      "spanId": "fedcba654321",
      "parentSpanId": "abcdef123456",
      "name": "SELECT users",
      "serviceName": "database",
      ...
    }
  ],
  "services": ["frontend", "database"]
}
```

**Error Response (404):**
```json
{
  "error": "Trace not found"
}
```

---

### 3. Get Trace by ID and Service

**Endpoint:** `GET /api/traces/{traceID}/services/{service}`

**Path Parameters:**
- `traceID`: The trace ID
- `service`: The service name

**Description:** Returns spans for a specific trace and service combination.

**Response:** Array of Span objects

**Zod Schema:**
```typescript
const GetTraceByIDAndServiceResponseSchema = z.array(SpanSchema);
```

**Example Request:**
```bash
curl "http://localhost:8000/api/traces/1234567890abcdef/services/frontend"
```

**Error Response (404):**
```json
{
  "error": "Spans not found for trace and service"
}
```

---

### 4. Get Span by ID

**Endpoint:** `GET /api/spans/{spanID}`

**Path Parameters:**
- `spanID`: The span ID

**Description:** Returns a single span by its ID.

**Response:** Span object

**Zod Schema:**
```typescript
const GetSpanByIDResponseSchema = SpanSchema;
```

**Example Request:**
```bash
curl "http://localhost:8000/api/spans/abcdef123456"
```

**Error Response (404):**
```json
{
  "error": "Span not found"
}
```

---

### 5. Get All Metrics

**Endpoint:** `GET /api/metrics`

**Query Parameters:**
- `service` (optional): Filter by service name
- `metric` (optional): Filter by metric name

**Description:** Returns all metrics in the store with optional filtering.

**Response:** Array of Metric objects

**Zod Schema:**
```typescript
const GetMetricsResponseSchema = z.array(MetricSchema);
```

**Example Request:**
```bash
curl "http://localhost:8000/api/metrics"
curl "http://localhost:8000/api/metrics?service=frontend"
curl "http://localhost:8000/api/metrics?service=frontend&metric=http_requests_total"
```

**Example Response:**
```json
[
  {
    "name": "http_requests_total",
    "description": "Total number of HTTP requests",
    "unit": "1",
    "type": "Sum",
    "dataPoints": [
      {
        "attributes": {
          "method": "GET",
          "status": "200"
        },
        "startTimeUnixNano": 1699564800000000000,
        "timeUnixNano": 1699564810000000000,
        "value": 150,
        "flags": 0
      }
    ],
    "serviceName": "frontend",
    "resourceAttributes": {
      "service.name": "frontend"
    },
    "scopeName": "http",
    "scopeVersion": "1.0",
    "receivedAt": "2023-11-10T00:00:10Z"
  }
]
```

---

### 6. Get Metrics by Service

**Endpoint:** `GET /api/metrics/{service}`

**Path Parameters:**
- `service`: The service name

**Description:** Returns all metrics for a specific service.

**Response:** Array of Metric objects

**Zod Schema:**
```typescript
const GetMetricsByServiceResponseSchema = z.array(MetricSchema);
```

**Example Request:**
```bash
curl "http://localhost:8000/api/metrics/frontend"
```

---

### 7. Get Metrics by Service and Name

**Endpoint:** `GET /api/metrics/{service}/{metricName}`

**Path Parameters:**
- `service`: The service name
- `metricName`: The metric name

**Description:** Returns a specific metric by service and metric name.

**Response:** Array of Metric objects (multiple data points over time)

**Zod Schema:**
```typescript
const GetMetricsByServiceAndNameResponseSchema = z.array(MetricSchema);
```

**Example Request:**
```bash
curl "http://localhost:8000/api/metrics/frontend/http_requests_total"
```

**Error Response (404):**
```json
{
  "error": "Metrics not found for service and metric name"
}
```

---

### 8. Get All Logs

**Endpoint:** `GET /api/logs`

**Query Parameters:**
- `filter` (optional): Filter logs by service name or log content

**Description:** Returns all logs in the store with optional filtering.

**Response:** Array of Log objects

**Zod Schema:**
```typescript
const GetLogsResponseSchema = z.array(LogSchema);
```

**Example Request:**
```bash
curl "http://localhost:8000/api/logs"
curl "http://localhost:8000/api/logs?filter=error"
```

**Example Response:**
```json
[
  {
    "timeUnixNano": 1699564800000000000,
    "observedTimeUnixNano": 1699564800001000000,
    "severityNumber": 17,
    "severityText": "Error",
    "body": "Failed to connect to database",
    "attributes": {
      "error.type": "ConnectionError"
    },
    "traceId": "1234567890abcdef",
    "spanId": "fedcba654321",
    "flags": 0,
    "serviceName": "backend",
    "resourceAttributes": {
      "service.name": "backend"
    },
    "scopeName": "app",
    "scopeVersion": "1.0",
    "receivedAt": "2023-11-10T00:00:00Z"
  }
]
```

---

### 9. Get Logs by Trace ID

**Endpoint:** `GET /api/logs/trace/{traceID}`

**Path Parameters:**
- `traceID`: The trace ID

**Description:** Returns all logs associated with a specific trace ID.

**Response:** Array of Log objects

**Zod Schema:**
```typescript
const GetLogsByTraceIDResponseSchema = z.array(LogSchema);
```

**Example Request:**
```bash
curl "http://localhost:8000/api/logs/trace/1234567890abcdef"
```

**Error Response (404):**
```json
{
  "error": "Logs not found for trace"
}
```

---

### 10. Get Service Topology

**Endpoint:** `GET /api/topology`

**Description:** Returns the service dependency topology/graph showing which services call other services.

**Response:** Topology object with nodes and edges

**Zod Schema:**
```typescript
const GetTopologyResponseSchema = TopologySchema;
```

**Example Request:**
```bash
curl "http://localhost:8000/api/topology"
```

**Example Response:**
```json
{
  "nodes": [
    {
      "service": "frontend",
      "depth": 0
    },
    {
      "service": "backend",
      "depth": 0
    },
    {
      "service": "database",
      "depth": 0
    }
  ],
  "edges": [
    {
      "source": "frontend",
      "target": "backend",
      "count": 42
    },
    {
      "source": "backend",
      "target": "database",
      "count": 38
    }
  ]
}
```

**Use Case:** This endpoint is perfect for visualizing service dependencies in a graph or network diagram.

---

### 11. Get Services List

**Endpoint:** `GET /api/services`

**Description:** Returns a list of all unique service names found in the collected telemetry data.

**Response:** Array of service name strings

**Zod Schema:**
```typescript
const GetServicesResponseSchema = z.array(z.string());
```

**Example Request:**
```bash
curl "http://localhost:8000/api/services"
```

**Example Response:**
```json
[
  "frontend",
  "backend",
  "database",
  "cache"
]
```

---

### 12. Get Store Statistics

**Endpoint:** `GET /api/stats`

**Description:** Returns statistics about the current state of the telemetry store, including counts and capacity limits.

**Response:** Stats object

**Zod Schema:**
```typescript
const GetStatsResponseSchema = StatsSchema;
```

**Example Request:**
```bash
curl "http://localhost:8000/api/stats"
```

**Example Response:**
```json
{
  "spanCount": 856,
  "metricCount": 2341,
  "logCount": 432,
  "traceCount": 124,
  "serviceCount": 4,
  "lastUpdated": "2023-11-10T00:05:23Z",
  "maxServiceSpanCount": 1000,
  "maxMetricCount": 3000,
  "maxLogCount": 1000
}
```

**Notes:**
- `maxServiceSpanCount`, `maxMetricCount`, and `maxLogCount` represent the circular buffer sizes
- When these limits are reached, the oldest data is automatically rotated out
- `lastUpdated` shows when the store was last modified

---

## Data Capacity and Rotation

The otel-tui store has the following capacity limits:

- **Spans**: 1,000 (oldest spans removed when limit reached)
- **Metrics**: 3,000 (oldest metrics removed when limit reached)
- **Logs**: 1,000 (oldest logs removed when limit reached)

This circular buffer approach ensures memory usage remains bounded while keeping the most recent telemetry data.

---

## TypeScript Integration Example

Here's a complete example of integrating the API with TypeScript:

```typescript
import { z } from 'zod';

// Import all schemas (defined above)

class OtelTuiClient {
  private baseUrl: string;

  constructor(baseUrl: string = 'http://localhost:8000/api') {
    this.baseUrl = baseUrl;
  }

  async getTraces(service?: string): Promise<z.infer<typeof SpanSchema>[]> {
    const url = service
      ? `${this.baseUrl}/traces?service=${encodeURIComponent(service)}`
      : `${this.baseUrl}/traces`;

    const response = await fetch(url);
    const data = await response.json();

    return z.array(SpanSchema).parse(data);
  }

  async getTrace(traceId: string): Promise<z.infer<typeof TraceSchema>> {
    const response = await fetch(`${this.baseUrl}/traces/${traceId}`);

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Failed to fetch trace');
    }

    const data = await response.json();
    return TraceSchema.parse(data);
  }

  async getMetrics(service?: string, metric?: string): Promise<z.infer<typeof MetricSchema>[]> {
    let url = `${this.baseUrl}/metrics`;
    const params = new URLSearchParams();

    if (service) params.append('service', service);
    if (metric) params.append('metric', metric);

    if (params.toString()) {
      url += `?${params.toString()}`;
    }

    const response = await fetch(url);
    const data = await response.json();

    return z.array(MetricSchema).parse(data);
  }

  async getLogs(filter?: string): Promise<z.infer<typeof LogSchema>[]> {
    const url = filter
      ? `${this.baseUrl}/logs?filter=${encodeURIComponent(filter)}`
      : `${this.baseUrl}/logs`;

    const response = await fetch(url);
    const data = await response.json();

    return z.array(LogSchema).parse(data);
  }

  async getTopology(): Promise<z.infer<typeof TopologySchema>> {
    const response = await fetch(`${this.baseUrl}/topology`);
    const data = await response.json();

    return TopologySchema.parse(data);
  }

  async getServices(): Promise<string[]> {
    const response = await fetch(`${this.baseUrl}/services`);
    const data = await response.json();

    return z.array(z.string()).parse(data);
  }

  async getStats(): Promise<z.infer<typeof StatsSchema>> {
    const response = await fetch(`${this.baseUrl}/stats`);
    const data = await response.json();

    return StatsSchema.parse(data);
  }
}

// Usage
const client = new OtelTuiClient();

// Get all traces
const traces = await client.getTraces();

// Get traces for a specific service
const frontendTraces = await client.getTraces('frontend');

// Get a specific trace
const trace = await client.getTrace('1234567890abcdef');

// Get topology
const topology = await client.getTopology();

// Get stats
const stats = await client.getStats();
```

---

## React Hooks Example

```typescript
import { useQuery } from '@tanstack/react-query';
import { OtelTuiClient } from './otel-client';

const client = new OtelTuiClient();

export function useTraces(service?: string) {
  return useQuery({
    queryKey: ['traces', service],
    queryFn: () => client.getTraces(service),
    refetchInterval: 5000, // Refresh every 5 seconds
  });
}

export function useTrace(traceId: string) {
  return useQuery({
    queryKey: ['trace', traceId],
    queryFn: () => client.getTrace(traceId),
    enabled: !!traceId,
  });
}

export function useMetrics(service?: string, metric?: string) {
  return useQuery({
    queryKey: ['metrics', service, metric],
    queryFn: () => client.getMetrics(service, metric),
    refetchInterval: 5000,
  });
}

export function useLogs(filter?: string) {
  return useQuery({
    queryKey: ['logs', filter],
    queryFn: () => client.getLogs(filter),
    refetchInterval: 5000,
  });
}

export function useTopology() {
  return useQuery({
    queryKey: ['topology'],
    queryFn: () => client.getTopology(),
    refetchInterval: 10000, // Refresh every 10 seconds
  });
}

export function useServices() {
  return useQuery({
    queryKey: ['services'],
    queryFn: () => client.getServices(),
    refetchInterval: 10000,
  });
}

export function useStats() {
  return useQuery({
    queryKey: ['stats'],
    queryFn: () => client.getStats(),
    refetchInterval: 5000,
  });
}

// Component usage
function TracesList() {
  const { data: traces, isLoading, error } = useTraces();

  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;

  return (
    <div>
      {traces?.map(span => (
        <div key={span.spanId}>
          {span.name} - {span.serviceName}
        </div>
      ))}
    </div>
  );
}
```

---

## Error Handling

All endpoints return a 404 status code with an error message when the requested resource is not found:

```json
{
  "error": "Resource not found"
}
```

Possible error messages:
- `"Trace not found"`
- `"Span not found"`
- `"Spans not found for trace and service"`
- `"Metrics not found for service and metric name"`
- `"Logs not found for trace"`

---

## Best Practices

1. **Polling**: The data in otel-tui is constantly updated as telemetry arrives. Consider polling endpoints every 3-5 seconds for real-time updates.

2. **Filtering**: Use query parameters to filter data on the server side rather than fetching all data and filtering on the client.

3. **Caching**: Implement client-side caching (e.g., with React Query or SWR) to reduce unnecessary requests.

4. **Error Handling**: Always handle 404 errors gracefully, as traces/spans/logs may be rotated out of the circular buffer.

5. **Performance**: The `/api/traces` endpoint can return large amounts of data. Always use the `service` filter when possible.

6. **Timestamp Parsing**: All timestamps are in Unix nanoseconds. Convert them appropriately:
   ```typescript
   const date = new Date(timeUnixNano / 1_000_000); // Convert to milliseconds
   ```

7. **Trace IDs and Span IDs**: These are hex-encoded strings, not integers.

---

## Contributing

If you encounter any issues or have suggestions for the API, please open an issue on the [GitHub repository](https://github.com/ymtdzzz/otel-tui).

---

## License

This API is part of otel-tui and follows the same license.
