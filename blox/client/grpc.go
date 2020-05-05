package client

import (
	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/prysmaticlabs/prysm/validator/flags"
	log "github.com/sirupsen/logrus"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"gopkg.in/urfave/cli.v2"
	"strings"
)

func NewGRPCConnection(ctx *cli.Context) (*grpc.ClientConn, error) {
	maxCallRecvMsgSize := 10 * 5 << 20 // Default 50Mb
	grpcRetries := ctx.Uint(flags.GrpcRetriesFlag.Name)
	endpoint := ctx.String(flags.BeaconRPCProviderFlag.Name)
	grpHeaders := strings.Split(ctx.String(flags.GrpcHeadersFlag.Name), ",")

	md := make(metadata.MD)
	for _, hdr := range grpHeaders {
		if len(hdr) == 0 {
			continue // to avoid unnecessary warnings
		}
		ss := strings.Split(hdr, "=")
		if len(ss) != 2 {
			log.Warnf("Incorrect gRPC header flag format. Skipping %v", hdr)
			continue
		}
		md.Set(ss[0], ss[1])
	}

	opts := []grpc.DialOption{
		grpc.WithInsecure(), // TODO - add other connection options
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxCallRecvMsgSize),
			grpcretry.WithMax(grpcRetries),
			grpc.Header(&md),
		),
		grpc.WithStatsHandler(&ocgrpc.ClientHandler{}),
		grpc.WithStreamInterceptor(middleware.ChainStreamClient(
			//grpc_opentracing.StreamClientInterceptor(),
			//grpc_prometheus.StreamClientInterceptor,
			grpcretry.StreamClientInterceptor(),
		)),
		//grpc.WithUnaryInterceptor(middleware.ChainUnaryClient(
		//	grpc_opentracing.UnaryClientInterceptor(),
		//	grpc_prometheus.UnaryClientInterceptor,
		//	grpc_retry.UnaryClientInterceptor(),
		//	nil, // TODO - GRPC logging
		//)),
	}

	return grpc.DialContext(ctx, endpoint, opts...)
}
