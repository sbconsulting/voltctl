/*
 * Copyright 2019-present Ciena Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
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
