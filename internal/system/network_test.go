package system_test

import (
	"net"
	"testing"

	"github.com/p-l/fringe/internal/system"
	"github.com/stretchr/testify/assert"
)

func TestFirstLocalIP(t *testing.T) {
	t.Parallel()

	t.Run("returns IPv6 loopback address if no local IP found", func(t *testing.T) {
		t.Parallel()

		ip := system.FirstLocalIP([]net.IP{})
		assert.Equal(t, net.IPv6loopback, ip)
	})

	t.Run("returns first not loopback IP found", func(t *testing.T) {
		t.Parallel()

		firstIP := net.ParseIP("192.168.1.2")

		ips := []net.IP{
			net.IPv6loopback,
			net.ParseIP("127.0.0.1"),
			firstIP,
			net.ParseIP("10.5.5.5"),
		}

		ip := system.FirstLocalIP(ips)
		assert.Equal(t, firstIP, ip)
	})
}
