package ioc

import (
	"coca-ai/internal/config"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// InitPrometheus 初始化 Prometheus 监控中间件
func InitPrometheus(server *gin.Engine) {
	// 暴露 metrics 接口
	server.GET("/metrics", gin.WrapH(promhttp.Handler()))
	log.Println("[Observability] Prometheus metrics exposed at /metrics")
}

// InitJaeger 初始化 Jaeger 链路追踪
func InitJaeger() {
	cfg := config.Get()
	if cfg.Jaeger.AgentHost == "" {
		log.Println("[Observability] Jaeger not configured, skipping")
		return
	}

	// 创建 Jaeger Exporter
	exp, err := jaeger.New(jaeger.WithAgentEndpoint(
		jaeger.WithAgentHost(cfg.Jaeger.AgentHost),
		jaeger.WithAgentPort(cfg.Jaeger.AgentPort),
	))
	if err != nil {
		log.Printf("[Observability] Failed to create Jaeger exporter: %v", err)
		return
	}

	// 创建 TraceProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("coca-backend"),
			semconv.DeploymentEnvironmentKey.String("production"),
		)),
	)

	// 设置全局 TraceProvider
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	log.Printf("[Observability] Jaeger tracing enabled (Host: %s:%s)",
		cfg.Jaeger.AgentHost, cfg.Jaeger.AgentPort)
}
