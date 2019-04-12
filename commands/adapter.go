package commands

import (
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"context"
	"github.com/opencord/voltha/protos/go/voltha"
	"time"
)

func listAllAdapters(conn *grpc.ClientConn, args []string) (*CommandResult, error) {
	client := voltha.NewVolthaGlobalServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	adapters, err := client.ListAdapters(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	result := CommandResult{
		Format: "table{{ .Id }}\t{{.Vendor}}\t{{.Version}}",
		Data:   adapters.Items,
	}

	return &result, nil
}
