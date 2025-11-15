package httpserver

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Offset int
	Limit  int
}

// TimeRangeParams holds time range filter parameters
type TimeRangeParams struct {
	StartTime *time.Time
	EndTime   *time.Time
}

// TraceFilterParams holds all trace filtering parameters
type TraceFilterParams struct {
	Service      string
	Status       string // "ok", "error", "unset"
	MinDuration  *time.Duration
	MaxDuration  *time.Duration
	TimeRange    TimeRangeParams
	Pagination   PaginationParams
	SortBy       string // "time", "duration", "name"
	SortOrder    string // "asc", "desc"
}

// LogFilterParams holds all log filtering parameters
type LogFilterParams struct {
	Service       string
	Severity      string // "trace", "debug", "info", "warn", "error", "fatal"
	MinSeverity   int32
	Body          string
	TraceID       string
	TimeRange     TimeRangeParams
	Pagination    PaginationParams
}

// MetricFilterParams holds all metric filtering parameters
type MetricFilterParams struct {
	Service    string
	MetricName string
	MetricType string // "Gauge", "Sum", "Histogram", "ExponentialHistogram", "Summary"
	TimeRange  TimeRangeParams
	Pagination PaginationParams
}

// ParsePaginationParams parses pagination query parameters
func ParsePaginationParams(r *http.Request) PaginationParams {
	params := PaginationParams{
		Offset: 0,
		Limit:  100, // Default limit
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if val, err := strconv.Atoi(offset); err == nil && val >= 0 {
			params.Offset = val
		}
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil && val > 0 {
			if val > 1000 {
				val = 1000 // Max limit
			}
			params.Limit = val
		}
	}

	return params
}

// ParseTimeRangeParams parses time range query parameters
func ParseTimeRangeParams(r *http.Request) TimeRangeParams {
	params := TimeRangeParams{}

	if start := r.URL.Query().Get("start_time"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			params.StartTime = &t
		} else if unix, err := strconv.ParseInt(start, 10, 64); err == nil {
			t := time.Unix(0, unix*1000000) // Assume milliseconds
			params.StartTime = &t
		}
	}

	if end := r.URL.Query().Get("end_time"); end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			params.EndTime = &t
		} else if unix, err := strconv.ParseInt(end, 10, 64); err == nil {
			t := time.Unix(0, unix*1000000) // Assume milliseconds
			params.EndTime = &t
		}
	}

	return params
}

// ParseTraceFilterParams parses all trace filter parameters
func ParseTraceFilterParams(r *http.Request) TraceFilterParams {
	params := TraceFilterParams{
		Service:    r.URL.Query().Get("service"),
		Status:     strings.ToLower(r.URL.Query().Get("status")),
		TimeRange:  ParseTimeRangeParams(r),
		Pagination: ParsePaginationParams(r),
		SortBy:     strings.ToLower(r.URL.Query().Get("sort_by")),
		SortOrder:  strings.ToLower(r.URL.Query().Get("sort_order")),
	}

	// Parse duration filters
	if minDur := r.URL.Query().Get("min_duration_ms"); minDur != "" {
		if val, err := strconv.ParseInt(minDur, 10, 64); err == nil {
			dur := time.Duration(val) * time.Millisecond
			params.MinDuration = &dur
		}
	}

	if maxDur := r.URL.Query().Get("max_duration_ms"); maxDur != "" {
		if val, err := strconv.ParseInt(maxDur, 10, 64); err == nil {
			dur := time.Duration(val) * time.Millisecond
			params.MaxDuration = &dur
		}
	}

	// Default sort
	if params.SortBy == "" {
		params.SortBy = "time"
	}
	if params.SortOrder == "" {
		params.SortOrder = "desc"
	}

	return params
}

// ParseLogFilterParams parses all log filter parameters
func ParseLogFilterParams(r *http.Request) LogFilterParams {
	params := LogFilterParams{
		Service:    r.URL.Query().Get("service"),
		Severity:   strings.ToLower(r.URL.Query().Get("severity")),
		Body:       r.URL.Query().Get("body"),
		TraceID:    r.URL.Query().Get("trace_id"),
		TimeRange:  ParseTimeRangeParams(r),
		Pagination: ParsePaginationParams(r),
	}

	// Parse minimum severity
	if minSev := r.URL.Query().Get("min_severity"); minSev != "" {
		params.MinSeverity = severityNameToNumber(minSev)
	}

	return params
}

// ParseMetricFilterParams parses all metric filter parameters
func ParseMetricFilterParams(r *http.Request) MetricFilterParams {
	params := MetricFilterParams{
		Service:    r.URL.Query().Get("service"),
		MetricName: r.URL.Query().Get("metric"),
		MetricType: r.URL.Query().Get("type"),
		TimeRange:  ParseTimeRangeParams(r),
		Pagination: ParsePaginationParams(r),
	}

	return params
}

// FilterSpans applies all filters to a slice of spans
func FilterSpans(spans []*telemetry.SpanData, params TraceFilterParams) []*telemetry.SpanData {
	filtered := make([]*telemetry.SpanData, 0, len(spans))

	for _, span := range spans {
		if !matchesSpanFilters(span, params) {
			continue
		}
		filtered = append(filtered, span)
	}

	// Sort
	sortSpans(filtered, params.SortBy, params.SortOrder)

	// Paginate
	return paginateSpans(filtered, params.Pagination)
}

// matchesSpanFilters checks if a span matches all filter criteria
func matchesSpanFilters(span *telemetry.SpanData, params TraceFilterParams) bool {
	// Service filter
	if params.Service != "" {
		serviceName := span.GetServiceName()
		spanName := span.GetSpanName()
		target := serviceName + " " + spanName
		if !strings.Contains(strings.ToLower(target), strings.ToLower(params.Service)) {
			return false
		}
	}

	// Status filter
	if params.Status != "" {
		statusCode := span.Span.Status().Code()
		switch params.Status {
		case "ok":
			if statusCode != ptrace.StatusCodeOk {
				return false
			}
		case "error":
			if statusCode != ptrace.StatusCodeError {
				return false
			}
		case "unset":
			if statusCode != ptrace.StatusCodeUnset {
				return false
			}
		}
	}

	// Duration filters
	duration := span.Span.EndTimestamp().AsTime().Sub(span.Span.StartTimestamp().AsTime())
	if params.MinDuration != nil && duration < *params.MinDuration {
		return false
	}
	if params.MaxDuration != nil && duration > *params.MaxDuration {
		return false
	}

	// Time range filter
	if params.TimeRange.StartTime != nil && span.ReceivedAt.Before(*params.TimeRange.StartTime) {
		return false
	}
	if params.TimeRange.EndTime != nil && span.ReceivedAt.After(*params.TimeRange.EndTime) {
		return false
	}

	return true
}

// sortSpans sorts spans based on the given criteria
func sortSpans(spans []*telemetry.SpanData, sortBy, sortOrder string) {
	ascending := sortOrder == "asc"

	switch sortBy {
	case "duration":
		sortSpansByDuration(spans, ascending)
	case "name":
		sortSpansByName(spans, ascending)
	case "time":
		fallthrough
	default:
		sortSpansByTime(spans, ascending)
	}
}

// Helper sorting functions
func sortSpansByTime(spans []*telemetry.SpanData, ascending bool) {
	// Simple bubble sort for small datasets (max 1000)
	for i := 0; i < len(spans); i++ {
		for j := i + 1; j < len(spans); j++ {
			swap := false
			if ascending {
				swap = spans[i].ReceivedAt.After(spans[j].ReceivedAt)
			} else {
				swap = spans[i].ReceivedAt.Before(spans[j].ReceivedAt)
			}
			if swap {
				spans[i], spans[j] = spans[j], spans[i]
			}
		}
	}
}

func sortSpansByDuration(spans []*telemetry.SpanData, ascending bool) {
	for i := 0; i < len(spans); i++ {
		for j := i + 1; j < len(spans); j++ {
			dur1 := spans[i].Span.EndTimestamp().AsTime().Sub(spans[i].Span.StartTimestamp().AsTime())
			dur2 := spans[j].Span.EndTimestamp().AsTime().Sub(spans[j].Span.StartTimestamp().AsTime())
			swap := false
			if ascending {
				swap = dur1 > dur2
			} else {
				swap = dur1 < dur2
			}
			if swap {
				spans[i], spans[j] = spans[j], spans[i]
			}
		}
	}
}

func sortSpansByName(spans []*telemetry.SpanData, ascending bool) {
	for i := 0; i < len(spans); i++ {
		for j := i + 1; j < len(spans); j++ {
			swap := false
			if ascending {
				swap = spans[i].GetSpanName() > spans[j].GetSpanName()
			} else {
				swap = spans[i].GetSpanName() < spans[j].GetSpanName()
			}
			if swap {
				spans[i], spans[j] = spans[j], spans[i]
			}
		}
	}
}

// paginateSpans applies pagination to spans
func paginateSpans(spans []*telemetry.SpanData, pagination PaginationParams) []*telemetry.SpanData {
	if pagination.Offset >= len(spans) {
		return []*telemetry.SpanData{}
	}

	end := pagination.Offset + pagination.Limit
	if end > len(spans) {
		end = len(spans)
	}

	return spans[pagination.Offset:end]
}

// FilterLogs applies all filters to a slice of logs
func FilterLogs(logs []*telemetry.LogData, params LogFilterParams) []*telemetry.LogData {
	filtered := make([]*telemetry.LogData, 0, len(logs))

	for _, log := range logs {
		if !matchesLogFilters(log, params) {
			continue
		}
		filtered = append(filtered, log)
	}

	return paginateLogs(filtered, params.Pagination)
}

// matchesLogFilters checks if a log matches all filter criteria
func matchesLogFilters(log *telemetry.LogData, params LogFilterParams) bool {
	// Service filter
	if params.Service != "" {
		if !strings.Contains(strings.ToLower(log.GetServiceName()), strings.ToLower(params.Service)) {
			return false
		}
	}

	// Severity filter
	if params.Severity != "" {
		if !strings.Contains(strings.ToLower(log.GetSeverity()), strings.ToLower(params.Severity)) {
			return false
		}
	}

	// Minimum severity filter
	if params.MinSeverity > 0 {
		if int32(log.Log.SeverityNumber()) < params.MinSeverity {
			return false
		}
	}

	// Body filter
	if params.Body != "" {
		if !strings.Contains(strings.ToLower(log.GetResolvedBody()), strings.ToLower(params.Body)) {
			return false
		}
	}

	// Trace ID filter
	if params.TraceID != "" {
		if log.GetTraceID() != params.TraceID {
			return false
		}
	}

	// Time range filter
	logTime := log.Log.Timestamp().AsTime()
	if params.TimeRange.StartTime != nil && logTime.Before(*params.TimeRange.StartTime) {
		return false
	}
	if params.TimeRange.EndTime != nil && logTime.After(*params.TimeRange.EndTime) {
		return false
	}

	return true
}

// paginateLogs applies pagination to logs
func paginateLogs(logs []*telemetry.LogData, pagination PaginationParams) []*telemetry.LogData {
	if pagination.Offset >= len(logs) {
		return []*telemetry.LogData{}
	}

	end := pagination.Offset + pagination.Limit
	if end > len(logs) {
		end = len(logs)
	}

	return logs[pagination.Offset:end]
}

// FilterMetrics applies all filters to a slice of metrics
func FilterMetrics(metrics []*telemetry.MetricData, params MetricFilterParams) []*telemetry.MetricData {
	filtered := make([]*telemetry.MetricData, 0, len(metrics))

	for _, metric := range metrics {
		if !matchesMetricFilters(metric, params) {
			continue
		}
		filtered = append(filtered, metric)
	}

	return paginateMetrics(filtered, params.Pagination)
}

// matchesMetricFilters checks if a metric matches all filter criteria
func matchesMetricFilters(metric *telemetry.MetricData, params MetricFilterParams) bool {
	// Service filter
	if params.Service != "" {
		serviceName := metric.GetServiceName()
		metricName := metric.GetMetricName()
		target := serviceName + " " + metricName
		if !strings.Contains(strings.ToLower(target), strings.ToLower(params.Service)) {
			return false
		}
	}

	// Metric name filter
	if params.MetricName != "" {
		if !strings.Contains(strings.ToLower(metric.GetMetricName()), strings.ToLower(params.MetricName)) {
			return false
		}
	}

	// Metric type filter
	if params.MetricType != "" {
		if !strings.EqualFold(metric.GetMetricTypeText(), params.MetricType) {
			return false
		}
	}

	// Time range filter
	if params.TimeRange.StartTime != nil && metric.ReceivedAt.Before(*params.TimeRange.StartTime) {
		return false
	}
	if params.TimeRange.EndTime != nil && metric.ReceivedAt.After(*params.TimeRange.EndTime) {
		return false
	}

	return true
}

// paginateMetrics applies pagination to metrics
func paginateMetrics(metrics []*telemetry.MetricData, pagination PaginationParams) []*telemetry.MetricData {
	if pagination.Offset >= len(metrics) {
		return []*telemetry.MetricData{}
	}

	end := pagination.Offset + pagination.Limit
	if end > len(metrics) {
		end = len(metrics)
	}

	return metrics[pagination.Offset:end]
}

// severityNameToNumber converts severity name to number
func severityNameToNumber(name string) int32 {
	switch strings.ToLower(name) {
	case "trace":
		return 1
	case "debug":
		return 5
	case "info":
		return 9
	case "warn":
		return 13
	case "error":
		return 17
	case "fatal":
		return 21
	default:
		return 0
	}
}
