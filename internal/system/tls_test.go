package system_test

import (
	"crypto/tls"
	"net"
	"testing"

	"github.com/p-l/fringe/internal/system"
	"github.com/stretchr/testify/assert"
)

func TestTLSConfigWithSelfSignedCert(t *testing.T) {
	t.Parallel()

	t.Run("TLS config is for TLS 1.2 minimum", func(t *testing.T) {
		t.Parallel()

		ips := []net.IP{net.ParseIP("127.0.0.1")}
		tlsConfig := system.TLSConfigWithSelfSignedCert(ips)

		// Creates config with TLS 1.2 as mininmum version
		assert.Equal(t, tls.VersionTLS12, int(tlsConfig.MinVersion))
	})

	t.Run("Make sure there's at lest one IP", func(t *testing.T) {
		t.Parallel()

		var ips []net.IP
		tlsConfig := system.TLSConfigWithSelfSignedCert(ips)

		assert.Nil(t, tlsConfig)
	})

	t.Run("Set certificates in TLS config", func(t *testing.T) {
		t.Parallel()

		ips := []net.IP{net.ParseIP("127.0.0.1")}
		tlsConfig := system.TLSConfigWithSelfSignedCert(ips)

		assert.NotEmpty(t, tlsConfig.Certificates)
	})
}
