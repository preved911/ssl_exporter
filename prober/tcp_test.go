package prober

import (
	"context"
	"crypto/tls"
	"net"
	"testing"
	"time"

	"github.com/ribbybibby/ssl_exporter/config"
	"github.com/ribbybibby/ssl_exporter/test"

	"github.com/prometheus/client_golang/prometheus"
	pconfig "github.com/prometheus/common/config"
)

// TestProbeTCP tests the typical case
func TestProbeTCP(t *testing.T) {
	server, _, _, caFile, teardown, err := test.SetupTCPServer()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer teardown()

	server.StartTLS()
	defer server.Close()

	module := config.Module{
		TLSConfig: pconfig.TLSConfig{
			CAFile:             caFile,
			InsecureSkipVerify: false,
		},
	}

	registry := prometheus.NewRegistry()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ProbeTCP(ctx, server.Listener.Addr().String(), module, registry); err != nil {
		t.Fatalf("error: %s", err)
	}
}

// TestProbeTCPInvalidName tests hitting the server on an address which isn't
// in the SANs (localhost)
func TestProbeTCPInvalidName(t *testing.T) {
	server, _, _, caFile, teardown, err := test.SetupTCPServer()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer teardown()

	server.StartTLS()
	defer server.Close()

	module := config.Module{
		TLSConfig: pconfig.TLSConfig{
			CAFile:             caFile,
			InsecureSkipVerify: false,
		},
	}

	_, listenPort, _ := net.SplitHostPort(server.Listener.Addr().String())

	registry := prometheus.NewRegistry()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ProbeTCP(ctx, "localhost:"+listenPort, module, registry); err == nil {
		t.Fatalf("expected error but err was nil")
	}
}

// TestProbeTCPServerName tests that the probe is successful when the
// servername is provided in the TLS config
func TestProbeTCPServerName(t *testing.T) {
	server, _, _, caFile, teardown, err := test.SetupTCPServer()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer teardown()

	server.StartTLS()
	defer server.Close()

	host, listenPort, _ := net.SplitHostPort(server.Listener.Addr().String())

	module := config.Module{
		TLSConfig: pconfig.TLSConfig{
			CAFile:             caFile,
			InsecureSkipVerify: false,
			ServerName:         host,
		},
	}

	registry := prometheus.NewRegistry()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ProbeTCP(ctx, "localhost:"+listenPort, module, registry); err != nil {
		t.Fatalf("error: %s", err)
	}
}

// TestProbeTCPExpired tests that the probe fails with an expired server cert
func TestProbeTCPExpired(t *testing.T) {
	server, _, _, caFile, teardown, err := test.SetupTCPServer()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer teardown()

	// Create a certificate with a notAfter date in the past
	certPEM, keyPEM := test.GenerateTestCertificate(time.Now().AddDate(0, 0, -1))
	testcert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf(err.Error())
	}
	server.TLS.Certificates = []tls.Certificate{testcert}

	server.StartTLS()
	defer server.Close()

	module := config.Module{
		TLSConfig: pconfig.TLSConfig{
			CAFile:             caFile,
			InsecureSkipVerify: false,
		},
	}

	registry := prometheus.NewRegistry()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ProbeTCP(ctx, server.Listener.Addr().String(), module, registry); err == nil {
		t.Fatalf("expected error but err is nil")
	}
}

// TestProbeTCPExpiredInsecure tests that the probe succeeds with an expired server cert
// when skipping cert verification
func TestProbeTCPExpiredInsecure(t *testing.T) {
	server, _, _, caFile, teardown, err := test.SetupTCPServer()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer teardown()

	// Create a certificate with a notAfter date in the past
	certPEM, keyPEM := test.GenerateTestCertificate(time.Now().AddDate(0, 0, -1))
	testcert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatalf(err.Error())
	}
	server.TLS.Certificates = []tls.Certificate{testcert}

	server.StartTLS()
	defer server.Close()

	module := config.Module{
		TLSConfig: pconfig.TLSConfig{
			CAFile:             caFile,
			InsecureSkipVerify: true,
		},
	}

	registry := prometheus.NewRegistry()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ProbeTCP(ctx, server.Listener.Addr().String(), module, registry); err != nil {
		t.Fatalf("error: %s", err)
	}

}

// TestProbeTCPStartTLSSMTP tests STARTTLS against a mock SMTP server
func TestProbeTCPStartTLSSMTP(t *testing.T) {
	server, _, _, caFile, teardown, err := test.SetupTCPServer()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer teardown()

	server.StartSMTP()
	defer server.Close()

	module := config.Module{
		TCP: config.TCPProbe{
			StartTLS: "smtp",
		},
		TLSConfig: pconfig.TLSConfig{
			CAFile:             caFile,
			InsecureSkipVerify: false,
		},
	}

	registry := prometheus.NewRegistry()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ProbeTCP(ctx, server.Listener.Addr().String(), module, registry); err != nil {
		t.Fatalf("error: %s", err)
	}
}

// TestProbeTCPStartTLSFTP tests STARTTLS against a mock FTP server
func TestProbeTCPStartTLSFTP(t *testing.T) {
	server, _, _, caFile, teardown, err := test.SetupTCPServer()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer teardown()

	server.StartFTP()
	defer server.Close()

	module := config.Module{
		TCP: config.TCPProbe{
			StartTLS: "ftp",
		},
		TLSConfig: pconfig.TLSConfig{
			CAFile:             caFile,
			InsecureSkipVerify: false,
		},
	}

	registry := prometheus.NewRegistry()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ProbeTCP(ctx, server.Listener.Addr().String(), module, registry); err != nil {
		t.Fatalf("error: %s", err)
	}
}

// TestProbeTCPStartTLSIMAP tests STARTTLS against a mock IMAP server
func TestProbeTCPStartTLSIMAP(t *testing.T) {
	server, _, _, caFile, teardown, err := test.SetupTCPServer()
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer teardown()

	server.StartIMAP()
	defer server.Close()

	module := config.Module{
		TCP: config.TCPProbe{
			StartTLS: "imap",
		},
		TLSConfig: pconfig.TLSConfig{
			CAFile:             caFile,
			InsecureSkipVerify: false,
		},
	}

	registry := prometheus.NewRegistry()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := ProbeTCP(ctx, server.Listener.Addr().String(), module, registry); err != nil {
		t.Fatalf("error: %s", err)
	}
}
