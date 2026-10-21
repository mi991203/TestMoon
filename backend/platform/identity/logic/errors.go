package logic

import (
	"google.golang.org/grpc/status"

	p "github.com/mi991203/TestMoon/libs/go/platform"
)

func Translate(e error) error {
	if _, ok := status.FromError(e); ok {
		return e
	}
	return p.DBError(e)
}
