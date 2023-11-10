/*
Copyright 2023 The CloudEvents Authors
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"context"
	"fmt"
	"log"
	"net"

	cepbv2 "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	cepb "github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	cloudevents "github.com/cloudevents/sdk-go/v2"

	"google.golang.org/grpc"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

const (
	MAX_EVENT_CACHE = 100
)

type cloudEventServer struct {
	cepb.UnimplementedCloudEventServiceServer
	EventChan chan *cloudevents.Event
}

func newCloudEventServer() *cloudEventServer {
	return &cloudEventServer{
		EventChan: make(chan *cloudevents.Event, MAX_EVENT_CACHE),
	}
}

func (svr *cloudEventServer) Publish(_ context.Context, pbEvt *cepb.CloudEvent) (*emptypb.Empty, error) {
	evt, err := cepbv2.FromProto(pbEvt)
	if err != nil {
		return nil, fmt.Errorf("failed to convert protobuf to cloudevent: %v", err)
	}

	log.Printf("received event:\n%s", evt)
	svr.EventChan <- evt

	return &emptypb.Empty{}, nil
}

func (svc *cloudEventServer) Subscribe(sub *cepb.Subscription, subServer cepb.CloudEventService_SubscribeServer) error {
	for evt := range svc.EventChan {
		pbEvt, err := cepbv2.ToProto(evt)
		if err != nil {
			return fmt.Errorf("failed to convert cloudevent to protobuf: %v", err)
		}
		log.Printf("sending event:\n%s", evt)
		if err := subServer.Send(pbEvt); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	cepb.RegisterCloudEventServiceServer(grpcServer, newCloudEventServer())
	log.Println("Starting server on port :50051")
	grpcServer.Serve(lis)
}
