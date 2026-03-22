package reload

import (
	"crypto/sha1"
	"encoding/base64"
	"io"
	"net/http"
	"sync"
)

type Hub struct {
	mu    sync.Mutex
	conns []io.Writer
}

func New() *Hub {
	return &Hub{}
}

func (h *Hub) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Hijack not supported", http.StatusInternalServerError)
			return
		}

		conn, bufReader, err := hijacker.Hijack()
		if err != nil {
			return
		}
		defer conn.Close()

		req, err := http.ReadRequest(bufReader.Reader)
		if err != nil {
			return
		}

		key := req.Header.Get("Sec-WebSocket-Key")
		if key == "" {
			http.Error(w, "Missing Sec-WebSocket-Key", http.StatusBadRequest)
			return
		}

		acceptKey := computeAcceptKey(key)

		resp := "HTTP/1.1 101 Switching Protocols\r\n" +
			"Upgrade: websocket\r\n" +
			"Connection: Upgrade\r\n" +
			"Sec-WebSocket-Accept: " + acceptKey + "\r\n" +
			"\r\n"

		if _, err := conn.Write([]byte(resp)); err != nil {
			return
		}

		h.mu.Lock()
		h.conns = append(h.conns, conn)
		h.mu.Unlock()

		buf := make([]byte, 1)
		for {
			_, err := conn.Read(buf)
			if err != nil {
				break
			}
		}

		h.mu.Lock()
		for i, c := range h.conns {
			if c == conn {
				h.conns = append(h.conns[:i], h.conns[i+1:]...)
				break
			}
		}
		h.mu.Unlock()
	})
}

func (h *Hub) Broadcast() {
	msg := "reload"
	frame := makeWebSocketFrame(msg)
	h.mu.Lock()
	for _, conn := range h.conns {
		conn.Write(frame)
	}
	h.mu.Unlock()
}

func makeWebSocketFrame(msg string) []byte {
	frame := make([]byte, 2+len(msg))
	frame[0] = 0x81
	frame[1] = byte(len(msg))
	copy(frame[2:], msg)
	return frame
}

func computeAcceptKey(key string) string {
	const magic = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	h := sha1.New()
	h.Write([]byte(key + magic))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
