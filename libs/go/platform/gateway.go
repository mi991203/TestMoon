package platform

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"net/http"
)

func NewGateway(c Config) *runtime.ServeMux {
	return runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{MarshalOptions: protojson.MarshalOptions{UseProtoNames: true, EmitUnpopulated: true}}),
		runtime.WithMetadata(func(ctx context.Context, r *http.Request) metadata.MD {
			md := metadata.Pairs("service-key", c.ServiceKey)
			for _, k := range []string{"x-testmoon-principal", "x-testmoon-time", "x-testmoon-signature", "authorization", "request-id", "cookie"} {
				if v := r.Header.Get(k); v != "" {
					md.Set(k, v)
				}
			}
			return md
		}),
		runtime.WithForwardResponseOption(func(ctx context.Context, w http.ResponseWriter, _ proto.Message) error {
			md, ok := runtime.ServerMetadataFromContext(ctx)
			if ok {
				for _, c := range md.HeaderMD.Get("set-cookie") {
					w.Header().Add("Set-Cookie", c)
				}
			}
			return nil
		}),
		runtime.WithOutgoingHeaderMatcher(func(key string) (string, bool) {
			if key == "set-cookie" {
				return "", false
			}
			return runtime.DefaultHeaderMatcher(key)
		}),
	)
}
