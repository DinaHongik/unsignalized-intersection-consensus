package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	pb "main/client/proto"
	config "main/config"
	direction "main/config/directionBoolean"

	"google.golang.org/grpc"
)

var GO_SERVER_PORT = config.BasePort

// Defines the gRPC server state, including port, vehicle info, and a mutex for safe concurrent access.
type server struct {
	pb.UnimplementedVehicleServiceServer
	Port    string
	Vehicle *pb.Vehicle
	mu      sync.Mutex
}

// Function name : StartServer
// initializes and launches a gRPC server instance for the given vehicle address.
func StartServer(address int32, direction string, number int32, electionStatus string) (*grpc.Server, string) {
	s := &server{
		Port:    fmt.Sprintf("%d", GO_SERVER_PORT+int(address)),                                                      // 포트 번호를 문자열로 변환
		Vehicle: &pb.Vehicle{Number: number, Address: address, Direction: direction, ElectionStatus: electionStatus}, // 기본 차량 정보로 초기화
	}

	// make TCP listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", GO_SERVER_PORT+int(address)))
	if err != nil {
		log.Printf("failed to listen: %v", err)
		return nil, s.Port
	}

	// make gRPC server
	grpcServer := grpc.NewServer(
		grpc.MaxSendMsgSize(1024*1024*10), // 10MB
		grpc.MaxRecvMsgSize(1024*1024*10), // 10MB
	)
	pb.RegisterVehicleServiceServer(grpcServer, s)

	// start server
	go func() {
		defer lis.Close()
		if err := grpcServer.Serve(lis); err != nil {
			fmt.Printf("failed to serve: %v", err)
		}
	}()
	return grpcServer, s.Port
}

// Function name: ReceiveRequest
// Handles a single voting request and returns an acknowledgment with direction info.
func (s *server) ReceiveRequest(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if req == nil {
		return nil, fmt.Errorf("received nil request")
	}

	if s.Vehicle.SendVotes == 0 {
		// Process the incoming vote
		s.Vehicle.SendVotes = 1

		// Validate direction compatibility and build the response message
		directionStatus := "False"
		if direction.DirectionBoolean(req.Vehicle.Direction, s.Vehicle.Direction) {
			directionStatus = "True"
		}

		response := &pb.Response{
			Message:           fmt.Sprintf("Vote registered from port %s to port %s", req.Port, s.Port),
			Status:            "acknowledged",
			DirectionStatus:   directionStatus,
			ResponseDirection: s.Vehicle.Direction,
			Vehicle:           s.Vehicle,
		}
		return response, nil
	} else {
		response := &pb.Response{
			Message: fmt.Sprintf("Vehicle %d has already voted", s.Vehicle.Number),
			Status:  "ignored",
			Vehicle: s.Vehicle,
		}
		return response, nil
	}
}

// Function name: LeaderElection
// Handles leader election requests and updates vehicle roles based on vote counts and timestamps.
func (s *server) LeaderElection(ctx context.Context, req *pb.Request) (*pb.Response, error) {
	if req.Vehicle.ElectionStatus == "Follower" {
		response := &pb.Response{
			Message: fmt.Sprintf("Vehicle %d is follower", s.Vehicle.Number),
			Status:  "ignored",
			Vehicle: s.Vehicle,
		}
		return response, nil
	}

	if s.Vehicle.ReceiveVotes < req.Vehicle.ReceiveVotes {
		response := &pb.Response{
			Message: fmt.Sprintf("Vehicle %d has already voted", s.Vehicle.Number),
			Status:  "acknowledged",
			Vehicle: s.Vehicle,
		}
		// update this server as follower with newer vote info
		s.Vehicle.ElectionStatus = "Follower"
		s.Vehicle.ReceiveVotes = req.Vehicle.ReceiveVotes
		s.Vehicle.ElectionTime = req.Vehicle.ElectionTime
		return response, nil
	} else {
		if s.Vehicle.ReceiveVotes == req.Vehicle.ReceiveVotes {
			reqTime := req.Vehicle.ElectionTime.AsTime().UnixNano()
			serverTime := s.Vehicle.ElectionTime.AsTime().UnixNano()

			if reqTime > serverTime {
				response := &pb.Response{
					Message: fmt.Sprintf("Vehicle %d has already voted", s.Vehicle.Number),
					Status:  "acknowledged",
					Vehicle: s.Vehicle,
				}
				// request has newer timestamp → this server becomes follower
				s.Vehicle.ElectionStatus = "Follower"
				s.Vehicle.ReceiveVotes = req.Vehicle.ReceiveVotes
				s.Vehicle.ElectionTime = req.Vehicle.ElectionTime
				return response, nil
			} else {
				response := &pb.Response{
					Message: fmt.Sprintf("Vehicle %d has already voted", s.Vehicle.Number),
					Status:  "ignored",
					Vehicle: s.Vehicle,
				}
				// this server has newer timestamp → request vehicle becomes follower
				req.Vehicle.ElectionStatus = "Follower"
				req.Vehicle.ReceiveVotes = s.Vehicle.ReceiveVotes
				req.Vehicle.ElectionTime = s.Vehicle.ElectionTime
				return response, nil
			}
		}
		// this server has more votes → request becomes follower
		response := &pb.Response{
			Message: fmt.Sprintf("Vehicle %d has already voted", s.Vehicle.Number),
			Status:  "ignored",
			Vehicle: s.Vehicle,
		}
		req.Vehicle.ElectionStatus = "Follower"
		req.Vehicle.ReceiveVotes = s.Vehicle.ReceiveVotes
		req.Vehicle.ElectionTime = s.Vehicle.ElectionTime
		return response, nil
	}
}

// Function name: UpdateVoteCount
// Determines the leader by comparing vote counts and timestamps, updating roles for both vehicles.
func (s *server) UpdateVoteCount(ctx context.Context, req *pb.Request) (*pb.Response, error) {

	if s.Vehicle.Number == req.Vehicle.Number {
		s.Vehicle.ReceiveVotes = req.Vehicle.ReceiveVotes
		s.Vehicle.ElectionTime = req.Vehicle.ElectionTime

		return &pb.Response{
			Message: fmt.Sprintf("Vote count updated for vehicle %d.\n", s.Vehicle.Number),
			Status:  "success",
		}, nil
	} else {
		return &pb.Response{
			Message: fmt.Sprintf("Number mismatch. Failed to update vehicle %d.\n", req.Vehicle.Number),
			Status:  "failed",
		}, nil
	}
}
