package main

import (
	"context"
	"fmt"
	"main/server"
	"math/rand"
	"sync"
	"time"

	pb "main/client/proto"
	config "main/config"
	direction "main/config/directionBoolean"
	utills "main/utills"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pbtimestamp "google.golang.org/protobuf/types/known/timestamppb"
)

// global variable //
var VEHICLES []int32
var GO_SERVER_PORT = config.BasePort
var TOTAL_VEHICLES int32
var PASS_COUNT int

// Function name: rpcConnectTo
// opens a gRPC connection to the given IP and returns a client with timeout.
func rpcConnectTo(ip string) (pb.VehicleServiceClient, *grpc.ClientConn, context.Context, context.CancelFunc, error) {
	conn, err := grpc.Dial(
		ip,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("did not connect: %v", err)
	}

	c := pb.NewVehicleServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	return c, conn, ctx, cancel, nil
}

// Function name: removeVehiclesIfQuorumReached
// Removes the vehicle and its linked co-vehicles from the VEHICLES list.
func removeVehiclesIfQuorumReached(vehicle *pb.Vehicle) bool {

	// 1) Remove the leader vehicle from the VEHICLES list
	VEHICLES = utills.RemoveValue(VEHICLES, vehicle.Number)
	PASS_COUNT++

	if len(vehicle.Covehicle) > 0 {
		// 2) Remove the first-level co-vehicle (if present)
		firstCovehicle := vehicle.Covehicle[0]
		if utills.Contains(VEHICLES, firstCovehicle.Number) {
			VEHICLES = utills.RemoveValue(VEHICLES, firstCovehicle.Number)
			PASS_COUNT++
		}

		// 3) If the first co-vehicle has sub-co-vehicles, process them
		if firstCovehicle.Covehicle != nil && len(firstCovehicle.Covehicle) > 0 {

			// 4) Remove all sub-level co-vehicles from the VEHICLES list
			for _, subCovehicle := range vehicle.Covehicle[0].Covehicle {
				if !utills.Contains(VEHICLES, subCovehicle.Number) {
					continue
				}
				VEHICLES = utills.RemoveValue(VEHICLES, subCovehicle.Number)
				PASS_COUNT++
			}
		}
	}
	return len(VEHICLES) == 0
}

// Function name: selectRandomVehicles
// Randomly selects n unique vehicles from the list.
func selectRandomVehicles(vehicles []int, n int) []int {
	selected := make([]int, 0, n)
	vehicleCopy := append([]int{}, vehicles...)
	for i := 0; i < n; i++ {
		idx := rand.Intn(len(vehicleCopy))
		selected = append(selected, vehicleCopy[idx])
		vehicleCopy = append(vehicleCopy[:idx], vehicleCopy[idx+1:]...)
	}
	return selected
}

// Function name: main
// Runs the full intersection consensus simulation and logs timing and consensus statistics.
func main() {

	var totalConsensusCount = 0
	var longTimeConsensusCount = 0
	totalStartTime := time.Now()

	// Adjustable simulation parameters
	const NUMBER_OF_TOTAL_VEHICLES = 300
	hvRatio := 0.1
	line := 4
	const VISION_TIME = 500

	totalVehicles := make([]int, NUMBER_OF_TOTAL_VEHICLES)
	for i := 0; i < NUMBER_OF_TOTAL_VEHICLES; i++ {
		totalVehicles[i] = i + 1
	}

	numHV := int(float64(NUMBER_OF_TOTAL_VEHICLES) * hvRatio)
	hvVehicles := selectRandomVehicles(totalVehicles, numHV)

	for len(totalVehicles) > 0 {
		totalConsensusCount++
		TIMEOUT := time.Now()

		// Select the number of vehicles participating in this consensus round.
		// Test Mode A: Random count per round
		var randomNum int32 = int32(rand.Intn(line*4) + 1)
		if randomNum > int32(len(totalVehicles)) {
			randomNum = int32(len(totalVehicles))
		}

		// Test Mode B: Fixed count per round (uncomment to use)
		// var TOTAL_VEHICLES int32 = randomNum
		// var randomNum int32 = line*4
		// if randomNum > int32(len(totalVehicles)) {
		// 	randomNum = int32(len(totalVehicles))
		// }

		var TOTAL_VEHICLES int32 = randomNum
		selectedVehicles := selectRandomVehicles(totalVehicles, int(TOTAL_VEHICLES))

		var RandomByzantine []int32

		for i := 0; i < len(selectedVehicles); i++ {
			VEHICLES = append(VEHICLES, int32(selectedVehicles[i]))
			if utills.ContainsInt(hvVehicles, selectedVehicles[i]) {
				RandomByzantine = append(RandomByzantine, int32(selectedVehicles[i]))
			}
		}

		var STOP_VEHICLES_PASS_TIME int

		for len(VEHICLES) > 0 {
			END_TIMEOUT := time.Now()
			duration := END_TIMEOUT.Sub(TIMEOUT)

			if len(VEHICLES) > 1 && duration-(time.Duration(STOP_VEHICLES_PASS_TIME)*time.Millisecond) >= time.Duration(VISION_TIME)*time.Millisecond {
				longTimeConsensusCount++
				time.Sleep(time.Duration(VISION_TIME) * time.Millisecond)

				for _, i := range VEHICLES {
					if len(VEHICLES) > 0 && !utills.Contains(RandomByzantine, i) {
						VEHICLES = utills.RemoveValue(VEHICLES, i)
						PASS_COUNT++
					}
				}
				break
			}

			var TOTAL_VEHICLES int32 = int32(len(VEHICLES))

			// Quorum based on simple majority (more than half of TOTAL_VEHICLES)
			var QUORUM int32 = (int32(TOTAL_VEHICLES))/2 + 1

			// Quorum based on unanimity (all vehicles must agree)
			// var QUORUM int32 = TOTAL_VEHICLES

			if TOTAL_VEHICLES == 0 {
				return
			}

			if TOTAL_VEHICLES == 1 {
				PASS_COUNT++
				VEHICLES = utills.RemoveValue(VEHICLES, 0)
				break
			}

			if TOTAL_VEHICLES == 2 {

				var VISION = 0

				for _, i := range VEHICLES {
					if utills.Contains(RandomByzantine, i) {
						break
					} else {
						VISION++
					}
				}

				if VISION != 2 {
					var RANDOM_PASS_TIME = 3000
					STOP_VEHICLES_PASS_TIME += RANDOM_PASS_TIME
					time.Sleep(time.Duration(RANDOM_PASS_TIME) * time.Millisecond)
				} else {
					time.Sleep(VISION_TIME * time.Millisecond)
				}

				if VEHICLES[0] == VEHICLES[1] {
					VEHICLES = utills.RemoveValue(VEHICLES, VEHICLES[1])
					PASS_COUNT++
				} else {
					VEHICLES = utills.RemoveValue(VEHICLES, VEHICLES[1])
					PASS_COUNT++
					VEHICLES = utills.RemoveValue(VEHICLES, VEHICLES[0])
					PASS_COUNT++
				}
				break
			}

			if TOTAL_VEHICLES >= 3 {
				PASS_COUNT = 0
				var directions = [12]string{"Rs", "Rl", "Rr", "Ls", "Ll", "Lr", "Ds", "Dl", "Dr", "Us", "Ul", "Ur"}

				var DirectionMap map[int32]string
				var grpcServers []*grpc.Server
				var dataMu sync.Mutex
				var wg sync.WaitGroup

				wg.Add(len(VEHICLES))
				DirectionMap = make(map[int32]string)

				for _, i := range VEHICLES {
					DirectionMap[i] = directions[rand.Intn(len(directions))]
					var electionStatus = "Candidate"

					go func(index int32, direction string, number int32) {
						defer wg.Done()
						grpcServer, _ := server.StartServer(index, direction, number, electionStatus)
						dataMu.Lock()
						grpcServers = append(grpcServers, grpcServer)
						dataMu.Unlock()
					}(i, DirectionMap[i], i)
				}

				wg.Wait()
				wg.Add(int(TOTAL_VEHICLES))

				serverData := make(map[int32]*pb.Vehicle)
				var done = false

				for _, i := range VEHICLES {

					go func(i int32) {
						defer wg.Done()
						for _, j := range VEHICLES {
							if i == j {
								continue
							}
							if utills.Contains(RandomByzantine, j) {
								continue
							}
							if utills.Contains(RandomByzantine, i) {
								continue
							}

							wg.Add(1)
							go func(j int32) {
								defer wg.Done()

								if done {
									return
								}

								time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

								addr := fmt.Sprintf("localhost:%d", GO_SERVER_PORT+int(j))
								client, conn, ctx, cancel, _ := rpcConnectTo(addr)
								defer conn.Close()
								defer cancel()

								r, _ := client.ReceiveRequest(
									ctx,
									&pb.Request{
										Vehicle: &pb.Vehicle{
											Number:    i,
											Address:   i,
											Direction: DirectionMap[i],
										},
										Port:          fmt.Sprintf("%d", GO_SERVER_PORT+int(i)),
										TotalVehicles: TOTAL_VEHICLES,
									},
								)

								if r == nil {
									return
								}

								if done {
									return
								}

								if r.Status == "acknowledged" {
									dataMu.Lock()
									defer dataMu.Unlock()
									if done {
										return
									}

									vehicle, exists := serverData[i]

									if !exists {
										vehicle = &pb.Vehicle{
											Number:         i,
											Address:        i,
											ReceiveVotes:   0,
											ElectionVote:   0,
											ElectionStatus: "Candidate",
										}
									}

									if r.DirectionStatus == "True" {
										var covehicles []*pb.Vehicle
										covehicles = vehicle.Covehicle
										var covehicleCheck = false

										for i := int32(0); i < int32(len(covehicles)); i++ {
											if direction.DirectionBoolean(covehicles[i].Direction, r.ResponseDirection) {
												covehicleCheck = true

												if covehicles[i].Covehicle == nil {
													covehicles[i].Covehicle = []*pb.Vehicle{}
												}
												covehicles[i].Covehicle = append(covehicles[i].Covehicle, r.Vehicle)
											}
										}

										if covehicleCheck == false {
											covehicles = append(covehicles, r.Vehicle)
											vehicle.Covehicle = covehicles
										}
									}

									vehicle.ElectionTime = pbtimestamp.Now()

									go func(vehicle *pb.Vehicle) {
										if done {
											return
										}

										vehicle.ReceiveVotes++
										addr := fmt.Sprintf("localhost:%d", GO_SERVER_PORT+int(vehicle.Address))
										client, conn, ctx, cancel, _ := rpcConnectTo(addr)

										defer conn.Close()
										defer cancel()

										_, _ = client.UpdateVoteCount(ctx, &pb.Request{
											Vehicle: vehicle,
										})

									}(vehicle)

									serverData[i] = vehicle

									if vehicle.ReceiveVotes >= QUORUM-1 {
										var wg sync.WaitGroup
										for _, k := range VEHICLES {
											if k == i {
												continue
											}
											if utills.Contains(RandomByzantine, k) {
												continue
											}

											wg.Add(1)
											go func(k int32) {
												defer wg.Done()
												if done {
													return
												}
												addr := fmt.Sprintf("localhost:%d", GO_SERVER_PORT+int(k))
												client, conn, ctx, cancel, err := rpcConnectTo(addr)
												if err != nil {

													return
												}
												defer conn.Close()
												defer cancel()

												r, err = client.LeaderElection(ctx, &pb.Request{
													Vehicle: vehicle,
												})

												if r == nil {
													return
												}

												if done {
													return
												}

												if r.Status == "acknowledged" {
													vehicle.ElectionVote++
													if vehicle.ElectionVote >= QUORUM-1 && vehicle.ElectionStatus == "Candidate" {
														removeVehiclesIfQuorumReached(vehicle)
														done = true
														return
													}
												} else if r.Status == "ignored" {
													serverData[i] = r.Vehicle
												}

											}(k)
										}

										wg.Wait()
									}
								} else {
									dataMu.Lock()
									vehicle, exists := serverData[i]
									if !exists {
										vehicle = &pb.Vehicle{
											Number:         i,
											Address:        i,
											ReceiveVotes:   0,
											ElectionVote:   0,
											ElectionStatus: "Candidate",
										}
										serverData[i] = vehicle
									}
									dataMu.Unlock()
								}

							}(j)
						}
					}(i)
				}

				wg.Wait()
				dataMu.Lock()

				for _, server := range grpcServers {
					if server != nil {
						server.GracefulStop()
					}
				}

				dataMu.Unlock()

				var NUMBER_OF_PASS_STOP_VEHICLES = rand.Intn(len(RandomByzantine) + 1)

				for i := 1; i <= NUMBER_OF_PASS_STOP_VEHICLES; i++ {
					var PASS_STOP_VEHICLES = RandomByzantine[rand.Intn(len(RandomByzantine))]
					VEHICLES = utills.RemoveValue(VEHICLES, int32(PASS_STOP_VEHICLES))
					PASS_COUNT++
					RandomByzantine = utills.RemoveValue(RandomByzantine, int32(PASS_STOP_VEHICLES))
				}

				if NUMBER_OF_PASS_STOP_VEHICLES >= 1 {
					var RANDOM_PASS_TIME = 3000
					STOP_VEHICLES_PASS_TIME += RANDOM_PASS_TIME
					time.Sleep(time.Duration(RANDOM_PASS_TIME) * time.Millisecond)
				}
			}
		}
		totalVehicles = utills.Difference(totalVehicles, selectedVehicles)

	}
	totalEndTime := time.Now()
	duration := totalEndTime.Sub(totalStartTime)

	fmt.Printf("Total consensus duration: %v\n", duration)
	fmt.Printf("Number of consensus rounds: %v\n", totalConsensusCount)
	fmt.Printf("Rounds exceeding %v ms: %v\n", VISION_TIME, longTimeConsensusCount)
	fmt.Printf("Vision-system consensus percentage: %v%%\n", longTimeConsensusCount*100/totalConsensusCount)
}
