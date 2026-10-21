package platform

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/mi991203/TestMoon/libs/go/transport"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/soft_delete"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Config struct {
	Name           string
	HTTP           string `env:"HTTP"`
	GRPC           string `env:"GRPC"`
	DSN            string `env:"DSN"`
	ServiceKey     string
	EncryptionKey  string
	Redis          string
	RedisPassword  string
	Identity       string
	Core           string
	Execution      string
	Mock           string
	Kafka          string
	WebOrigin      string
	MockHost       string
	SecureCookie   bool
	TLSCert        string
	TLSKey         string
	TLSCA          string
	OrgConcurrency int
	MaxItems       int
	MaxSteps       int
	MaxRunSeconds  int
	QueueSeconds   int
	RetentionDays  int
}

// Config loading is deterministic: bound environment > selected YAML > defaults.
func Load(name string) Config {
	v := viper.New()
	v.SetEnvPrefix("TESTMOON")
	v.AutomaticEnv()
	file := ""
	for i, a := range os.Args {
		if a == "--config" && i+1 < len(os.Args) {
			file = os.Args[i+1]
		}
	}
	if file != "" {
		v.SetConfigFile(file)
		if err := v.ReadInConfig(); err != nil {
			panic(err)
		}
	}
	defaults := map[string]any{"redis": "127.0.0.1:6379", "identity": "127.0.0.1:9091", "core": "127.0.0.1:9092", "execution": "127.0.0.1:9093", "mock": "127.0.0.1:9094", "kafka": "127.0.0.1:9092", "web_origin": "http://localhost:8080", "mock_host": "mock.localhost", "retention_days": 30, "org_concurrency": 4, "max_items": 100, "max_steps": 500, "max_run_seconds": 3600, "queue_seconds": 300, "log_level": "info"}
	for k, d := range defaults {
		v.SetDefault(k, d)
		_ = v.BindEnv(k)
	}
	c := Config{Name: name, ServiceKey: v.GetString("service_key"), EncryptionKey: v.GetString("encryption_key"), Redis: v.GetString("redis"), RedisPassword: v.GetString("redis_password"), Identity: v.GetString("identity"), Core: v.GetString("core"), Execution: v.GetString("execution"), Mock: v.GetString("mock"), Kafka: v.GetString("kafka"), WebOrigin: v.GetString("web_origin"), MockHost: v.GetString("mock_host"), SecureCookie: v.GetBool("secure_cookie"), TLSCert: v.GetString("tls_cert"), TLSKey: v.GetString("tls_key"), TLSCA: v.GetString("tls_ca"), OrgConcurrency: v.GetInt("org_concurrency"), MaxItems: v.GetInt("max_items"), MaxSteps: v.GetInt("max_steps"), MaxRunSeconds: v.GetInt("max_run_seconds"), QueueSeconds: v.GetInt("queue_seconds"), RetentionDays: v.GetInt("retention_days")}
	for _, f := range []string{"HTTP", "GRPC", "DSN"} {
		key := name + "_" + strings.ToLower(f)
		_ = v.BindEnv(key)
		reflect.ValueOf(&c).Elem().FieldByName(f).SetString(v.GetString(key))
	}
	if len(c.ServiceKey) < 32 {
		panic("TESTMOON_SERVICE_KEY must contain at least 32 characters")
	}
	if len(c.EncryptionKey) != 64 {
		panic("TESTMOON_ENCRYPTION_KEY must be 32 bytes in hex")
	}
	if _, err := hex.DecodeString(c.EncryptionKey); err != nil {
		panic("invalid encryption key")
	}
	if c.OrgConcurrency < 1 || c.OrgConcurrency > 64 || c.MaxItems < 1 || c.MaxItems > 100 || c.MaxSteps < 1 || c.MaxSteps > 500 || c.MaxRunSeconds < 1 || c.MaxRunSeconds > 3600 || c.QueueSeconds < 1 {
		panic("invalid execution limits")
	}
	if c.RetentionDays < 1 {
		panic("retention_days must be positive")
	}
	l, err := logrus.ParseLevel(v.GetString("log_level"))
	if err != nil {
		panic(err)
	}
	logrus.SetLevel(l)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	return c
}

type Base struct {
	ID        string                `gorm:"type:char(36);primaryKey" json:"id"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
	DeletedAt soft_delete.DeletedAt `gorm:"type:bigint;not null;default:0;softDelete:milli;index" json:"-"`
}

func NewBase() Base {
	u, err := uuid.NewV7()
	if err != nil {
		panic(err)
	}
	return Base{ID: u.String()}
}
func DB(c Config, models ...any) *gorm.DB {
	db, err := gorm.Open(mysql.Open(c.DSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent), TranslateError: true})
	if err != nil {
		panic(fmt.Errorf("%s database connection failed", c.Name))
	}
	sql, err := db.DB()
	if err != nil {
		panic(err)
	}
	sql.SetMaxOpenConns(20)
	sql.SetMaxIdleConns(5)
	sql.SetConnMaxLifetime(5 * time.Minute)
	if err = db.AutoMigrate(models...); err != nil {
		panic(fmt.Errorf("%s migration: %w", c.Name, err))
	}
	return db
}
func Redis(c Config) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: c.Redis, Password: c.RedisPassword})
}

var Validate = validator.New(validator.WithRequiredStructEnabled())

func Check(v any) error {
	if err := Validate.Struct(v); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return nil
}
func Bad(s string) error      { return status.Error(codes.InvalidArgument, s) }
func Denied() error           { return status.Error(codes.PermissionDenied, "没有操作权限") }
func NotFound() error         { return status.Error(codes.NotFound, "资源不存在") }
func Conflict(s string) error { return status.Error(codes.Aborted, s) }
func DBError(e error) error {
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return NotFound()
	}
	if errors.Is(e, gorm.ErrDuplicatedKey) {
		return status.Error(codes.AlreadyExists, "记录已存在")
	}
	if e != nil {
		return status.Error(codes.Internal, "数据库操作失败")
	}
	return nil
}
func Decode(value any, out any) error {
	b, e := json.Marshal(transport.Fields(value))
	if e != nil {
		return Bad("参数序列化失败")
	}
	if e = json.Unmarshal(b, out); e != nil {
		return Bad("参数类型错误: " + e.Error())
	}
	return Check(out)
}
func Fields(value any) map[string]any { return transport.Fields(value) }
func Unpack(value proto.Message, out any) error {
	if value == nil {
		return errors.New("empty response")
	}
	b, e := json.Marshal(transport.Fields(value))
	if e != nil {
		return e
	}
	return json.Unmarshal(b, out)
}
func Token() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}
func Hash(s string) string { v := sha256.Sum256([]byte(s)); return hex.EncodeToString(v[:]) }
func Seal(key, s string) string {
	k, _ := hex.DecodeString(key)
	b, _ := aes.NewCipher(k)
	a, _ := cipher.NewGCM(b)
	n := make([]byte, a.NonceSize())
	if _, e := rand.Read(n); e != nil {
		panic(e)
	}
	return "enc1:" + base64.RawStdEncoding.EncodeToString(a.Seal(n, n, []byte(s), nil))
}
func Open(key, s string) (string, error) {
	if !strings.HasPrefix(s, "enc1:") {
		return "", errors.New("unencrypted document")
	}
	k, e := hex.DecodeString(key)
	if e != nil {
		return "", e
	}
	b, e := aes.NewCipher(k)
	if e != nil {
		return "", e
	}
	a, e := cipher.NewGCM(b)
	if e != nil {
		return "", e
	}
	d, e := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(s, "enc1:"))
	if e != nil || len(d) < a.NonceSize() {
		return "", errors.New("invalid encrypted document")
	}
	p, e := a.Open(nil, d[:a.NonceSize()], d[a.NonceSize():], nil)
	return string(p), e
}

type Principal struct {
	UserID        string            `json:"user_id"`
	Username      string            `json:"username"`
	Admin         bool              `json:"admin"`
	Organizations map[string]string `json:"organizations"`
	SessionHash   string            `json:"session_hash"`
	CSRF          string            `json:"csrf"`
}
type principalKey struct{}

func Actor(ctx context.Context) Principal { p, _ := ctx.Value(principalKey{}).(Principal); return p }
func WithActor(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}
func RequireOrg(ctx context.Context, org string) error {
	p := Actor(ctx)
	if org == "" || p.Organizations[org] == "" {
		return Denied()
	}
	return nil
}
func RequireOrgAdmin(ctx context.Context, org string) error {
	if Actor(ctx).Organizations[org] != "admin" {
		return Denied()
	}
	return nil
}
func RequireAdmin(ctx context.Context) error {
	if !Actor(ctx).Admin {
		return Denied()
	}
	return nil
}
func sign(key, p, ts string) string {
	m := hmac.New(sha256.New, []byte(key))
	m.Write([]byte(ts + "." + p))
	return hex.EncodeToString(m.Sum(nil))
}
func PrincipalHeaders(c Config, p Principal) http.Header {
	h := http.Header{}
	b, _ := json.Marshal(p)
	s := base64.RawURLEncoding.EncodeToString(b)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	h.Set("X-Testmoon-Principal", s)
	h.Set("X-Testmoon-Time", ts)
	h.Set("X-Testmoon-Signature", sign(c.ServiceKey, s, ts))
	return h
}
func verifyPrincipal(c Config, md metadata.MD) (Principal, error) {
	var p Principal
	s := First(md, "x-testmoon-principal")
	ts := First(md, "x-testmoon-time")
	sig := First(md, "x-testmoon-signature")
	if s == "" {
		return p, nil
	}
	n, e := strconv.ParseInt(ts, 10, 64)
	if e != nil || time.Since(time.Unix(n, 0)).Abs() > time.Minute || !hmac.Equal([]byte(sig), []byte(sign(c.ServiceKey, s, ts))) {
		return p, Denied()
	}
	b, e := base64.RawURLEncoding.DecodeString(s)
	if e != nil {
		return p, Denied()
	}
	e = json.Unmarshal(b, &p)
	return p, e
}
func First(md metadata.MD, k string) string {
	v := md.Get(k)
	if len(v) > 0 {
		return v[0]
	}
	return ""
}
func Header(ctx context.Context, key string) string {
	md, _ := metadata.FromIncomingContext(ctx)
	return First(md, key)
}
func Outgoing(ctx context.Context, c Config) context.Context {
	md := metadata.Pairs("service-key", c.ServiceKey, "request-id", RequestID(ctx))
	for k, v := range PrincipalHeaders(c, Actor(ctx)) {
		md.Set(strings.ToLower(k), v...)
	}
	return metadata.NewOutgoingContext(ctx, md)
}
func Dial(c Config, addr string) *grpc.ClientConn {
	var opts credentials.TransportCredentials = insecure.NewCredentials()
	if c.TLSCA != "" {
		var e error
		opts, e = credentials.NewClientTLSFromFile(c.TLSCA, "")
		if e != nil {
			panic(e)
		}
	}
	conn, e := grpc.NewClient(addr, grpc.WithTransportCredentials(opts), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(16<<20), grpc.MaxCallSendMsgSize(16<<20)))
	if e != nil {
		panic(e)
	}
	return conn
}

type requestIDKey struct{}

func RequestID(ctx context.Context) string {
	s, _ := ctx.Value(requestIDKey{}).(string)
	if s == "" {
		s = uuid.NewString()
	}
	return s
}
func Serve(c Config, register func(*grpc.Server), gateway func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	interceptor := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		md, _ := metadata.FromIncomingContext(ctx)
		if !hmac.Equal([]byte(First(md, "service-key")), []byte(c.ServiceKey)) {
			return nil, Denied()
		}
		p, e := verifyPrincipal(c, md)
		if e != nil {
			return nil, e
		}
		id := First(md, "request-id")
		if id == "" {
			id = uuid.NewString()
		}
		ctx = context.WithValue(WithActor(ctx, p), requestIDKey{}, id)
		defer func() {
			if recover() != nil {
				logrus.WithFields(logrus.Fields{"request_id": id, "service": c.Name, "method": info.FullMethod}).Error("request panic")
				resp = nil
				err = status.Error(codes.Internal, "服务内部错误")
			}
		}()
		return handler(ctx, req)
	}
	options := []grpc.ServerOption{grpc.UnaryInterceptor(interceptor), grpc.MaxRecvMsgSize(16 << 20), grpc.MaxSendMsgSize(16 << 20)}
	if c.TLSCert != "" {
		creds, e := credentials.NewServerTLSFromFile(c.TLSCert, c.TLSKey)
		if e != nil {
			panic(e)
		}
		options = append(options, grpc.Creds(creds))
	}
	g := grpc.NewServer(options...)
	register(g)
	ln, e := net.Listen("tcp", c.GRPC)
	if e != nil {
		panic(e)
	}
	go func() {
		if err := g.Serve(ln); err != nil {
			logrus.Error("gRPC server stopped")
		}
	}()
	mux := NewGateway(c)
	var tc credentials.TransportCredentials = insecure.NewCredentials()
	if c.TLSCA != "" {
		tc, e = credentials.NewClientTLSFromFile(c.TLSCA, "")
		if e != nil {
			panic(e)
		}
	}
	if e = gateway(ctx, mux, c.GRPC, []grpc.DialOption{grpc.WithTransportCredentials(tc), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(16 << 20))}); e != nil {
		panic(e)
	}
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			w.WriteHeader(200)
			return
		}
		if !hmac.Equal([]byte(r.Header.Get("X-Service-Key")), []byte(c.ServiceKey)) {
			http.Error(w, "forbidden", 403)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 12<<20)
		mux.ServeHTTP(w, r)
	})
	server := &http.Server{Addr: c.HTTP, Handler: h, ReadHeaderTimeout: 10 * time.Second, IdleTimeout: 65 * time.Second}
	go func() {
		if e := server.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
			logrus.Error("HTTP server stopped")
			stop()
		}
	}()
	logrus.WithField("service", c.Name).Info("service started")
	<-ctx.Done()
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdown)
	done := make(chan struct{})
	go func() { g.GracefulStop(); close(done) }()
	select {
	case <-done:
	case <-shutdown.Done():
		g.Stop()
	}
}
