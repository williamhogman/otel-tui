# HTTP API Filtering Guide

This document provides comprehensive information about all filtering, pagination, and sorting capabilities available in the otel-tui HTTP API.

## Table of Contents

- [Common Parameters](#common-parameters)
- [Trace Filters](#trace-filters)
- [Metric Filters](#metric-filters)
- [Log Filters](#log-filters)
- [Pagination Headers](#pagination-headers)
- [Examples](#examples)

---

## Common Parameters

These parameters are available across most list endpoints:

### Pagination

| Parameter | Type | Description | Default | Max |
|-----------|------|-------------|---------|-----|
| `offset` | integer | Number of items to skip | 0 | N/A |
| `limit` | integer | Maximum number of items to return | 100 | 1000 |

**Example:**
```
GET /api/traces?offset=20&limit=50
```

### Time Range

| Parameter | Type | Description | Format |
|-----------|------|-------------|--------|
| `start_time` | string/integer | Filter items after this time | RFC3339 or Unix milliseconds |
| `end_time` | string/integer | Filter items before this time | RFC3339 or Unix milliseconds |

**Examples:**
```
# RFC3339 format
GET /api/traces?start_time=2024-01-01T00:00:00Z&end_time=2024-01-02T00:00:00Z

# Unix milliseconds
GET /api/traces?start_time=1704067200000&end_time=1704153600000
```

---

## Trace Filters

**Endpoint:** `GET /api/traces`

### Available Filters

| Parameter | Type | Description | Example Values |
|-----------|------|-------------|----------------|
| `service` | string | Filter by service name (case-insensitive substring) | `frontend`, `api` |
| `status` | string | Filter by span status code | `ok`, `error`, `unset` |
| `min_duration_ms` | integer | Minimum span duration in milliseconds | `100`, `1000` |
| `max_duration_ms` | integer | Maximum span duration in milliseconds | `500`, `5000` |
| `sort_by` | string | Sort field | `time`, `duration`, `name` |
| `sort_order` | string | Sort direction | `asc`, `desc` |

### Sort Options

- `time` (default): Sort by received timestamp
- `duration`: Sort by span duration
- `name`: Sort by span name alphabetically

### Status Values

- `ok`: Successful spans (status code = Ok)
- `error`: Failed spans (status code = Error)
- `unset`: Spans without explicit status

### Example Requests

```bash
# Get error traces only
GET /api/traces?status=error

# Get slow requests (> 1 second)
GET /api/traces?min_duration_ms=1000

# Get traces for frontend service in the last hour, sorted by duration
GET /api/traces?service=frontend&start_time=2024-01-01T10:00:00Z&sort_by=duration&sort_order=desc

# Get fast requests (< 100ms) with errors
GET /api/traces?status=error&max_duration_ms=100

# Paginated results
GET /api/traces?offset=0&limit=20

# Complex query: Frontend errors in the last 24h, duration 100-5000ms, sorted by duration
GET /api/traces?service=frontend&status=error&min_duration_ms=100&max_duration_ms=5000&sort_by=duration&sort_order=desc&limit=50
```

### Response Headers

All list endpoints include pagination metadata in response headers:

```http
X-Total-Count: 1000
X-Filtered-Count: 42
X-Offset: 0
X-Limit: 100
```

---

## Metric Filters

**Endpoint:** `GET /api/metrics`

### Available Filters

| Parameter | Type | Description | Example Values |
|-----------|------|-------------|----------------|
| `service` | string | Filter by service name (case-insensitive substring) | `frontend`, `backend` |
| `metric` | string | Filter by metric name (case-insensitive substring) | `http_requests`, `cpu_usage` |
| `type` | string | Filter by metric type | `Gauge`, `Sum`, `Histogram`, `Summary` |

### Metric Types

- `Gauge`: Point-in-time measurements
- `Sum`: Cumulative values
- `Histogram`: Distribution of values
- `ExponentialHistogram`: Exponential histogram
- `Summary`: Summary statistics

### Example Requests

```bash
# Get all HTTP-related metrics
GET /api/metrics?metric=http

# Get all gauges for the backend service
GET /api/metrics?service=backend&type=Gauge

# Get metrics from the last 5 minutes
GET /api/metrics?start_time=2024-01-01T10:55:00Z&end_time=2024-01-01T11:00:00Z

# Paginated histogram metrics
GET /api/metrics?type=Histogram&offset=0&limit=50

# Complex query: Backend HTTP metrics from last hour
GET /api/metrics?service=backend&metric=http&start_time=2024-01-01T10:00:00Z
```

---

## Log Filters

**Endpoint:** `GET /api/logs`

### Available Filters

| Parameter | Type | Description | Example Values |
|-----------|------|-------------|----------------|
| `service` | string | Filter by service name (case-insensitive substring) | `backend`, `database` |
| `severity` | string | Filter by severity text (case-insensitive substring) | `error`, `warn`, `info` |
| `min_severity` | string | Filter by minimum severity level | `trace`, `debug`, `info`, `warn`, `error`, `fatal` |
| `body` | string | Filter by log body content (case-insensitive substring) | `connection`, `timeout` |
| `trace_id` | string | Filter logs by trace ID | `1234567890abcdef` |

### Severity Levels

Severity levels in ascending order:

| Level | Numeric Value |
|-------|---------------|
| `trace` | 1 |
| `debug` | 5 |
| `info` | 9 |
| `warn` | 13 |
| `error` | 17 |
| `fatal` | 21 |

### Example Requests

```bash
# Get all error logs
GET /api/logs?severity=error

# Get all logs with severity >= WARN
GET /api/logs?min_severity=warn

# Get backend logs containing "connection"
GET /api/logs?service=backend&body=connection

# Get logs for a specific trace
GET /api/logs?trace_id=1234567890abcdef

# Get error logs from the last 10 minutes
GET /api/logs?severity=error&start_time=2024-01-01T10:50:00Z&end_time=2024-01-01T11:00:00Z

# Paginated error and fatal logs
GET /api/logs?min_severity=error&offset=0&limit=100

# Complex query: Backend errors containing "database" in last hour
GET /api/logs?service=backend&min_severity=error&body=database&start_time=2024-01-01T10:00:00Z
```

---

## Pagination Headers

All list endpoints return pagination metadata in HTTP headers. Your client can use these to build pagination UI:

```typescript
const response = await fetch('http://localhost:8000/api/traces?offset=0&limit=20');

const totalCount = parseInt(response.headers.get('X-Total-Count') || '0');
const filteredCount = parseInt(response.headers.get('X-Filtered-Count') || '0');
const offset = parseInt(response.headers.get('X-Offset') || '0');
const limit = parseInt(response.headers.get('X-Limit') || '0');

const currentPage = Math.floor(offset / limit) + 1;
const totalPages = Math.ceil(filteredCount / limit);
const hasNextPage = offset + limit < filteredCount;
const hasPreviousPage = offset > 0;
```

---

## Examples

### React Component with Filters

```typescript
import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';

interface TraceFilters {
  service?: string;
  status?: 'ok' | 'error' | 'unset';
  minDuration?: number;
  maxDuration?: number;
  offset: number;
  limit: number;
  sortBy: 'time' | 'duration' | 'name';
  sortOrder: 'asc' | 'desc';
}

function TraceList() {
  const [filters, setFilters] = useState<TraceFilters>({
    offset: 0,
    limit: 50,
    sortBy: 'time',
    sortOrder: 'desc',
  });

  const { data: traces, isLoading } = useQuery({
    queryKey: ['traces', filters],
    queryFn: async () => {
      const params = new URLSearchParams();

      if (filters.service) params.append('service', filters.service);
      if (filters.status) params.append('status', filters.status);
      if (filters.minDuration) params.append('min_duration_ms', filters.minDuration.toString());
      if (filters.maxDuration) params.append('max_duration_ms', filters.maxDuration.toString());
      params.append('offset', filters.offset.toString());
      params.append('limit', filters.limit.toString());
      params.append('sort_by', filters.sortBy);
      params.append('sort_order', filters.sortOrder);

      const response = await fetch(`http://localhost:8000/api/traces?${params}`);
      const data = await response.json();

      return {
        traces: data,
        totalCount: parseInt(response.headers.get('X-Total-Count') || '0'),
        filteredCount: parseInt(response.headers.get('X-Filtered-Count') || '0'),
      };
    },
    refetchInterval: 5000,
  });

  return (
    <div>
      <div className="filters">
        <input
          placeholder="Service name"
          value={filters.service || ''}
          onChange={(e) => setFilters({ ...filters, service: e.target.value, offset: 0 })}
        />

        <select
          value={filters.status || ''}
          onChange={(e) => setFilters({ ...filters, status: e.target.value as any, offset: 0 })}
        >
          <option value="">All Status</option>
          <option value="ok">OK</option>
          <option value="error">Error</option>
          <option value="unset">Unset</option>
        </select>

        <input
          type="number"
          placeholder="Min duration (ms)"
          value={filters.minDuration || ''}
          onChange={(e) => setFilters({ ...filters, minDuration: parseInt(e.target.value), offset: 0 })}
        />

        <select
          value={filters.sortBy}
          onChange={(e) => setFilters({ ...filters, sortBy: e.target.value as any })}
        >
          <option value="time">Time</option>
          <option value="duration">Duration</option>
          <option value="name">Name</option>
        </select>

        <button onClick={() => setFilters({ ...filters, sortOrder: filters.sortOrder === 'asc' ? 'desc' : 'asc' })}>
          {filters.sortOrder === 'asc' ? '↑' : '↓'}
        </button>
      </div>

      {isLoading ? (
        <div>Loading...</div>
      ) : (
        <div>
          <div>Showing {traces?.filteredCount} of {traces?.totalCount} traces</div>

          <div className="trace-list">
            {traces?.traces.map((trace) => (
              <div key={trace.spanId}>
                <span>{trace.serviceName}</span>
                <span>{trace.name}</span>
                <span>{trace.durationText}</span>
                <span>{trace.status.code}</span>
              </div>
            ))}
          </div>

          <div className="pagination">
            <button
              disabled={filters.offset === 0}
              onClick={() => setFilters({ ...filters, offset: Math.max(0, filters.offset - filters.limit) })}
            >
              Previous
            </button>

            <span>
              Page {Math.floor(filters.offset / filters.limit) + 1} of{' '}
              {Math.ceil((traces?.filteredCount || 0) / filters.limit)}
            </span>

            <button
              disabled={filters.offset + filters.limit >= (traces?.filteredCount || 0)}
              onClick={() => setFilters({ ...filters, offset: filters.offset + filters.limit })}
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
```

### Log Severity Filter Component

```typescript
function LogViewer() {
  const [severity, setSeverity] = useState<string>('');
  const [service, setService] = useState<string>('');

  const { data: logs } = useQuery({
    queryKey: ['logs', severity, service],
    queryFn: async () => {
      const params = new URLSearchParams();
      if (severity) params.append('min_severity', severity);
      if (service) params.append('service', service);
      params.append('limit', '100');

      const response = await fetch(`http://localhost:8000/api/logs?${params}`);
      return response.json();
    },
    refetchInterval: 3000,
  });

  return (
    <div>
      <div className="filters">
        <select value={severity} onChange={(e) => setSeverity(e.target.value)}>
          <option value="">All Severities</option>
          <option value="debug">Debug+</option>
          <option value="info">Info+</option>
          <option value="warn">Warn+</option>
          <option value="error">Error+</option>
          <option value="fatal">Fatal</option>
        </select>

        <input
          placeholder="Service filter"
          value={service}
          onChange={(e) => setService(e.target.value)}
        />
      </div>

      <div className="logs">
        {logs?.map((log, idx) => (
          <div key={idx} className={`log log-${log.severityText.toLowerCase()}`}>
            <span className="timestamp">{new Date(log.timeUnixNano / 1000000).toLocaleString()}</span>
            <span className="severity">{log.severityText}</span>
            <span className="service">{log.serviceName}</span>
            <span className="body">{log.body}</span>
          </div>
        ))}
      </div>
    </div>
  );
}
```

---

## Performance Tips

1. **Use Pagination**: Always use `limit` to avoid loading too much data at once. The API enforces a maximum of 1000 items per request.

2. **Filter Server-Side**: Use query parameters to filter on the server rather than fetching all data and filtering client-side.

3. **Appropriate Polling Intervals**:
   - Traces: 3-5 seconds
   - Metrics: 5-10 seconds
   - Logs: 2-3 seconds for active debugging
   - Stats/Services: 10-30 seconds

4. **Time Range Filters**: When viewing historical data, always use `start_time` and `end_time` to reduce payload size.

5. **Status Filters**: If you only care about errors, use `status=error` to dramatically reduce data transfer.

6. **Combine Filters**: Multiple filters are applied with AND logic, so combining them reduces the result set.

---

## CORS Support

The API supports full CORS for external hosting:

- **Allowed Origins**: `*` (all origins)
- **Allowed Methods**: `GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD`
- **Allowed Headers**: `*` (all headers)
- **Credentials**: Supported
- **Max Age**: 24 hours (86400 seconds)

This means you can host your web GUI on any domain and it will be able to access the otel-tui API.

---

## Error Responses

When filters produce no results, you'll receive an empty array with appropriate pagination headers:

```json
[]
```

Response headers:
```http
X-Total-Count: 1000
X-Filtered-Count: 0
X-Offset: 0
X-Limit: 100
```

This allows your UI to display "No results found" messages appropriately.

---

## Best Practices

1. **Start Broad, Then Narrow**: Begin with fewer filters and add more as users refine their search.

2. **Show Filter State**: Always display active filters to users so they know what they're viewing.

3. **Preserve Filter State**: Store filter preferences in URL query parameters or local storage.

4. **Loading States**: Show loading indicators when filters change and data is being fetched.

5. **Debounce Text Inputs**: For text filters like `service` or `body`, debounce input to avoid excessive API calls.

6. **Progressive Enhancement**: Load initial data quickly, then apply advanced filters as needed.

---

## Future Enhancements

Potential future filtering capabilities (not yet implemented):

- Full-text search across all fields
- Regex pattern matching
- Tag-based filtering
- Saved filter presets
- Filter suggestions based on available data
- Real-time filter statistics

If you need additional filtering capabilities, please open an issue on the [GitHub repository](https://github.com/ymtdzzz/otel-tui).
