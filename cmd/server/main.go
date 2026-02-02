package main

import (
	"0TLink/internal/auth"
	"0TLink/internal/tunnel"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

var (
	portMutex sync.Mutex
	nextPort  = 8081
)

func main() {
	tlsConfig, err := auth.GetTLSConfig(
		"certs/server.crt",
		"certs/server.key",
		"certs/ca.crt",
		true,
	)
	if err != nil {
		log.Fatalf("TLS config error: %v", err)
	}

	tlsConfig.MinVersion = tls.VersionTLS13
	tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert

	controlLn, err := tls.Listen("tcp", ":7000", tlsConfig)
	if err != nil {
		log.Fatalf("Control plane error: %v", err)
	}
	defer controlLn.Close()

	log.Println("[Sidecar-Net] Relay active on :7000")
	log.Println("[Sidecar-Net] Waiting for developer agents...")

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	go func() {
		for {
			conn, err := controlLn.Accept()
			if err != nil {
				return
			}
			go handleAgentConnection(conn)
		}
	}()

	<-ctx.Done()
	log.Println("[Sidecar-Net] Shutting down relay server")
}

func handleAgentConnection(conn net.Conn) {
	defer conn.Close()

	tlsConn := conn.(*tls.Conn)
	if err := tlsConn.Handshake(); err != nil {
		log.Printf("mTLS handshake failed: %v", err)
		return
	}

	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return
	}

	clientID := state.PeerCertificates[0].Subject.CommonName

	portMutex.Lock()
	assignedPort := nextPort
	nextPort++
	portMutex.Unlock()

	session, err := tunnel.SetupSession(tlsConn, true, tunnel.DefaultConfig)
	if err != nil {
		log.Printf("Yamux setup failed for %s: %v", clientID, err)
		return
	}

	publicLn, err := net.Listen("tcp", fmt.Sprintf(":%d", assignedPort))
	if err != nil {
		log.Printf("Failed to bind port %d for %s: %v", assignedPort, clientID, err)
		return
	}
	defer publicLn.Close()

	log.Printf("[Sidecar-Net] Agent [%s] Online. Access via :%d", clientID, assignedPort)

	for {
		userConn, err := publicLn.Accept()
		if err != nil {
			if session.IsClosed() {
				log.Printf("[Sidecar-Net] Session ended for %s", clientID)
				return
			}
			continue
		}

		stream, err := session.Open()
		if err != nil {
			userConn.Close()
			if session.IsClosed() {
				return
			}
			continue
		}


		go tunnel.Join(userConn, stream, tunnel.DefaultConfig)
	}
}
