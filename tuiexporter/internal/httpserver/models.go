package httpserver

import (
	"time"

	"github.com/ymtdzzz/otel-tui/tuiexporter/internal/telemetry"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// SpanJSON represents a span in JSON format
type SpanJSON struct {
	TraceID           string                 `json:"traceId"`
	SpanID            string                 `json:"spanId"`
	ParentSpanID      string                 `json:"parentSpanId,omitempty"`
	Name              string                 `json:"name"`
	Kind              string                 `json:"kind"`
	StartTimeUnixNano int64                  `json:"startTimeUnixNano"`
	EndTimeUnixNano   int64                  `json:"endTimeUnixNano"`
	DurationNano      int64                  `json:"durationNano"`
	DurationText      string                 `json:"durationText"`
	Attributes        map[string]interface{} `json:"attributes"`
	Status            SpanStatusJSON         `json:"status"`
	Events            []SpanEventJSON        `json:"events"`
	Links             []SpanLinkJSON         `json:"links"`
	ServiceName       string                 `json:"serviceName"`
	ResourceAttributes map[string]interface{} `json:"resourceAttributes"`
	ScopeName         string                 `json:"scopeName"`
	ScopeVersion      string                 `json:"scopeVersion"`
	ReceivedAt        time.Time              `json:"receivedAt"`
}

// SpanStatusJSON represents span status
type SpanStatusJSON struct {
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
}

// SpanEventJSON represents a span event
type SpanEventJSON struct {
	Name               string                 `json:"name"`
	TimeUnixNano       int64                  `json:"timeUnixNano"`
	Attributes         map[string]interface{} `json:"attributes"`
	DroppedAttributesCount uint32             `json:"droppedAttributesCount"`
}

// SpanLinkJSON represents a span link
type SpanLinkJSON struct {
	TraceID                string                 `json:"traceId"`
	SpanID                 string                 `json:"spanId"`
	TraceState             string                 `json:"traceState,omitempty"`
	Attributes             map[string]interface{} `json:"attributes"`
	DroppedAttributesCount uint32                 `json:"droppedAttributesCount"`
}

// MetricJSON represents a metric in JSON format
type MetricJSON struct {
	Name               string                 `json:"name"`
	Description        string                 `json:"description,omitempty"`
	Unit               string                 `json:"unit,omitempty"`
	Type               string                 `json:"type"`
	DataPoints         []DataPointJSON        `json:"dataPoints"`
	ServiceName        string                 `json:"serviceName"`
	ResourceAttributes map[string]interface{} `json:"resourceAttributes"`
	ScopeName          string                 `json:"scopeName"`
	ScopeVersion       string                 `json:"scopeVersion"`
	ReceivedAt         time.Time              `json:"receivedAt"`
}

// DataPointJSON represents a generic data point
type DataPointJSON struct {
	Attributes        map[string]interface{} `json:"attributes"`
	StartTimeUnixNano int64                  `json:"startTimeUnixNano,omitempty"`
	TimeUnixNano      int64                  `json:"timeUnixNano"`
	// For Gauge and Sum
	Value *float64 `json:"value,omitempty"`
	// For Histogram
	Count         *uint64   `json:"count,omitempty"`
	Sum           *float64  `json:"sum,omitempty"`
	BucketCounts  []uint64  `json:"bucketCounts,omitempty"`
	ExplicitBounds []float64 `json:"explicitBounds,omitempty"`
	Min           *float64  `json:"min,omitempty"`
	Max           *float64  `json:"max,omitempty"`
	// For Summary
	QuantileValues []QuantileJSON `json:"quantileValues,omitempty"`
	// Flags
	Flags uint32 `json:"flags,omitempty"`
}

// QuantileJSON represents a quantile value
type QuantileJSON struct {
	Quantile float64 `json:"quantile"`
	Value    float64 `json:"value"`
}

// LogJSON represents a log in JSON format
type LogJSON struct {
	TimeUnixNano       int64                  `json:"timeUnixNano"`
	ObservedTimeUnixNano int64                `json:"observedTimeUnixNano"`
	SeverityNumber     int32                  `json:"severityNumber"`
	SeverityText       string                 `json:"severityText"`
	Body               string                 `json:"body"`
	Attributes         map[string]interface{} `json:"attributes"`
	TraceID            string                 `json:"traceId,omitempty"`
	SpanID             string                 `json:"spanId,omitempty"`
	Flags              uint32                 `json:"flags"`
	ServiceName        string                 `json:"serviceName"`
	ResourceAttributes map[string]interface{} `json:"resourceAttributes"`
	ScopeName          string                 `json:"scopeName"`
	ScopeVersion       string                 `json:"scopeVersion"`
	ReceivedAt         time.Time              `json:"receivedAt"`
}

// TraceJSON represents a complete trace with all spans
type TraceJSON struct {
	TraceID  string     `json:"traceId"`
	Spans    []SpanJSON `json:"spans"`
	Services []string   `json:"services"`
}

// TopologyJSON represents service topology
type TopologyJSON struct {
	Nodes []TopologyNodeJSON `json:"nodes"`
	Edges []TopologyEdgeJSON `json:"edges"`
}

// TopologyNodeJSON represents a service node
type TopologyNodeJSON struct {
	Service string `json:"service"`
	Depth   int    `json:"depth"`
}

// TopologyEdgeJSON represents a connection between services
type TopologyEdgeJSON struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Count  int    `json:"count"`
}

// StatsJSON represents store statistics
type StatsJSON struct {
	SpanCount          int       `json:"spanCount"`
	MetricCount        int       `json:"metricCount"`
	LogCount           int       `json:"logCount"`
	TraceCount         int       `json:"traceCount"`
	ServiceCount       int       `json:"serviceCount"`
	LastUpdated        time.Time `json:"lastUpdated"`
	MaxServiceSpanCount int      `json:"maxServiceSpanCount"`
	MaxMetricCount     int       `json:"maxMetricCount"`
	MaxLogCount        int       `json:"maxLogCount"`
}

// Conversion functions

// SpanDataToJSON converts SpanData to SpanJSON
func SpanDataToJSON(sd *telemetry.SpanData) SpanJSON {
	span := sd.Span
	duration := span.EndTimestamp().AsTime().Sub(span.StartTimestamp().AsTime())

	return SpanJSON{
		TraceID:           span.TraceID().String(),
		SpanID:            span.SpanID().String(),
		ParentSpanID:      span.ParentSpanID().String(),
		Name:              span.Name(),
		Kind:              span.Kind().String(),
		StartTimeUnixNano: int64(span.StartTimestamp()),
		EndTimeUnixNano:   int64(span.EndTimestamp()),
		DurationNano:      duration.Nanoseconds(),
		DurationText:      sd.GetDurationText(),
		Attributes:        attributesToMap(span.Attributes()),
		Status: SpanStatusJSON{
			Code:    span.Status().Code().String(),
			Message: span.Status().Message(),
		},
		Events:             eventsToJSON(span.Events()),
		Links:              linksToJSON(span.Links()),
		ServiceName:        sd.GetServiceName(),
		ResourceAttributes: attributesToMap(sd.ResourceSpan.Resource().Attributes()),
		ScopeName:          sd.ScopeSpans.Scope().Name(),
		ScopeVersion:       sd.ScopeSpans.Scope().Version(),
		ReceivedAt:         sd.ReceivedAt,
	}
}

// MetricDataToJSON converts MetricData to MetricJSON
func MetricDataToJSON(md *telemetry.MetricData) MetricJSON {
	metric := md.Metric

	return MetricJSON{
		Name:               metric.Name(),
		Description:        metric.Description(),
		Unit:               metric.Unit(),
		Type:               metric.Type().String(),
		DataPoints:         metricDataPointsToJSON(metric),
		ServiceName:        md.GetServiceName(),
		ResourceAttributes: attributesToMap(md.ResourceMetric.Resource().Attributes()),
		ScopeName:          md.ScopeMetric.Scope().Name(),
		ScopeVersion:       md.ScopeMetric.Scope().Version(),
		ReceivedAt:         md.ReceivedAt,
	}
}

// LogDataToJSON converts LogData to LogJSON
func LogDataToJSON(ld *telemetry.LogData) LogJSON {
	log := ld.Log

	return LogJSON{
		TimeUnixNano:         int64(log.Timestamp()),
		ObservedTimeUnixNano: int64(log.ObservedTimestamp()),
		SeverityNumber:       int32(log.SeverityNumber()),
		SeverityText:         log.SeverityText(),
		Body:                 log.Body().AsString(),
		Attributes:           attributesToMap(log.Attributes()),
		TraceID:              log.TraceID().String(),
		SpanID:               log.SpanID().String(),
		Flags:                uint32(log.Flags()),
		ServiceName:          ld.GetServiceName(),
		ResourceAttributes:   attributesToMap(ld.ResourceLog.Resource().Attributes()),
		ScopeName:            ld.ScopeLog.Scope().Name(),
		ScopeVersion:         ld.ScopeLog.Scope().Version(),
		ReceivedAt:           ld.ReceivedAt,
	}
}

// Helper functions

func attributesToMap(attrs pcommon.Map) map[string]interface{} {
	result := make(map[string]interface{})
	attrs.Range(func(k string, v pcommon.Value) bool {
		result[k] = valueToInterface(v)
		return true
	})
	return result
}

func valueToInterface(v pcommon.Value) interface{} {
	switch v.Type() {
	case pcommon.ValueTypeStr:
		return v.Str()
	case pcommon.ValueTypeInt:
		return v.Int()
	case pcommon.ValueTypeDouble:
		return v.Double()
	case pcommon.ValueTypeBool:
		return v.Bool()
	case pcommon.ValueTypeMap:
		return attributesToMap(v.Map())
	case pcommon.ValueTypeSlice:
		slice := v.Slice()
		result := make([]interface{}, slice.Len())
		for i := 0; i < slice.Len(); i++ {
			result[i] = valueToInterface(slice.At(i))
		}
		return result
	case pcommon.ValueTypeBytes:
		return v.Bytes().AsRaw()
	default:
		return nil
	}
}

func eventsToJSON(events ptrace.SpanEventSlice) []SpanEventJSON {
	result := make([]SpanEventJSON, events.Len())
	for i := 0; i < events.Len(); i++ {
		event := events.At(i)
		result[i] = SpanEventJSON{
			Name:                   event.Name(),
			TimeUnixNano:           int64(event.Timestamp()),
			Attributes:             attributesToMap(event.Attributes()),
			DroppedAttributesCount: event.DroppedAttributesCount(),
		}
	}
	return result
}

func linksToJSON(links ptrace.SpanLinkSlice) []SpanLinkJSON {
	result := make([]SpanLinkJSON, links.Len())
	for i := 0; i < links.Len(); i++ {
		link := links.At(i)
		result[i] = SpanLinkJSON{
			TraceID:                link.TraceID().String(),
			SpanID:                 link.SpanID().String(),
			TraceState:             link.TraceState().AsRaw(),
			Attributes:             attributesToMap(link.Attributes()),
			DroppedAttributesCount: link.DroppedAttributesCount(),
		}
	}
	return result
}

func metricDataPointsToJSON(metric *pmetric.Metric) []DataPointJSON {
	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		return gaugeDataPointsToJSON(metric.Gauge().DataPoints())
	case pmetric.MetricTypeSum:
		return sumDataPointsToJSON(metric.Sum().DataPoints())
	case pmetric.MetricTypeHistogram:
		return histogramDataPointsToJSON(metric.Histogram().DataPoints())
	case pmetric.MetricTypeExponentialHistogram:
		return exponentialHistogramDataPointsToJSON(metric.ExponentialHistogram().DataPoints())
	case pmetric.MetricTypeSummary:
		return summaryDataPointsToJSON(metric.Summary().DataPoints())
	default:
		return []DataPointJSON{}
	}
}

func gaugeDataPointsToJSON(dps pmetric.NumberDataPointSlice) []DataPointJSON {
	result := make([]DataPointJSON, dps.Len())
	for i := 0; i < dps.Len(); i++ {
		dp := dps.At(i)
		var value float64
		if dp.ValueType() == pmetric.NumberDataPointValueTypeInt {
			value = float64(dp.IntValue())
		} else {
			value = dp.DoubleValue()
		}
		result[i] = DataPointJSON{
			Attributes:        attributesToMap(dp.Attributes()),
			StartTimeUnixNano: int64(dp.StartTimestamp()),
			TimeUnixNano:      int64(dp.Timestamp()),
			Value:             &value,
			Flags:             uint32(dp.Flags()),
		}
	}
	return result
}

func sumDataPointsToJSON(dps pmetric.NumberDataPointSlice) []DataPointJSON {
	return gaugeDataPointsToJSON(dps) // Same structure
}

func histogramDataPointsToJSON(dps pmetric.HistogramDataPointSlice) []DataPointJSON {
	result := make([]DataPointJSON, dps.Len())
	for i := 0; i < dps.Len(); i++ {
		dp := dps.At(i)
		count := dp.Count()
		sum := dp.Sum()

		bucketCounts := make([]uint64, dp.BucketCounts().Len())
		for j := 0; j < dp.BucketCounts().Len(); j++ {
			bucketCounts[j] = dp.BucketCounts().At(j)
		}

		explicitBounds := make([]float64, dp.ExplicitBounds().Len())
		for j := 0; j < dp.ExplicitBounds().Len(); j++ {
			explicitBounds[j] = dp.ExplicitBounds().At(j)
		}

		var min, max *float64
		if dp.HasMin() {
			minVal := dp.Min()
			min = &minVal
		}
		if dp.HasMax() {
			maxVal := dp.Max()
			max = &maxVal
		}

		result[i] = DataPointJSON{
			Attributes:        attributesToMap(dp.Attributes()),
			StartTimeUnixNano: int64(dp.StartTimestamp()),
			TimeUnixNano:      int64(dp.Timestamp()),
			Count:             &count,
			Sum:               &sum,
			BucketCounts:      bucketCounts,
			ExplicitBounds:    explicitBounds,
			Min:               min,
			Max:               max,
			Flags:             uint32(dp.Flags()),
		}
	}
	return result
}

func exponentialHistogramDataPointsToJSON(dps pmetric.ExponentialHistogramDataPointSlice) []DataPointJSON {
	result := make([]DataPointJSON, dps.Len())
	for i := 0; i < dps.Len(); i++ {
		dp := dps.At(i)
		count := dp.Count()
		sum := dp.Sum()

		var min, max *float64
		if dp.HasMin() {
			minVal := dp.Min()
			min = &minVal
		}
		if dp.HasMax() {
			maxVal := dp.Max()
			max = &maxVal
		}

		result[i] = DataPointJSON{
			Attributes:        attributesToMap(dp.Attributes()),
			StartTimeUnixNano: int64(dp.StartTimestamp()),
			TimeUnixNano:      int64(dp.Timestamp()),
			Count:             &count,
			Sum:               &sum,
			Min:               min,
			Max:               max,
			Flags:             uint32(dp.Flags()),
		}
	}
	return result
}

func summaryDataPointsToJSON(dps pmetric.SummaryDataPointSlice) []DataPointJSON {
	result := make([]DataPointJSON, dps.Len())
	for i := 0; i < dps.Len(); i++ {
		dp := dps.At(i)
		count := dp.Count()
		sum := dp.Sum()

		quantiles := make([]QuantileJSON, dp.QuantileValues().Len())
		for j := 0; j < dp.QuantileValues().Len(); j++ {
			qv := dp.QuantileValues().At(j)
			quantiles[j] = QuantileJSON{
				Quantile: qv.Quantile(),
				Value:    qv.Value(),
			}
		}

		result[i] = DataPointJSON{
			Attributes:        attributesToMap(dp.Attributes()),
			StartTimeUnixNano: int64(dp.StartTimestamp()),
			TimeUnixNano:      int64(dp.Timestamp()),
			Count:             &count,
			Sum:               &sum,
			QuantileValues:    quantiles,
			Flags:             uint32(dp.Flags()),
		}
	}
	return result
}
