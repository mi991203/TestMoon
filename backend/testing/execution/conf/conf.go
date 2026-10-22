package conf

import (
	p "github.com/mi991203/TestMoon/libs/go/platform"
)

type Config = p.Config

func Load() Config { return p.Load("execution") }
