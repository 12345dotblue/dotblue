package im

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func FeishuInboundHandler(r *ghttp.Request) {
	connectionID := strings.TrimSpace(r.Get("id").String())
	if connectionID == "" {
		r.Response.WriteStatus(http.StatusBadRequest, "Connection ID is required")
		return
	}

	record, err := defaultConnectionRepository.GetByID(r.Context(), connectionID)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to load connection")
		return
	}
	if record == nil || record.Platform != "feishu" {
		r.Response.WriteStatus(http.StatusNotFound, "Connection not found")
		return
	}

	var payload map[string]any
	if err := r.Parse(&payload); err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, "Invalid payload")
		return
	}

	// Keep challenge handshake compatible for future webhook-style validation.
	if challenge := str(payload["challenge"]); challenge != "" {
		r.Response.WriteJson(g.Map{"challenge": challenge})
		return
	}

	adapter, err := GetAdapter(record.Platform)
	if err != nil {
		r.Response.WriteStatus(http.StatusInternalServerError, "Adapter not available")
		return
	}

	events, err := adapter.ParseInbound(r.Context(), payload)
	if err != nil {
		r.Response.WriteStatus(http.StatusBadRequest, err.Error())
		return
	}

	result, err := defaultInboundPipeline.PersistEvents(r.Context(), toConnection(*record), events)
	if err != nil {
		g.Log().Errorf(r.Context(), "Failed to persist inbound events: %v", err)
		r.Response.WriteStatus(http.StatusInternalServerError, "Failed to persist inbound event")
		return
	}

	r.Response.WriteJson(g.Map{
		"success":    true,
		"accepted":   result.Accepted,
		"duplicated": result.Duplicated,
	})
}
