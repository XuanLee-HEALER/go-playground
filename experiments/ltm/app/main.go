package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	ctx := context.Background()
	telemetry, shutdown, err := setupTelemetry(ctx)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = shutdown(context.Background())
	}()

	rand.Seed(time.Now().UnixNano())

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware("ltm-lab"))

	r.GET("/ping", func(c *gin.Context) {
		start := time.Now()
		route := c.FullPath()
		telemetry.Metrics.Inflight.Add(
			c.Request.Context(),
			1,
			metric.WithAttributes(attribute.String("route", route)),
		)
		defer telemetry.Metrics.Inflight.Add(
			c.Request.Context(),
			-1,
			metric.WithAttributes(attribute.String("route", route)),
		)
		handleRequest(c, telemetry, route, start, 0)
	})

	r.GET("/work", func(c *gin.Context) {
		start := time.Now()
		route := c.FullPath()
		telemetry.Metrics.Inflight.Add(
			c.Request.Context(),
			1,
			metric.WithAttributes(attribute.String("route", route)),
		)
		defer telemetry.Metrics.Inflight.Add(
			c.Request.Context(),
			-1,
			metric.WithAttributes(attribute.String("route", route)),
		)

		// Simulate work so latency metrics and traces look interesting.
		sleepMs := randomSleep(200, 1200)
		time.Sleep(time.Duration(sleepMs) * time.Millisecond)
		handleRequest(c, telemetry, route, start, sleepMs)
	})

	r.GET("/slow", func(c *gin.Context) {
		start := time.Now()
		route := c.FullPath()
		telemetry.Metrics.Inflight.Add(
			c.Request.Context(),
			1,
			metric.WithAttributes(attribute.String("route", route)),
		)
		defer telemetry.Metrics.Inflight.Add(
			c.Request.Context(),
			-1,
			metric.WithAttributes(attribute.String("route", route)),
		)

		// Long requests keep inflight > 0 long enough for Prometheus to scrape it.
		sleepMs := randomSleep(6000, 9000)
		time.Sleep(time.Duration(sleepMs) * time.Millisecond)
		handleRequest(c, telemetry, route, start, sleepMs)
	})

	r.GET("/error", func(c *gin.Context) {
		start := time.Now()
		route := c.FullPath()
		telemetry.Metrics.Inflight.Add(
			c.Request.Context(),
			1,
			metric.WithAttributes(attribute.String("route", route)),
		)
		defer telemetry.Metrics.Inflight.Add(
			c.Request.Context(),
			-1,
			metric.WithAttributes(attribute.String("route", route)),
		)

		// Force an error response to demonstrate error traces and logs.
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "simulated failure"})
		handleRequest(c, telemetry, route, start, 0)
	})

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}

func handleRequest(c *gin.Context, telemetry *Telemetry, route string, start time.Time, workMs int) {

	span := trace.SpanFromContext(c.Request.Context())
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	status := c.Writer.Status()
	telemetry.Metrics.Requests.Add(
		c.Request.Context(),
		1,
		metric.WithAttributes(
			attribute.String("route", route),
			attribute.Int("status", status),
		),
	)

	telemetry.Metrics.LatencyMs.Record(
		c.Request.Context(),
		float64(time.Since(start).Milliseconds()),
		metric.WithAttributes(attribute.String("route", route)),
	)

	msg := fmt.Sprintf("route=%s status=%d work_ms=%d trace_id=%s span_id=%s", route, status, workMs, traceID, spanID)
	emitLog(
		c.Request.Context(),
		telemetry.Logger,
		msg,
		log.String("route", route),
		log.Int("status", status),
		log.Int("work_ms", workMs),
		log.String("trace_id", traceID),
		log.String("span_id", spanID),
	)
}

func randomSleep(minMs, maxMs int) int {
	if maxMs <= minMs {
		return minMs
	}
	return rand.Intn(maxMs-minMs+1) + minMs
}
