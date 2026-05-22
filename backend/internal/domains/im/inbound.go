package im

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func PlatformInboundHandler(r *ghttp.Request) {
	handleInbound(r, strings.TrimSpace(r.Get("platform").String()))
}

func FeishuInboundHandler(r *ghttp.Request) {
	handleInbound(r, "feishu")
}

func handleInbound(r *ghttp.Request, defaultPlatform string) {
	connectionID := strings.TrimSpace(r.Get("id").String())
	platform := firstNonEmpty(strings.TrimSpace(r.Get("platform").String()), strings.TrimSpace(defaultPlatform))
	if connectionID == "" || platform == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "Connection ID and platform are required")
		return
	}

	conn, err := defaultConnectionService.GetConnectionByPlatform(r.Context(), connectionID, platform)
	if err == ErrConnectionNotFound {
		r.Response.WriteStatus(http.StatusNotFound, "Connection not found")
		return
	}
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to load connection")
		return
	}

	var payload map[string]any
	if err := r.Parse(&payload); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, "Invalid payload")
		return
	}

	adapter, err := GetAdapter(conn.Platform)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Adapter not available")
		return
	}

	var events []InboundEvent
	if webhookAdapter, ok := adapter.(InboundWebhookAdapter); ok {
		result, err := webhookAdapter.HandleInboundWebhook(r.Context(), InboundWebhookRequest{
			Connection: conn,
			Headers:    collectInboundHeaders(r),
			Payload:    payload,
		})
		if err != nil {
			r.Response.WriteStatus(http.StatusBadRequest, err.Error())
			return
		}
		if result != nil {
			if result.ImmediateResponse != nil {
				r.Response.WriteJson(result.ImmediateResponse)
				return
			}
			events = result.Events
		}
	} else {
		events, err = adapter.ParseInbound(r.Context(), payload)
		if err != nil {
			r.Response.WriteStatus(http.StatusBadRequest, err.Error())
			return
		}
	}

	writeInboundPipelineResponse(r, conn, events)
}

func writeInboundPipelineResponse(r *ghttp.Request, conn Connection, events []InboundEvent) {
	result, err := defaultInboundPipeline.PersistEvents(r.Context(), conn, events)
	if err != nil {
		g.Log().Errorf(r.Context(), "Failed to persist inbound events: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to persist inbound event")
		return
	}

	r.Response.WriteJson(map[string]any{
		"success":    true,
		"accepted":   result.Accepted,
		"duplicated": result.Duplicated,
	})
}

func collectInboundHeaders(r *ghttp.Request) map[string]string {
	headers := make(map[string]string, len(r.Header))
	for key, values := range r.Header {
		headers[key] = strings.Join(values, ",")
	}
	return headers
}
