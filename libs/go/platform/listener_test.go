package platform

import "testing"

func TestLocalGRPCEndpoint(t *testing.T) {
	for _, tt := range []struct{ address, want string }{
		{"0.0.0.0:9091", "127.0.0.1:9091"},
		{":9092", "127.0.0.1:9092"},
		{"[::]:9093", "[::1]:9093"},
		{"127.0.0.1:9091", "127.0.0.1:9091"},
		{"[::1]:9093", "[::1]:9093"},
		{"identity:9091", "identity:9091"},
		{"invalid", "invalid"},
	} {
		t.Run(tt.address, func(t *testing.T) {
			if got := localGRPCEndpoint(tt.address); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
