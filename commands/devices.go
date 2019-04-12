package commands

import (
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"context"
	"github.com/opencord/voltha/protos/go/voltha"
	"log"
	"time"
)

func listAllDevices(conn *grpc.ClientConn, args []string) (*CommandResult, error) {
	client := voltha.NewVolthaGlobalServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	adapters, err := client.ListDevices(ctx, &empty.Empty{})
	if err != nil {
		log.Fatalf("NOOOOO: %s\n", err)
	}

	result := CommandResult{
		Format: "table{{ .Id }}\t{{.Type}}\t{{.Root}}\t{{.ParentId}}\t{{.SerialNumber}}\t{{.Vlan}}\t{{.AdminState}}\t{{.OperStatus}}\t{{.ConnectStatus}}",
		Data:   adapters.Items,
	}

	return &result, nil
}

func preprovisionDevice(conn *grpc.ClientConn, args []string) (*CommandResult, error) {
	return nil, nil
}

func enableDevice(conn *grpc.ClientConn, args []string) (*CommandResult, error) {
	return nil, nil
}
