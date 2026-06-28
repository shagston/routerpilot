package telegram

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	sdkPlanner "github.com/shagston/routerpilot/sdk/planner"
	"github.com/shagston/routerpilot/sdk/types"
)

type Bot struct {
	token   string
	app     intentExecutor
	client  *http.Client
	offset  int64
	mu      sync.Mutex
	stopCh  chan struct{}
	doneCh  chan struct{}
	baseURL string
}

type intentExecutor interface {
	ExecuteIntent(ctx context.Context, intent sdkPlanner.Intent, interactive bool) (*types.Execution, error)
	PreviewPlan(ctx context.Context, intent sdkPlanner.Intent) (types.ContextSnapshot, types.Plan, error)
}

type Update struct {
	UpdateID int64   `json:"update_id"`
	Message  *Message `json:"message"`
}

type Message struct {
	MessageID int64  `json:"message_id"`
	Chat     Chat   `json:"chat"`
	Text     string `json:"text"`
	From     User   `json:"from"`
}

type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

type SendMessageReq struct {
	ChatID    int64  `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

func NewBot(token string, app intentExecutor) *Bot {
	return &Bot{
		token:  token,
		app:    app,
		client: &http.Client{Timeout: 30 * time.Second},
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
		baseURL: fmt.Sprintf("https://api.telegram.org/bot%s", token),
	}
}

func (b *Bot) Start(ctx context.Context) {
	go b.pollLoop(ctx)
}

func (b *Bot) Stop() {
	close(b.stopCh)
	<-b.doneCh
}

func (b *Bot) pollLoop(ctx context.Context) {
	defer close(b.doneCh)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-b.stopCh:
			return
		case <-ticker.C:
			b.poll(ctx)
		}
	}
}

func (b *Bot) poll(ctx context.Context) {
	b.mu.Lock()
	offset := b.offset
	b.mu.Unlock()

	params := url.Values{}
	params.Set("offset", strconv.FormatInt(offset, 10))
	params.Set("timeout", "10")
	params.Set("allowed_updates", `["message"]`)

	updates, err := b.getUpdates(ctx, params)
	if err != nil {
		slog.Error("telegram poll error", "error", err)
		return
	}

	for _, upd := range updates {
		if upd.Message != nil {
			b.handleMessage(ctx, upd.Message)
		}
		if upd.UpdateID >= offset {
			b.mu.Lock()
			b.offset = upd.UpdateID + 1
			b.mu.Unlock()
		}
	}
}

func (b *Bot) getUpdates(ctx context.Context, params url.Values) ([]Update, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", b.baseURL+"/getUpdates?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Ok     bool     `json:"ok"`
		Result []Update `json:"result"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if !result.Ok {
		return nil, fmt.Errorf("telegram API error: %s", string(body))
	}

	return result.Result, nil
}

func (b *Bot) SendMessage(chatID int64, text string) error {
	payload := SendMessageReq{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "HTML",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", b.baseURL+"/sendMessage", strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram send error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func (b *Bot) handleMessage(ctx context.Context, msg *Message) {
	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return
	}

	var reply string

	switch {
	case strings.HasPrefix(text, "/start"), strings.HasPrefix(text, "/help"):
		reply = b.helpText()
	case strings.HasPrefix(text, "/tools"):
		reply = b.toolsText(ctx)
	case strings.HasPrefix(text, "/ping "):
		reply = b.execIntent(ctx, "ping", map[string]any{"target": strings.TrimPrefix(text, "/ping ")})
	case strings.HasPrefix(text, "/plan "):
		reply = b.planIntent(ctx, strings.TrimPrefix(text, "/plan "))
	default:
		reply = b.execIntent(ctx, text, nil)
	}

	if err := b.SendMessage(msg.Chat.ID, reply); err != nil {
		slog.Error("telegram send error", "error", err)
	}
}

func (b *Bot) helpText() string {
	return `<b>RouterPilot Telegram Bot</b>

Available commands:
• <code>/ping &lt;host&gt;</code> — Ping a host
• <code>/plan &lt;intent&gt;</code> — Execute an intent
• <code>/tools</code> — List all tools
• <code>/help</code> — This help
• <code>&lt;intent&gt;</code> — Execute any intent directly

Example: <code>/plan dns.lookup google.com</code>
Example: <code>ping 8.8.8.8</code>
Example: <code>system.info</code>`
}

func (b *Bot) toolsText(ctx context.Context) string {
	return "Use <code>/plan &lt;intent&gt;</code> with one of:\n<code>ping, interface.status, interface.set, ip.show, ip.set, route.show, route.add, system.info, system.uptime, system.reboot, system.logs, system.memory, system.disk, system.processes, dns.lookup, dns.status, dns.flush, wifi.scan, wifi.status, dhcp.leases, firewall.status, firewall.reload, network.traceroute, network.neighbors, network.connections, service.list, service.restart, package.list, vpn.status, bridge.status, diagnose</code>"
}

func (b *Bot) execIntent(ctx context.Context, intentName string, args map[string]any) string {
	if args == nil {
		args = map[string]any{}
	}

	intent := sdkPlanner.Intent{
		Name:      intentName,
		Arguments: args,
	}

	exec, err := b.app.ExecuteIntent(ctx, intent, false)
	if err != nil {
		// Check if it's a safety error that needs confirmation
		return fmt.Sprintf("❌ Error: %v", err)
	}

	if exec == nil || exec.Result == nil {
		return "✅ Done (no result)"
	}

	return formatResult(exec.Result)
}

func (b *Bot) planIntent(ctx context.Context, input string) string {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "Usage: /plan &lt;intent&gt; [args...]\nExample: /plan ping target=8.8.8.8"
	}

	intentName := parts[0]
	args := map[string]any{}

	for _, part := range parts[1:] {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			args[kv[0]] = kv[1]
		}
	}

	if intentName == "ping" && len(args) == 0 && len(parts) == 2 {
		args["target"] = parts[1]
	}

	return b.execIntent(ctx, intentName, args)
}

func formatResult(result *types.ToolResult) string {
	if result == nil {
		return "✅ Done"
	}

	if !result.Success {
		return fmt.Sprintf("❌ Failed: %s", result.Error)
	}

	output := result.Output
	if output == nil {
		return "✅ Success"
	}

	// Format structured output as text
	var lines []string
	for k, v := range output {
		switch val := v.(type) {
		case string:
			lines = append(lines, fmt.Sprintf("<b>%s</b>: %s", k, escapeHTML(val)))
		case float64:
			lines = append(lines, fmt.Sprintf("<b>%s</b>: %.2f", k, val))
		case int:
			lines = append(lines, fmt.Sprintf("<b>%s</b>: %d", k, val))
		case bool:
			lines = append(lines, fmt.Sprintf("<b>%s</b>: %v", k, val))
		case []any:
			lines = append(lines, fmt.Sprintf("<b>%s</b>: %d items", k, len(val)))
		case map[string]any:
			lines = append(lines, fmt.Sprintf("<b>%s</b>: %d fields", k, len(val)))
		default:
			lines = append(lines, fmt.Sprintf("<b>%s</b>: %v", k, val))
		}
	}

	if len(lines) == 0 {
		return "✅ Success"
	}

	return strings.Join(lines, "\n")
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
