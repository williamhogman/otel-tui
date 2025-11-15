# otel-tui Web GUI Implementation Guide

This document provides comprehensive requirements and specifications for building a web-based GUI that replicates all functionality of the otel-tui terminal interface.

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Core Features](#core-features)
4. [Page-by-Page Specifications](#page-by-page-specifications)
5. [Interaction Patterns](#interaction-patterns)
6. [Real-time Updates](#real-time-updates)
7. [Keyboard Shortcuts](#keyboard-shortcuts)
8. [Technical Requirements](#technical-requirements)
9. [UI/UX Guidelines](#uiux-guidelines)

---

## Overview

### What is otel-tui?

otel-tui is a terminal-based OpenTelemetry viewer that displays traces, metrics, and logs in real-time. It collects telemetry data via OTLP (gRPC/HTTP), Zipkin, and Prometheus protocols and provides an interactive interface for analyzing distributed system behavior.

### Web GUI Goals

The web GUI should:
- **Replicate all TUI functionality** in a modern web interface
- **Provide better UX** with clickable elements, hover states, and rich visualizations
- **Support real-time updates** with 3-5 second polling intervals
- **Enable advanced filtering** beyond what the TUI offers
- **Work across browsers** (Chrome, Firefox, Safari, Edge)
- **Be responsive** (desktop-first, with tablet support)

---

## Architecture

### Communication Flow

```
┌─────────────────┐
│  Applications   │ (Send telemetry)
│  (OTLP/Zipkin)  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   otel-tui      │ (Receives on ports 4317/4318/9411)
│   Collector     │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  HTTP API       │ (Port 8000, REST endpoints)
│  (This guide    │
│   documents)    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Web GUI       │ (React/Vue/Svelte app)
│   (You build)   │
└─────────────────┘
```

### API Base URL

```
http://localhost:8000/api
```

All API documentation is available in:
- `HTTP_API_INTEGRATION.md` - Complete API reference with Zod schemas
- `HTTP_API_FILTERS.md` - Filtering and pagination guide

---

## Core Features

### 1. Multi-Page Navigation

The TUI has 5 main pages accessible via Tab/Shift+Tab. The web GUI should have:

- **Traces Page** - List of all trace spans with filtering
- **Timeline Page** - Timeline visualization of a selected trace
- **Metrics Page** - Metrics list with time-series charts
- **Logs Page** - Log entries with severity filtering
- **Topology Page** - Service dependency graph

### 2. Real-Time Data

- Auto-refresh every **3-5 seconds** for active views
- Visual indicator when new data arrives
- Pause/resume auto-refresh toggle
- Last updated timestamp display

### 3. Filtering & Search

- **Service filter** - Filter by service name (all pages)
- **Text search** - Search span names, log bodies, metric names
- **Status filter** - Filter traces by OK/Error/Unset
- **Time range** - Filter by time window (last 5m, 1h, 6h, custom)
- **Severity filter** - Filter logs by severity level
- **Duration filter** - Filter traces by min/max duration

### 4. Sorting

- **Traces**: Sort by time, duration, or name (asc/desc)
- **Metrics**: Sort by time or name
- **Logs**: Sort by timestamp (default: newest first)

### 5. Pagination

- Show 50-100 items per page (configurable)
- Display pagination controls with page numbers
- Show total count and filtered count
- Use API pagination headers (X-Total-Count, X-Filtered-Count)

### 6. Detail Views

- Click on any item to open detail panel/modal
- Show all attributes, events, links
- Copy buttons for IDs and values
- JSON export option

---

## Page-by-Page Specifications

### 📊 Page 1: Traces

**Purpose**: Display all incoming trace spans with filtering and sorting capabilities.

#### Layout

```
┌─────────────────────────────────────────────────────────────┐
│ Traces                                    🔄 Auto-refresh: ON│
├─────────────────────────────────────────────────────────────┤
│ Filters:                                                     │
│ [Service: ______] [Status: All ▼] [Duration: All ▼]        │
│ [Sort: Time ▼] [Order: Desc ▼] [Time Range: Last 1h ▼]    │
├─────────────────────────────────────────────────────────────┤
│ Showing 45 of 856 traces                                    │
├──────┬────────────┬────────────┬──────────┬─────────┬──────┤
│  ●   │ Service    │ Span Name  │ Duration │ Status  │ Time │
├──────┼────────────┼────────────┼──────────┼─────────┼──────┤
│  🔴  │ frontend   │ GET /api   │ 234ms    │ Error   │ 14:32│
│  🟢  │ backend    │ query DB   │ 45ms     │ OK      │ 14:32│
│  🟢  │ cache      │ get user   │ 12ms     │ OK      │ 14:31│
│  ...                                                        │
├─────────────────────────────────────────────────────────────┤
│ ◄ Prev  [1] 2 3 4 5 ... 18  Next ►       50 per page ▼   │
└─────────────────────────────────────────────────────────────┘
```

#### Columns

| Column | Description | Sortable | Filterable |
|--------|-------------|----------|------------|
| **Status Icon** | 🟢 OK, 🔴 Error, ⚪ Unset | No | Yes (dropdown) |
| **Service** | Service name | Yes | Yes (text input) |
| **Span Name** | Operation name | Yes | Yes (via search) |
| **Duration** | Span duration (formatted) | Yes | Yes (min/max sliders) |
| **Status** | OK/Error/Unset text | Yes | Yes (dropdown) |
| **Time** | Received timestamp | Yes | Yes (time range picker) |

#### Interactions

1. **Click on row**: Open detailed span view (see [Span Detail Modal](#span-detail-modal))
2. **Hover on row**: Highlight and show tooltip with trace ID
3. **Click on service**: Filter table to that service
4. **Double-click on span name**: Copy to clipboard
5. **Right-click**: Context menu with:
   - View full trace timeline
   - View related logs
   - Copy trace ID
   - Copy span ID
   - Export span as JSON

#### API Calls

```typescript
// Fetch traces with filters
GET /api/traces?service={service}&status={status}&min_duration_ms={min}&max_duration_ms={max}&sort_by={sortBy}&sort_order={order}&offset={offset}&limit={limit}&start_time={startTime}&end_time={endTime}

// Poll every 5 seconds for updates
setInterval(() => fetchTraces(), 5000);
```

#### Filters Panel

**Service Filter**
```typescript
<input
  type="text"
  placeholder="Filter by service..."
  onChange={debounce(setServiceFilter, 300)}
/>
```

**Status Filter**
```typescript
<select value={statusFilter}>
  <option value="">All Status</option>
  <option value="ok">OK</option>
  <option value="error">Error</option>
  <option value="unset">Unset</option>
</select>
```

**Duration Filter**
```typescript
<div>
  Min: <input type="number" placeholder="ms" />
  Max: <input type="number" placeholder="ms" />
</div>
```

**Time Range Filter**
```typescript
<select value={timeRange}>
  <option value="5m">Last 5 minutes</option>
  <option value="15m">Last 15 minutes</option>
  <option value="1h">Last 1 hour</option>
  <option value="6h">Last 6 hours</option>
  <option value="24h">Last 24 hours</option>
  <option value="custom">Custom range...</option>
</select>
```

---

### 🕐 Page 2: Timeline (Trace Detail)

**Purpose**: Visualize a single trace's spans across services in a timeline/waterfall chart.

#### Layout

```
┌─────────────────────────────────────────────────────────────┐
│ ◄ Back to Traces     Trace: 1234567890abcdef               │
├─────────────────────────────────────────────────────────────┤
│ Services: frontend → backend → database                     │
│ Total Duration: 234ms  |  Spans: 5  |  Status: Error 🔴    │
├─────────────────────────────────────────────────────────────┤
│                    Timeline (0ms ───────────► 234ms)        │
├─────────────────────────────────────────────────────────────┤
│ frontend                                                     │
│ │ GET /api           ████████████████████░░░░░░ 234ms      │
│   └─ render UI       ░░░░░░░░░░░░░░░░████░░░░░ 45ms       │
├─────────────────────────────────────────────────────────────┤
│ backend                                                      │
│   └─ query           ░░░░███████░░░░░░░░░░░░░░ 120ms      │
│     └─ validate      ░░░░██░░░░░░░░░░░░░░░░░░░ 15ms  🔴   │
├─────────────────────────────────────────────────────────────┤
│ database                                                     │
│     └─ SELECT        ░░░░░░░██████████░░░░░░░░ 90ms       │
├─────────────────────────────────────────────────────────────┤
│ Related Logs (3)                              [View All]    │
│ • 14:32:15 [ERROR] Query validation failed                  │
│ • 14:32:14 [INFO] Processing request                        │
└─────────────────────────────────────────────────────────────┘
```

#### Timeline Visualization

**Requirements**:
1. **Waterfall Chart**: Each span is a horizontal bar
2. **Time Scale**: Linear scale showing milliseconds
3. **Nesting**: Child spans indented under parents
4. **Color Coding**:
   - 🟢 Green: OK status
   - 🔴 Red: Error status
   - 🟡 Yellow: Unset status
   - 🔵 Blue: Different span kinds (Client, Server, Internal, etc.)
5. **Hover State**: Show tooltip with:
   - Full span name
   - Exact duration
   - Start/end timestamps
   - Attributes preview
6. **Click**: Open span detail modal

#### Service Groups

Group spans by service with collapsible sections:

```typescript
<ServiceGroup service="frontend" spanCount={2}>
  <Span name="GET /api" duration="234ms" status="error" />
  <Span name="render UI" duration="45ms" status="ok" indent={1} />
</ServiceGroup>
```

#### Related Logs

Show logs that share the same trace ID:

```typescript
// Fetch related logs
GET /api/logs/trace/{traceId}

// Display in timeline
logs.map(log => (
  <LogEntry
    timestamp={log.timeUnixNano}
    severity={log.severityText}
    body={log.body}
    service={log.serviceName}
  />
))
```

#### API Calls

```typescript
// Get all spans for trace
GET /api/traces/{traceId}

// Get logs for trace
GET /api/logs/trace/{traceId}
```

---

### 📈 Page 3: Metrics

**Purpose**: Display metrics with time-series charts and filtering.

#### Layout

```
┌─────────────────────────────────────────────────────────────┐
│ Metrics                                   🔄 Auto-refresh: ON│
├─────────────────────────────────────────────────────────────┤
│ Filters:                                                     │
│ [Service: ______] [Metric: ______] [Type: All ▼]           │
├─────────────────────────────────────────────────────────────┤
│ Showing 42 of 2341 metrics                                  │
├────────────┬───────────────────────┬─────────┬──────────────┤
│ Service    │ Metric Name           │ Type    │ Latest Value │
├────────────┼───────────────────────┼─────────┼──────────────┤
│ frontend   │ http_requests_total   │ Sum     │ 1,234        │
│ frontend   │ http_request_duration │ Histogram│ p95: 234ms  │
│ backend    │ cpu_usage_percent     │ Gauge   │ 45.3%        │
│ database   │ query_duration_ms     │ Histogram│ p99: 890ms  │
│ ...                                                          │
└─────────────────────────────────────────────────────────────┘
```

#### Metric Detail View

Click on a metric to see time-series chart:

```
┌─────────────────────────────────────────────────────────────┐
│ ◄ Back     frontend • http_requests_total                   │
├─────────────────────────────────────────────────────────────┤
│ Type: Sum  │  Unit: 1  │  Description: Total HTTP requests  │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1400 ┤                                              ╭──     │
│       │                                         ╭────╯       │
│  1200 ┤                                    ╭────╯            │
│       │                               ╭────╯                 │
│  1000 ┤                          ╭────╯                      │
│       │                     ╭────╯                           │
│   800 ┤                ╭────╯                                │
│       │           ╭────╯                                     │
│   600 ┤      ╭────╯                                          │
│       ├──────╯                                               │
│       └───┬───┬───┬───┬───┬───┬───┬───┬───┬───┬───         │
│         14:20 14:25 14:30 14:35 14:40 14:45 14:50          │
│                                                              │
├─────────────────────────────────────────────────────────────┤
│ Data Points (showing last 100)                              │
│ Time        │ Value │ Attributes                            │
│ 14:50:23    │ 1,234 │ method=GET, status=200                │
│ 14:50:18    │ 1,220 │ method=GET, status=200                │
│ ...                                                          │
└─────────────────────────────────────────────────────────────┘
```

#### Chart Types by Metric Type

**Gauge**: Line chart showing current values over time
```typescript
<LineChart data={dataPoints} yAxis="value" xAxis="time" />
```

**Sum**: Area chart showing cumulative values
```typescript
<AreaChart data={dataPoints} yAxis="value" xAxis="time" />
```

**Histogram**: Distribution chart (bar chart)
```typescript
<BarChart data={buckets} yAxis="count" xAxis="bucket" />
```

**Summary**: Percentile lines (p50, p90, p95, p99)
```typescript
<MultiLineChart
  data={dataPoints}
  lines={['p50', 'p90', 'p95', 'p99']}
/>
```

#### API Calls

```typescript
// Get metrics list
GET /api/metrics?service={service}&metric={metric}&type={type}&limit=100

// Get specific metric details
GET /api/metrics/{service}/{metricName}

// Poll every 10 seconds
setInterval(() => fetchMetrics(), 10000);
```

---

### 📝 Page 4: Logs

**Purpose**: Display log entries with severity filtering and search.

#### Layout

```
┌─────────────────────────────────────────────────────────────┐
│ Logs                                      🔄 Auto-refresh: ON│
├─────────────────────────────────────────────────────────────┤
│ Filters:                                                     │
│ [Service: ______] [Severity: All ▼] [Search: ______]       │
│ [Time Range: Last 1h ▼]                                    │
├─────────────────────────────────────────────────────────────┤
│ Showing 123 of 432 logs                                     │
├──────┬──────────┬─────────┬────────────────────────────────┤
│ Sev  │ Service  │ Time    │ Message                        │
├──────┼──────────┼─────────┼────────────────────────────────┤
│ 🔴 E │ backend  │ 14:32:15│ Query validation failed: ...  │
│ 🟡 W │ cache    │ 14:32:10│ Cache miss for key user:123   │
│ 🔵 I │ frontend │ 14:32:05│ Processing request GET /api   │
│ 🟢 D │ database │ 14:31:58│ Connection pool: 5/10 active  │
│ ...                                                         │
└─────────────────────────────────────────────────────────────┘
```

#### Severity Levels

Display with color coding:

| Severity | Icon | Color | Description |
|----------|------|-------|-------------|
| FATAL    | ⚫   | Black | Critical errors |
| ERROR    | 🔴   | Red | Errors |
| WARN     | 🟡   | Yellow | Warnings |
| INFO     | 🔵   | Blue | Informational |
| DEBUG    | 🟢   | Green | Debug messages |
| TRACE    | ⚪   | Gray | Trace messages |

#### Log Detail Modal

Click on a log entry to see full details:

```
┌─────────────────────────────────────────────────────────────┐
│ Log Details                                             [X] │
├─────────────────────────────────────────────────────────────┤
│ Time: 2024-01-15 14:32:15.123456 UTC                       │
│ Severity: ERROR (17)                                        │
│ Service: backend                                            │
│ Trace ID: 1234567890abcdef [View Trace]                   │
│ Span ID: fedcba654321                                      │
├─────────────────────────────────────────────────────────────┤
│ Body:                                                       │
│ Query validation failed: Invalid user ID format            │
│                                                             │
│ Attributes:                                                 │
│ • error.type: ValidationError                              │
│ • user.id: abc-invalid                                     │
│ • query.duration_ms: 12                                    │
│                                                             │
│ Resource Attributes:                                        │
│ • service.name: backend                                    │
│ • service.version: 1.2.3                                   │
│ • host.name: backend-pod-abc123                            │
├─────────────────────────────────────────────────────────────┤
│ [Copy JSON] [Copy Body] [View Related Trace]              │
└─────────────────────────────────────────────────────────────┘
```

#### Severity Filter

```typescript
<select value={minSeverity}>
  <option value="">All Severities</option>
  <option value="trace">Trace+</option>
  <option value="debug">Debug+</option>
  <option value="info">Info+</option>
  <option value="warn">Warn+</option>
  <option value="error">Error+</option>
  <option value="fatal">Fatal</option>
</select>
```

#### API Calls

```typescript
// Get logs with filters
GET /api/logs?service={service}&min_severity={severity}&body={searchText}&offset={offset}&limit={limit}

// Get log detail
// Logs are fetched in full, no separate detail endpoint needed

// Poll every 3 seconds for active monitoring
setInterval(() => fetchLogs(), 3000);
```

---

### 🕸️ Page 5: Topology (Service Map)

**Purpose**: Visualize service dependencies as a directed graph.

#### Layout

```
┌─────────────────────────────────────────────────────────────┐
│ Topology                                  🔄 Auto-refresh: ON│
├─────────────────────────────────────────────────────────────┤
│                                                              │
│                     ┌──────────┐                            │
│                     │ frontend │                            │
│                     └─────┬────┘                            │
│                           │ 42 calls                         │
│                           ▼                                  │
│                     ┌──────────┐                            │
│                     │ backend  │                            │
│                     └─────┬────┘                            │
│                           │ 38 calls                         │
│                           ▼                                  │
│                     ┌──────────┐                            │
│                     │ database │                            │
│                     └──────────┘                            │
│                                                              │
│        ┌──────────┐                                         │
│        │  cache   │◄───────────┐                           │
│        └──────────┘            │ 12 calls                   │
│                                │                             │
├─────────────────────────────────────────────────────────────┤
│ Legend: Circle size = request volume                        │
│         Arrow thickness = call frequency                    │
└─────────────────────────────────────────────────────────────┘
```

#### Graph Visualization

**Node Properties**:
- **Service Name**: Displayed in the center of node
- **Size**: Proportional to number of spans
- **Color**:
  - 🟢 Green: All spans OK
  - 🔴 Red: Has error spans
  - 🟡 Yellow: Mixed status

**Edge Properties**:
- **Direction**: Shows call direction (A → B means A calls B)
- **Label**: Shows call count
- **Thickness**: Proportional to call frequency
- **Color**:
  - Gray: All calls successful
  - Red: Some calls failed

#### Interactions

1. **Click on node**: Show service details panel with:
   - Service name
   - Total span count
   - Error rate
   - Average duration
   - List of dependent services
   - Button to "Filter traces for this service"

2. **Click on edge**: Show edge details:
   - Source → Target
   - Call count
   - Average latency
   - Error rate
   - Button to "View traces for this path"

3. **Hover on node**: Highlight node and connected edges

4. **Zoom/Pan**: Allow zooming and panning for large graphs

#### Layout Algorithms

Use automatic graph layout (choose one):
- **Hierarchical**: Top-down flow (recommended)
- **Force-directed**: Physics-based layout
- **Circular**: Services in a circle

#### API Calls

```typescript
// Get topology data
GET /api/topology

// Response includes nodes and edges
const topology = {
  nodes: [
    { service: "frontend", depth: 0 },
    { service: "backend", depth: 0 },
    { service: "database", depth: 0 }
  ],
  edges: [
    { source: "frontend", target: "backend", count: 42 },
    { source: "backend", target: "database", count: 38 }
  ]
};

// Poll every 10 seconds
setInterval(() => fetchTopology(), 10000);
```

#### Rendering Libraries

Recommended libraries:
- **React Flow** (React)
- **D3.js** (Any framework)
- **Cytoscape.js** (Any framework)
- **vis.js** (Any framework)

---

## Interaction Patterns

### Global Navigation

```
┌─────────────────────────────────────────────────────────────┐
│ otel-tui                         📊 Stats  ⚙️ Settings  ?  │
├─────────────────────────────────────────────────────────────┤
│ Traces │ Timeline │ Metrics │ Logs │ Topology              │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│                     (Page Content)                           │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Tab Navigation**: Click on page names to switch views

**Stats Button**: Shows modal with:
```
Current Data:
• Traces: 856 (max 1000)
• Metrics: 2341 (max 3000)
• Logs: 432 (max 1000)
• Services: 4
• Last Updated: 2 seconds ago

Connection:
• OTLP gRPC: port 4317
• OTLP HTTP: port 4318
• HTTP API: port 8000
```

**Settings Button**: Configure:
- Auto-refresh interval
- Items per page
- Theme (light/dark)
- Time format (12h/24h)
- Timezone

### Common UI Patterns

#### Loading States

```typescript
{isLoading ? (
  <div className="spinner">
    <Spinner /> Loading traces...
  </div>
) : (
  <TraceTable data={traces} />
)}
```

#### Empty States

```typescript
{traces.length === 0 ? (
  <div className="empty-state">
    <Icon name="traces" />
    <h3>No traces found</h3>
    <p>Traces will appear here as they are collected</p>
    <Button onClick={() => setFilters({})}>Clear filters</Button>
  </div>
) : (
  <TraceTable data={traces} />
)}
```

#### Error States

```typescript
{error ? (
  <div className="error-state">
    <Icon name="error" />
    <h3>Failed to load traces</h3>
    <p>{error.message}</p>
    <Button onClick={retry}>Retry</Button>
  </div>
) : (
  <TraceTable data={traces} />
)}
```

#### Copy to Clipboard

```typescript
<button onClick={() => {
  navigator.clipboard.writeText(traceId);
  showToast('Trace ID copied!');
}}>
  📋 Copy
</button>
```

### Detail Modals

All detail views should be in modals/slide-outs with:
- **Close button** (X in corner)
- **Overlay** (click outside to close)
- **Keyboard support** (ESC to close)
- **Copy buttons** for IDs and values
- **JSON export** button
- **Related data links** (e.g., "View related logs")

---

## Real-time Updates

### Polling Strategy

```typescript
function useAutoRefresh(fetchFn, interval = 5000, enabled = true) {
  useEffect(() => {
    if (!enabled) return;

    const id = setInterval(fetchFn, interval);
    return () => clearInterval(id);
  }, [fetchFn, interval, enabled]);
}

// Usage
const [autoRefresh, setAutoRefresh] = useState(true);
useAutoRefresh(fetchTraces, 5000, autoRefresh);
```

### Recommended Intervals

| Page | Interval | Reason |
|------|----------|--------|
| Traces | 5 seconds | High-frequency data |
| Timeline | N/A | Static view, no refresh needed |
| Metrics | 10 seconds | Metrics change less frequently |
| Logs | 3 seconds | Important for debugging |
| Topology | 10 seconds | Graph doesn't change often |
| Stats | 5 seconds | Show current counts |

### New Data Indication

Show visual indicator when new data arrives:

```typescript
<div className="page-header">
  Traces
  {hasNewData && (
    <span className="badge">
      {newCount} new
    </span>
  )}
</div>
```

---

## Keyboard Shortcuts

Implement keyboard shortcuts matching TUI patterns:

| Key | Action |
|-----|--------|
| `Tab` | Next page |
| `Shift + Tab` | Previous page |
| `r` | Refresh current page |
| `f` | Focus filter input |
| `/` | Focus search input |
| `?` | Show keyboard shortcuts help |
| `Esc` | Close modal/clear focus |
| `j` / `k` | Navigate table rows (vim-style) |
| `Enter` | Open selected item detail |
| `c` | Copy selected ID |

Implementation:

```typescript
useEffect(() => {
  const handleKeyPress = (e: KeyboardEvent) => {
    if (e.target.tagName === 'INPUT') return; // Don't interfere with inputs

    switch(e.key) {
      case 'r':
        fetchData();
        break;
      case 'f':
        filterInputRef.current?.focus();
        break;
      case '?':
        setShowHelp(true);
        break;
    }
  };

  window.addEventListener('keydown', handleKeyPress);
  return () => window.removeEventListener('keydown', handleKeyPress);
}, []);
```

---

## Technical Requirements

### Frontend Stack Recommendations

**Frameworks**: (Choose one)
- React + TypeScript
- Vue 3 + TypeScript
- Svelte + TypeScript

**State Management**:
- Zustand (React)
- Pinia (Vue)
- Svelte stores (Svelte)

**Data Fetching**:
- TanStack Query (React Query) ✅ Recommended
- SWR
- Axios with custom hooks

**UI Components**:
- shadcn/ui (React)
- Headless UI
- Radix UI
- PrimeVue (Vue)

**Charts**:
- Recharts (React) ✅ Recommended
- Apache ECharts
- Chart.js
- D3.js

**Graph Visualization**:
- React Flow ✅ Recommended
- Cytoscape.js
- vis.js

### API Integration

Use the comprehensive API documentation:
- `HTTP_API_INTEGRATION.md` - Full Zod schemas
- `HTTP_API_FILTERS.md` - Filtering examples

Example client:

```typescript
import { z } from 'zod';

class OtelTuiAPI {
  constructor(private baseUrl = 'http://localhost:8000/api') {}

  async getTraces(filters: TraceFilters) {
    const params = new URLSearchParams();
    Object.entries(filters).forEach(([key, value]) => {
      if (value) params.append(key, String(value));
    });

    const response = await fetch(`${this.baseUrl}/traces?${params}`);
    const data = await response.json();

    return {
      traces: SpanSchema.array().parse(data),
      totalCount: parseInt(response.headers.get('X-Total-Count') || '0'),
      filteredCount: parseInt(response.headers.get('X-Filtered-Count') || '0'),
    };
  }

  // ... other methods
}
```

### Performance Considerations

1. **Virtualization**: Use virtual scrolling for large lists (react-window, react-virtuoso)
2. **Memoization**: Memoize expensive computations
3. **Debouncing**: Debounce search inputs (300ms)
4. **Lazy Loading**: Load detail views on-demand
5. **Code Splitting**: Split by route
6. **Caching**: Cache API responses with React Query

---

## UI/UX Guidelines

### Color Scheme

**Status Colors**:
- 🟢 Success: `#10b981` (green-500)
- 🔴 Error: `#ef4444` (red-500)
- 🟡 Warning: `#f59e0b` (yellow-500)
- 🔵 Info: `#3b82f6` (blue-500)
- ⚪ Neutral: `#6b7280` (gray-500)

**Theme**:
- Light mode (default)
- Dark mode (toggle)

### Typography

- **Headings**: Inter, system-ui, sans-serif
- **Body**: Inter, system-ui, sans-serif
- **Code**: 'Fira Code', 'Courier New', monospace

### Spacing

Use consistent spacing scale (Tailwind-style):
- 0.25rem (1)
- 0.5rem (2)
- 0.75rem (3)
- 1rem (4)
- 1.5rem (6)
- 2rem (8)

### Responsive Breakpoints

- Desktop: 1280px+ (primary target)
- Tablet: 768px-1279px
- Mobile: < 768px (minimal support)

### Accessibility

- **ARIA labels**: Add to all interactive elements
- **Keyboard navigation**: Full keyboard support
- **Focus indicators**: Clear focus states
- **Color contrast**: WCAG AA compliant
- **Screen readers**: Semantic HTML

---

## Implementation Checklist

### Phase 1: Core Structure
- [ ] Set up project with TypeScript
- [ ] Install dependencies (React Query, Zod, chart library)
- [ ] Create API client with Zod schemas
- [ ] Implement global navigation
- [ ] Add theme toggle (light/dark)

### Phase 2: Traces Page
- [ ] Build traces table with sorting
- [ ] Add filtering (service, status, duration)
- [ ] Implement pagination
- [ ] Add span detail modal
- [ ] Add copy buttons
- [ ] Implement auto-refresh

### Phase 3: Timeline Page
- [ ] Create waterfall chart component
- [ ] Group spans by service
- [ ] Show parent-child relationships
- [ ] Add related logs section
- [ ] Implement zoom/pan

### Phase 4: Metrics Page
- [ ] Build metrics table
- [ ] Add metric detail view with charts
- [ ] Implement chart switching by metric type
- [ ] Add filtering
- [ ] Add auto-refresh

### Phase 5: Logs Page
- [ ] Build logs table
- [ ] Add severity filtering
- [ ] Add search functionality
- [ ] Create log detail modal
- [ ] Link logs to traces
- [ ] Implement auto-refresh

### Phase 6: Topology Page
- [ ] Choose graph library
- [ ] Render service nodes
- [ ] Draw edges with counts
- [ ] Add node/edge click handlers
- [ ] Implement layout algorithm
- [ ] Add auto-refresh

### Phase 7: Polish
- [ ] Add loading states
- [ ] Add empty states
- [ ] Add error handling
- [ ] Implement keyboard shortcuts
- [ ] Add settings panel
- [ ] Add stats modal
- [ ] Performance optimization
- [ ] Accessibility audit
- [ ] Browser testing

---

## Additional Features (Nice to Have)

### 1. Dark Mode
Toggle between light and dark themes with persistent storage.

### 2. Export Functionality
- Export traces as JSON
- Export metrics as CSV
- Export topology as PNG/SVG

### 3. Saved Filters
Save commonly-used filter combinations as presets.

### 4. Trace Comparison
Select and compare multiple traces side-by-side.

### 5. Custom Time Zones
Allow users to view timestamps in their local timezone.

### 6. Notifications
Desktop notifications for new errors or critical logs.

### 7. Search History
Keep history of recent searches for quick access.

### 8. Advanced Filtering
- Filter by attribute values
- Regex pattern matching
- Multiple condition combinations (AND/OR)

---

## Support & Resources

### Documentation
- `HTTP_API_INTEGRATION.md` - Complete API reference
- `HTTP_API_FILTERS.md` - Filtering guide with examples
- This document - Full feature specifications

### Example Requests

```bash
# Get recent error traces
curl "http://localhost:8000/api/traces?status=error&limit=20&sort_by=time&sort_order=desc"

# Get backend metrics
curl "http://localhost:8000/api/metrics?service=backend"

# Get error logs from last hour
curl "http://localhost:8000/api/logs?min_severity=error&start_time=2024-01-15T13:00:00Z"

# Get topology
curl "http://localhost:8000/api/topology"

# Get stats
curl "http://localhost:8000/api/stats"
```

### Testing with Sample Data

If you need sample data:
1. Run otel-tui with `--from-json-file` to load test data
2. Use the OpenTelemetry SDK to generate test telemetry
3. Use the otel-tui demo examples from the GitHub repo

---

## Conclusion

This guide provides comprehensive specifications for building a web GUI that replicates all otel-tui functionality. The HTTP API provides all necessary data with powerful filtering capabilities.

**Key Success Factors**:
1. Real-time updates with appropriate polling intervals
2. Comprehensive filtering matching or exceeding TUI capabilities
3. Clear visual hierarchy and intuitive navigation
4. Responsive and performant even with large datasets
5. Keyboard shortcuts for power users

For questions or clarifications, refer to:
- API documentation: `HTTP_API_INTEGRATION.md`
- Filter guide: `HTTP_API_FILTERS.md`
- Original TUI: https://github.com/ymtdzzz/otel-tui

**Happy building!** 🚀
