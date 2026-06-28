package api

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/shagston/routerpilot/internal/app"
	sdkPlanner "github.com/shagston/routerpilot/sdk/planner"
)

const wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

type WSClient struct {
	conn   netConn
	mu     sync.Mutex
	closed bool
	app    *app.App
}

type netConn interface {
	io.ReadWriteCloser
	SetReadDeadline(t time.Time) error
	SetWriteDeadline(t time.Time) error
}

type wsMsg struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type wsCmd struct {
	Intent string         `json:"intent"`
	Args   map[string]any `json:"args"`
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgradeToWS(w, r)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}

	client := &WSClient{
		conn: conn,
		app:  s.App,
	}
	defer client.close()

	go client.writePump(r.Context())
	client.readPump(r.Context())
}

func upgradeToWS(w http.ResponseWriter, r *http.Request) (netConn, error) {
	if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
		return nil, errors.New("not a websocket upgrade")
	}

	key := r.Header.Get("Sec-WebSocket-Key")
	if key == "" {
		return nil, errors.New("missing Sec-WebSocket-Key")
	}

	h := sha1.New()
	h.Write([]byte(key + wsGUID))
	acceptKey := base64.StdEncoding.EncodeToString(h.Sum(nil))

	hj, ok := w.(http.Hijacker)
	if !ok {
		return nil, errors.New("server does not support hijacking")
	}

	conn, bufrw, err := hj.Hijack()
	if err != nil {
		return nil, fmt.Errorf("hijack failed: %w", err)
	}

	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + acceptKey + "\r\n\r\n"
	if _, err := bufrw.WriteString(resp); err != nil {
		conn.Close()
		return nil, err
	}
	if err := bufrw.Flush(); err != nil {
		conn.Close()
		return nil, err
	}

	return &hijackedConn{conn: conn, buf: bufrw}, nil
}

type hijackedConn struct {
	conn net.Conn
	buf  *bufio.ReadWriter
}

func (h *hijackedConn) Read(p []byte) (int, error) {
	return h.buf.Read(p)
}

func (h *hijackedConn) Write(p []byte) (int, error) {
	n, err := h.buf.Write(p)
	if err != nil {
		return n, err
	}
	if err := h.buf.Flush(); err != nil {
		return n, err
	}
	return n, nil
}

func (h *hijackedConn) Close() error {
	return h.conn.Close()
}

func (h *hijackedConn) SetReadDeadline(t time.Time) error {
	return h.conn.SetReadDeadline(t)
}

func (h *hijackedConn) SetWriteDeadline(t time.Time) error {
	return h.conn.SetWriteDeadline(t)
}

func (c *WSClient) readPump(ctx context.Context) {
	defer c.close()

	for {
		if c.closed {
			return
		}

		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		msgType, payload, err := readFrame(c.conn)
		if err != nil {
			if !errors.Is(err, io.EOF) && !strings.Contains(err.Error(), "use of closed") {
				log.Printf("websocket read error: %v", err)
			}
			return
		}

		if msgType != opText {
			continue
		}

		var cmd wsCmd
		if err := json.Unmarshal(payload, &cmd); err != nil {
			c.sendJSON(map[string]any{"type": "error", "data": "invalid JSON"})
			continue
		}

		go c.executeCommand(ctx, cmd)
	}
}

func (c *WSClient) executeCommand(ctx context.Context, cmd wsCmd) {
	c.sendJSON(map[string]any{"type": "status", "data": fmt.Sprintf("executing %s...", cmd.Intent)})

	intent := sdkPlanner.Intent{
		Name:      cmd.Intent,
		Arguments: cmd.Args,
	}

	execCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	execution, err := c.app.ExecuteIntent(execCtx, intent, false)
	if err != nil {
		c.sendJSON(map[string]any{
			"type": "error",
			"data": err.Error(),
		})
		return
	}

	c.sendJSON(map[string]any{
		"type": "result",
		"data": execution,
	})
}

func (c *WSClient) writePump(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if c.closed {
				return
			}
			c.sendPing()
		}
	}
}

func (c *WSClient) sendJSON(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("json marshal error: %v", err)
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := writeFrame(c.conn, opText, data); err != nil {
		log.Printf("websocket write error: %v", err)
		c.closed = true
	}
}

func (c *WSClient) sendPing() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := writeFrame(c.conn, opPing, nil); err != nil {
		c.closed = true
	}
}

func (c *WSClient) close() {
	c.mu.Lock()
	c.closed = true
	c.mu.Unlock()
	c.conn.Close()
}

// WebSocket frame opcodes
const (
	opText   = 1
	opClose  = 8
	opPing   = 9
	opPong   = 10
)

func readFrame(r io.Reader) (opcode byte, payload []byte, err error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(r, header); err != nil {
		return 0, nil, err
	}

	opcode = header[0] & 0x0F
	masked := header[1]&0x80 != 0
	length := int64(header[1] & 0x7F)

	switch {
	case length == 126:
		ext := make([]byte, 2)
		if _, err := io.ReadFull(r, ext); err != nil {
			return 0, nil, err
		}
		length = int64(binary.BigEndian.Uint16(ext))
	case length == 127:
		ext := make([]byte, 8)
		if _, err := io.ReadFull(r, ext); err != nil {
			return 0, nil, err
		}
		length = int64(binary.BigEndian.Uint64(ext))
	}

	var maskKey [4]byte
	if masked {
		if _, err := io.ReadFull(r, maskKey[:]); err != nil {
			return 0, nil, err
		}
	}

	payload = make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return 0, nil, err
	}

	if masked {
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}

	if opcode == opPing {
		// Respond with pong automatically
		return opcode, payload, nil
	}

	return opcode, payload, nil
}

func writeFrame(w io.Writer, opcode byte, payload []byte) error {
	header := []byte{0x80 | opcode} // FIN + opcode

	length := len(payload)
	switch {
	case length <= 125:
		header = append(header, byte(length))
	case length <= 65535:
		header = append(header, 126)
		ext := make([]byte, 2)
		binary.BigEndian.PutUint16(ext, uint16(length))
		header = append(header, ext...)
	default:
		header = append(header, 127)
		ext := make([]byte, 8)
		binary.BigEndian.PutUint64(ext, uint64(length))
		header = append(header, ext...)
	}

	if _, err := w.Write(header); err != nil {
		return err
	}
	if len(payload) > 0 {
		if _, err := w.Write(payload); err != nil {
			return err
		}
	}
	return nil
}
