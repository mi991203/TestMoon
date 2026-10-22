package conf

import (
	"os"

	p "github.com/mi991203/TestMoon/libs/go/platform"
)

type Config = p.Config

func Load() Config { return p.Load("mock") }

func PublicHTTPAddress() string {
	if addr := os.Getenv("TESTMOON_MOCK_PUBLIC_HTTP"); addr != "" {
		return addr
	}
	return "127.0.0.1:8095"
}
