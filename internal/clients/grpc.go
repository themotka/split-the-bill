package clients

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	ssov1 "github.com/themotka/proto/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"time"
)

type Client struct {
	api ssov1.OAuthClient
	log *slog.Logger
}

func New(ctx context.Context, log *slog.Logger,
	addr string, timeout time.Duration, retriesCount int) (*Client, error) {
	const op = "grpc.New"

	retryOpts := []retry.CallOption{
		retry.WithMax(uint(retriesCount)),
		retry.WithCodes(codes.DeadlineExceeded, codes.NotFound, codes.Aborted),
		retry.WithPerRetryTimeout(timeout),
	}

	logOpts := []logging.Option{
		logging.WithLogOnEvents(logging.PayloadReceived, logging.PayloadSent),
	}

	cc, err := grpc.DialContext(ctx, addr,
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			retry.UnaryClientInterceptor(retryOpts...),
			logging.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
		))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	log.Info("dialing server", "addr", addr)
	return &Client{
		api: ssov1.NewOAuthClient(cc),
	}, nil
}

func (c *Client) Register(ctx context.Context, email, password string) (*ssov1.RegisterResponse, error) {
	const op = "grpc.Register"

	resp, err := c.api.Register(ctx, &ssov1.RegisterRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return resp, fmt.Errorf("%s: %w", op, err)
	}
	return resp, nil
}

func (c *Client) Login(ctx context.Context, email, password string) (*ssov1.LoginResponse, error) {
	const op = "grpc.Login"

	resp, err := c.api.Login(ctx, &ssov1.LoginRequest{
		AppId:    1,
		Email:    email,
		Password: password,
	})
	if err != nil {
		return resp, fmt.Errorf("%s: %w", op, err)
	}
	return resp, nil
}

func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
