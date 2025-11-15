package httpserver

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
)

type Server struct {
	store *telemetry.Store
	mux   *http.ServeMux
}

func NewServer(store *telemetry.Store) *Server {
	s := &Server{
		store: store,
		mux:   http.NewServeMux(),
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	// Traces endpoints
	s.mux.HandleFunc("GET /api/traces", s.handleGetTraces)
	s.mux.HandleFunc("GET /api/traces/{traceID}", s.handleGetTraceByID)
	s.mux.HandleFunc("GET /api/traces/{traceID}/services/{service}", s.handleGetTraceByIDAndService)
	s.mux.HandleFunc("GET /api/spans/{spanID}", s.handleGetSpanByID)

	// Metrics endpoints
	s.mux.HandleFunc("GET /api/metrics", s.handleGetMetrics)
	s.mux.HandleFunc("GET /api/metrics/{service}", s.handleGetMetricsByService)
	s.mux.HandleFunc("GET /api/metrics/{service}/{metricName}", s.handleGetMetricsByServiceAndName)

	// Logs endpoints
	s.mux.HandleFunc("GET /api/logs", s.handleGetLogs)
	s.mux.HandleFunc("GET /api/logs/trace/{traceID}", s.handleGetLogsByTraceID)

	// Topology endpoint
	s.mux.HandleFunc("GET /api/topology", s.handleGetTopology)

	// Services endpoint
	s.mux.HandleFunc("GET /api/services", s.handleGetServices)

	// Stats endpoint
	s.mux.HandleFunc("GET /api/stats", s.handleGetStats)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Apply permissive CORS headers for external hosting
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH, HEAD")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Expose-Headers", "*")
	w.Header().Set("Access-Control-Max-Age", "86400")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	s.mux.ServeHTTP(w, r)
}

// Trace handlers

func (s *Server) handleGetTraces(w http.ResponseWriter, r *http.Request) {
	// Parse filter parameters
	filterParams := ParseTraceFilterParams(r)

	// Get all spans
	spans := s.store.GetSvcSpans()

	// Apply filters
	filtered := FilterSpans(*spans, filterParams)

	// Convert to JSON
	result := make([]SpanJSON, len(filtered))
	for i, span := range filtered {
		result[i] = SpanDataToJSON(span)
	}

	// Add pagination metadata to response headers
	w.Header().Set("X-Total-Count", strconv.Itoa(len(*spans)))
	w.Header().Set("X-Filtered-Count", strconv.Itoa(len(filtered)))
	w.Header().Set("X-Offset", strconv.Itoa(filterParams.Pagination.Offset))
	w.Header().Set("X-Limit", strconv.Itoa(filterParams.Pagination.Limit))

	respondJSON(w, http.StatusOK, result)
}

func (s *Server) handleGetTraceByID(w http.ResponseWriter, r *http.Request) {
	traceID := r.PathValue("traceID")

	cache := s.store.GetTraceCache()
	spans, ok := cache.GetSpansByTraceID(traceID)
	if !ok {
		respondError(w, http.StatusNotFound, "Trace not found")
		return
	}

	// Get unique services in this trace
	serviceSet := make(map[string]bool)
	spanJSONs := make([]SpanJSON, len(spans))
	for i, span := range spans {
		spanJSONs[i] = SpanDataToJSON(span)
		serviceSet[span.GetServiceName()] = true
	}

	services := make([]string, 0, len(serviceSet))
	for service := range serviceSet {
		services = append(services, service)
	}

	result := TraceJSON{
		TraceID:  traceID,
		Spans:    spanJSONs,
		Services: services,
	}

	respondJSON(w, http.StatusOK, result)
}

func (s *Server) handleGetTraceByIDAndService(w http.ResponseWriter, r *http.Request) {
	traceID := r.PathValue("traceID")
	service := r.PathValue("service")

	cache := s.store.GetTraceCache()
	spans, ok := cache.GetSpansByTraceIDAndSvc(traceID, service)
	if !ok {
		respondError(w, http.StatusNotFound, "Spans not found for trace and service")
		return
	}

	result := make([]SpanJSON, len(spans))
	for i, span := range spans {
		result[i] = SpanDataToJSON(span)
	}

	respondJSON(w, http.StatusOK, result)
}

func (s *Server) handleGetSpanByID(w http.ResponseWriter, r *http.Request) {
	spanID := r.PathValue("spanID")

	cache := s.store.GetTraceCache()
	span, ok := cache.GetSpanByID(spanID)
	if !ok {
		respondError(w, http.StatusNotFound, "Span not found")
		return
	}

	result := SpanDataToJSON(span)
	respondJSON(w, http.StatusOK, result)
}

// Metric handlers

func (s *Server) handleGetMetrics(w http.ResponseWriter, r *http.Request) {
	// Parse filter parameters
	filterParams := ParseMetricFilterParams(r)

	// Get all metrics
	s.store.ApplyFilterMetrics("")
	metrics := s.store.GetFilteredMetrics()

	// Apply filters
	filtered := FilterMetrics(*metrics, filterParams)

	// Convert to JSON
	result := make([]MetricJSON, len(filtered))
	for i, metric := range filtered {
		result[i] = MetricDataToJSON(metric)
	}

	// Add pagination metadata to response headers
	w.Header().Set("X-Total-Count", strconv.Itoa(len(*metrics)))
	w.Header().Set("X-Filtered-Count", strconv.Itoa(len(filtered)))
	w.Header().Set("X-Offset", strconv.Itoa(filterParams.Pagination.Offset))
	w.Header().Set("X-Limit", strconv.Itoa(filterParams.Pagination.Limit))

	respondJSON(w, http.StatusOK, result)
}

func (s *Server) handleGetMetricsByService(w http.ResponseWriter, r *http.Request) {
	service := r.PathValue("service")

	s.store.ApplyFilterMetrics(service)
	metrics := s.store.GetFilteredMetrics()

	result := make([]MetricJSON, len(*metrics))
	for i, metric := range *metrics {
		result[i] = MetricDataToJSON(metric)
	}

	respondJSON(w, http.StatusOK, result)
}

func (s *Server) handleGetMetricsByServiceAndName(w http.ResponseWriter, r *http.Request) {
	service := r.PathValue("service")
	metricName := r.PathValue("metricName")

	cache := s.store.GetMetricCache()
	metrics, ok := cache.GetMetricsBySvcAndMetricName(service, metricName)
	if !ok {
		respondError(w, http.StatusNotFound, "Metrics not found for service and metric name")
		return
	}

	result := make([]MetricJSON, len(metrics))
	for i, metric := range metrics {
		result[i] = MetricDataToJSON(metric)
	}

	respondJSON(w, http.StatusOK, result)
}

// Log handlers

func (s *Server) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	// Parse filter parameters
	filterParams := ParseLogFilterParams(r)

	// Get all logs
	s.store.ApplyFilterLogs("")
	logs := s.store.GetFilteredLogs()

	// Apply filters
	filtered := FilterLogs(*logs, filterParams)

	// Convert to JSON
	result := make([]LogJSON, len(filtered))
	for i, log := range filtered {
		result[i] = LogDataToJSON(log)
	}

	// Add pagination metadata to response headers
	w.Header().Set("X-Total-Count", strconv.Itoa(len(*logs)))
	w.Header().Set("X-Filtered-Count", strconv.Itoa(len(filtered)))
	w.Header().Set("X-Offset", strconv.Itoa(filterParams.Pagination.Offset))
	w.Header().Set("X-Limit", strconv.Itoa(filterParams.Pagination.Limit))

	respondJSON(w, http.StatusOK, result)
}

func (s *Server) handleGetLogsByTraceID(w http.ResponseWriter, r *http.Request) {
	traceID := r.PathValue("traceID")

	cache := s.store.GetLogCache()
	logs, ok := cache.GetLogsByTraceID(traceID)
	if !ok {
		respondError(w, http.StatusNotFound, "Logs not found for trace")
		return
	}

	result := make([]LogJSON, len(logs))
	for i, log := range logs {
		result[i] = LogDataToJSON(log)
	}

	respondJSON(w, http.StatusOK, result)
}

// Topology handler

func (s *Server) handleGetTopology(w http.ResponseWriter, r *http.Request) {
	cache := s.store.GetTraceCache()

	// Build topology from dependency graph
	topology := s.buildTopology(cache)

	respondJSON(w, http.StatusOK, topology)
}

func (s *Server) buildTopology(cache *telemetry.TraceCache) TopologyJSON {
	// Access the internal span map to build dependencies
	// We'll need to get the dependencies similar to how getDependencies works
	nodes := make(map[string]*TopologyNodeJSON)
	edges := make(map[string]*TopologyEdgeJSON)

	// Get all spans and build the graph
	spans := s.store.GetSvcSpans()
	for _, spanData := range *spans {
		span := spanData.Span
		serviceName := spanData.GetServiceName()

		// Add node if not exists
		if _, ok := nodes[serviceName]; !ok {
			nodes[serviceName] = &TopologyNodeJSON{
				Service: serviceName,
				Depth:   0, // Will calculate later
			}
		}

		// Check for parent span
		parentSpanID := span.ParentSpanID().String()
		if parentSpanID != "" && !span.ParentSpanID().IsEmpty() {
			if parentSpan, ok := cache.GetSpanByID(parentSpanID); ok {
				parentServiceName := parentSpan.GetServiceName()

				// Don't create edge if parent and child are same service
				if parentServiceName != serviceName {
					edgeKey := parentServiceName + "->" + serviceName

					// Add parent node if not exists
					if _, ok := nodes[parentServiceName]; !ok {
						nodes[parentServiceName] = &TopologyNodeJSON{
							Service: parentServiceName,
							Depth:   0,
						}
					}

					// Add or increment edge
					if edge, ok := edges[edgeKey]; ok {
						edge.Count++
					} else {
						edges[edgeKey] = &TopologyEdgeJSON{
							Source: parentServiceName,
							Target: serviceName,
							Count:  1,
						}
					}
				}
			}
		}
	}

	// Convert maps to slices
	nodeSlice := make([]TopologyNodeJSON, 0, len(nodes))
	for _, node := range nodes {
		nodeSlice = append(nodeSlice, *node)
	}

	edgeSlice := make([]TopologyEdgeJSON, 0, len(edges))
	for _, edge := range edges {
		edgeSlice = append(edgeSlice, *edge)
	}

	return TopologyJSON{
		Nodes: nodeSlice,
		Edges: edgeSlice,
	}
}

// Services handler

func (s *Server) handleGetServices(w http.ResponseWriter, r *http.Request) {
	serviceSet := make(map[string]bool)

	// Collect services from spans
	spans := s.store.GetSvcSpans()
	for _, span := range *spans {
		serviceSet[span.GetServiceName()] = true
	}

	services := make([]string, 0, len(serviceSet))
	for service := range serviceSet {
		services = append(services, service)
	}

	respondJSON(w, http.StatusOK, services)
}

// Stats handler

func (s *Server) handleGetStats(w http.ResponseWriter, r *http.Request) {
	spans := s.store.GetSvcSpans()

	// Get filtered metrics and logs to get accurate counts
	s.store.ApplyFilterMetrics("")
	metrics := s.store.GetFilteredMetrics()

	s.store.ApplyFilterLogs("")
	logs := s.store.GetFilteredLogs()

	// Count unique traces
	traceSet := make(map[string]bool)
	for _, span := range *spans {
		traceSet[span.Span.TraceID().String()] = true
	}

	// Count unique services
	serviceSet := make(map[string]bool)
	for _, span := range *spans {
		serviceSet[span.GetServiceName()] = true
	}

	stats := StatsJSON{
		SpanCount:           len(*spans),
		MetricCount:         len(*metrics),
		LogCount:            len(*logs),
		TraceCount:          len(traceSet),
		ServiceCount:        len(serviceSet),
		LastUpdated:         s.store.UpdatedAt(),
		MaxServiceSpanCount: telemetry.MAX_SERVICE_SPAN_COUNT,
		MaxMetricCount:      telemetry.MAX_METRIC_COUNT,
		MaxLogCount:         telemetry.MAX_LOG_COUNT,
	}

	respondJSON(w, http.StatusOK, stats)
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// Add Lock/Unlock methods to make the store lockable from outside
// These are convenience methods that wrap the mutex

func (s *Server) getServicesByMetrics() []string {
	s.store.ApplyFilterMetrics("")
	metrics := s.store.GetFilteredMetrics()

	serviceSet := make(map[string]bool)
	for _, metric := range *metrics {
		serviceSet[metric.GetServiceName()] = true
	}

	services := make([]string, 0, len(serviceSet))
	for service := range serviceSet {
		services = append(services, service)
	}
	return services
}

func (s *Server) getServicesByLogs() []string {
	s.store.ApplyFilterLogs("")
	logs := s.store.GetFilteredLogs()

	serviceSet := make(map[string]bool)
	for _, log := range *logs {
		serviceSet[log.GetServiceName()] = true
	}

	services := make([]string, 0, len(serviceSet))
	for service := range serviceSet {
		services = append(services, service)
	}
	return services
}

// getAllServices combines services from traces, metrics, and logs
func (s *Server) getAllServices() []string {
	serviceSet := make(map[string]bool)

	// From traces
	spans := s.store.GetSvcSpans()
	for _, span := range *spans {
		serviceSet[span.GetServiceName()] = true
	}

	// From metrics
	for _, service := range s.getServicesByMetrics() {
		serviceSet[service] = true
	}

	// From logs
	for _, service := range s.getServicesByLogs() {
		serviceSet[service] = true
	}

	services := make([]string, 0, len(serviceSet))
	for service := range serviceSet {
		if service != "" && !strings.Contains(service, "unknown") {
			services = append(services, service)
		}
	}

	return services
}
