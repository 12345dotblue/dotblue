package im

import "testing"

func TestBindingMatchesEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		binding AgentBinding
		event   InboundEvent
		want    bool
	}{
		{
			name: "group mention default",
			binding: AgentBinding{
				TriggerMode: TriggerModeMentionOnly,
				AllowGroup:  true,
				AllowDM:     true,
			},
			event: InboundEvent{
				ChatType:    "group",
				MentionsBot: true,
			},
			want: true,
		},
		{
			name: "group without mention rejected",
			binding: AgentBinding{
				TriggerMode: TriggerModeMentionOnly,
				AllowGroup:  true,
				AllowDM:     true,
			},
			event: InboundEvent{
				ChatType:    "group",
				MentionsBot: false,
			},
			want: false,
		},
		{
			name: "dm default accepted",
			binding: AgentBinding{
				TriggerMode: TriggerModeMentionOnly,
				AllowGroup:  true,
				AllowDM:     true,
			},
			event: InboundEvent{
				ChatType: "p2p",
			},
			want: true,
		},
		{
			name: "keyword matched",
			binding: AgentBinding{
				TriggerMode: TriggerModeKeyword,
				AllowGroup:  true,
				AllowDM:     true,
				TriggerConfig: map[string]any{
					"keywords": []any{"deploy", "发布"},
				},
			},
			event: InboundEvent{
				ChatType: "group",
				Text:     "请帮我 deploy 一下",
			},
			want: true,
		},
		{
			name: "dm blocked by allow dm",
			binding: AgentBinding{
				TriggerMode: TriggerModeAllMessages,
				AllowGroup:  true,
				AllowDM:     false,
			},
			event: InboundEvent{
				ChatType: "p2p",
				Text:     "hello",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := bindingMatchesEvent(tt.binding, tt.event); got != tt.want {
				t.Fatalf("bindingMatchesEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}
