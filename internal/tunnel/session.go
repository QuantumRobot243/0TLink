package tunnel

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/hashicorp/yamux"
)

type Config struct {
	SniffTimeout     time.Duration
	MaxSniffLength   int
	YamuxMaxWindow   int
	LogLevel         int
	AllowedMethods   [][]byte
}

var DefaultConfig = Config{
	SniffTimeout:   2 * time.Second,
	MaxSniffLength: 512,
	YamuxMaxWindow: 256 * 1024,
	LogLevel:       1,
	AllowedMethods: [][]byte{
		[]byte("GET "), []byte("POST "), []byte("PUT "),
		[]byte("DELETE "), []byte("HEAD "), []byte("OPTIONS "),
		[]byte("PATCH "),
	},
}

type Matcher func([]byte) string

func Join(src, dst net.Conn, cfg Config) {
	defer src.Close()
	defer dst.Close()

	br := bufio.NewReader(src)

	src.SetReadDeadline(time.Now().Add(cfg.SniffTimeout))
	header, _ := br.Peek(cfg.MaxSniffLength)
	src.SetReadDeadline(time.Time{})

	protocol := RunMatchers(header, cfg)
	if cfg.LogLevel >= 1 {
		log.Printf("[Sidecar-Net] Protocol Identified: %s | Remote: %s", protocol, src.RemoteAddr())
	}

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		transfer(dst, br, cfg)
	}()

	go func() {
		defer wg.Done()
		transfer(src, dst, cfg)
	}()

	wg.Wait()
}

func transfer(dst io.Writer, src io.Reader, cfg Config) {
	_, err := io.Copy(dst, src)
	if err != nil && cfg.LogLevel >= 2 {
		log.Printf("[Sidecar-Net] Debug: Stream terminated: %v", err)
	}

	if closer, ok := dst.(interface{ CloseWrite() error }); ok {
		closer.CloseWrite()
	} else if closer, ok := dst.(io.Closer); ok {
		closer.Close()
	}
}

func RunMatchers(header []byte, cfg Config) string {
	if len(header) == 0 {
		return "Handshake/Pending"
	}

	matchers := []Matcher{
		MatchTLS,
		func(h []byte) string { return MatchHTTP(h, cfg.AllowedMethods) },
		MatchSSH,
		MatchPostgres,
	}

	for _, matcher := range matchers {
		if result := matcher(header); result != "" {
			return result
		}
	}
	return "TCP/Opaque"
}

func MatchTLS(h []byte) string {
	if len(h) < 3 { return "" }
	if h[0] == 0x16 && h[1] == 0x03 && h[2] <= 0x03 {
		return "TLS/Encrypted"
	}
	return ""
}

func MatchHTTP(h []byte, methods [][]byte) string {
	for _, m := range methods {
		if bytes.HasPrefix(h, m) {
			return "HTTP/1.x"
		}
	}
	return ""
}

func MatchSSH(h []byte) string {
	if bytes.HasPrefix(h, []byte("SSH-")) {
		return "SSH"
	}
	return ""
}

func MatchPostgres(h []byte) string {
	if len(h) >= 8 && bytes.Contains(h, []byte("PostgreSQL")) {
		return "PostgreSQL"
	}
	return ""
}

func SetupSession(conn net.Conn, isServer bool, cfg Config) (*yamux.Session, error) {
	yCfg := yamux.DefaultConfig()
	yCfg.MaxStreamWindowSize = uint32(cfg.YamuxMaxWindow)
	yCfg.EnableKeepAlive = true

	if isServer {
		return yamux.Server(conn, yCfg)
	}
	return yamux.Client(conn, yCfg)
}
