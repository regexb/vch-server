package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/begizi/vch-server/gcp"
	"github.com/begizi/vch-server/luis"
	"github.com/begizi/vch-server/pb"
	"github.com/begizi/vch-server/redis"
	"github.com/begizi/vch-server/tunnel"
	"github.com/begizi/vch-server/voice"
)

const (
	port      = "PORT"
	gRPCPort  = "GRPC_PORT"
	redisAddr = "REDIS_ADDR"
)

func main() {
	port := os.Getenv(port)
	// default for port
	if port == "" {
		port = "8080"
	}

	gRPCPort := os.Getenv(gRPCPort)
	// default for port
	if gRPCPort == "" {
		gRPCPort = "9001"
	}

	redisAddr := os.Getenv(redisAddr)
	// default for redis
	if redisAddr == "" {
		redisAddr = ":6379"
	}

	// Setup Queue
	queue, err := redis.NewRedisQueue(redisAddr)
	if err != nil {
		panic(err)
	}

	// Setup GCP Speech
	client, err := gcp.NewGCPSpeechConv()
	if err != nil {
		panic(err)
	}

	// Setup luis client
	luisClient := luis.NewClient(nil, "fe4586e0-03a9-4fb3-b49a-be7e74b3fc15", "1de93e00db2e4d128168115876e5391e")

	// Context
	ctx := context.Background()

	// Error chan
	errc := make(chan error)

	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.NewContext(logger).With("ts", log.DefaultTimestampUTC)
		logger = log.NewContext(logger).With("caller", log.DefaultCaller)
	}

	// Business domain.
	var voiceService voice.Service
	{
		voiceService = voice.NewBasicService(client, queue, luisClient)
		voiceService = voice.ServiceLoggingMiddleware(logger)(voiceService)
	}

	var voiceEndpoint endpoint.Endpoint
	{
		voiceLogger := log.NewContext(logger).With("method", "Voice")
		voiceEndpoint = voice.MakeVoiceEndpoint(voiceService)
		voiceEndpoint = voice.EndpointLoggingMiddleware(voiceLogger)(voiceEndpoint)
	}

	// Interrupt handler
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// HTTP transport
	go func() {
		var voiceHandler http.Handler
		{
			endpoints := voice.Endpoints{
				VoiceEndpoint: voiceEndpoint,
			}
			logger := log.NewContext(logger).With("transport", "HTTP")
			voiceHandler = voice.MakeVoiceHTTPServer(ctx, endpoints, logger)
		}

		logger.Log("msg", "HTTP Server Started", "port", port)
		errc <- http.ListenAndServe(":"+port, accessControl(voiceHandler))
	}()

	// gRPC transport
	go func() {
		lis, err := net.Listen("tcp", ":"+gRPCPort)
		if err != nil {
			errc <- err
			return
		}
		defer lis.Close()

		s := grpc.NewServer()

		// Mechanical domain.
		var vch pb.VCHServer
		{
			t, err := tunnel.MakeTunnelServer(queue, logger)
			if err != nil {
				errc <- err
				return
			}
			vch = t
		}

		pb.RegisterVCHServer(s, vch)

		logger.Log("msg", "GRPC Server Started", "port", gRPCPort)
		errc <- s.Serve(lis)
	}()

	logger.Log("exit", <-errc)
}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}
