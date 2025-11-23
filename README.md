# Distributed Consensus for CAV Priority at Unsignalized Intersections

This repository contains a Go/gRPC implementation of a voting-based distributed consensus algorithm that assigns passing priority to Connected Autonomous Vehicles (CAVs) at unsignalized intersections under mixed traffic (CAV + human-driven vehicles, HV).

The code is an experimental prototype that simulates the algorithm proposed in:

> Y. Lee and Y. Yoon, “A Distributed Consensus Algorithm for Prioritizing Autonomous Vehicle Passing at Unsignalized Intersections under Mixed Traffic,” 
under review at IEEE Transactions on Vehicular Technology, 2025.
>

## 1. Problem Setting

At unsignalized intersections, multiple vehicles may arrive at (almost) the same time, and deciding who should pass first is non-trivial, especially when:

- CAVs and human-driven vehicles (HVs) coexist.
- Vehicle-to-vehicle (V2V) communication is asynchronous and unreliable.
- Some vehicles may be unresponsive (crash faults).

The goal of this implementation is to:

- Let CAVs **reach a fast, safe, and consistent agreement** on a passing order via distributed voting.
- **Fallback to a vision-based rule** (license plate lexicographical order) when consensus takes too long.

## 2. Algorithm Overview
<p align="center">
  <img src="https://github.com/user-attachments/assets/4ceb27c1-4e7d-4e61-918d-35d55196d747" width="500" />
  <!-- <img src="https://github.com/user-attachments/assets/efb2ba7a-881f-4a9f-b9a5-c8a9f8f8f713" width="250" /> -->
</p>

The algorithm is inspired by the **Raft** consensus protocol but adapted for vehicular environments:

- Each vehicle starts in a **Candidate** state.
- Vehicles exchange **vote requests** over gRPC.
- A vehicle that gathers votes from a **majority (quorum)** becomes the **Leader** and gains passing priority.
- Other vehicles switch to **Follower** and accept the Leader’s decision.
- If consensus is not reached within a predefined threshold `T_vision`, the system falls back to a **vision-based priority rule** (e.g., license plate order) to ensure liveness.

The implementation focuses on:

- **Safety**: at most one Leader is elected per consensus round (no conflicting decisions).
- **Liveness**: as long as a majority of responsive CAVs exist, the system eventually decides; otherwise, the vision fallback guarantees progress.

## 3. Implementation Details

### 3.1. Technologies

- **Language**: Go
- **RPC Framework**: gRPC over HTTP/2
- **Message Definition**: Protocol Buffers (`.proto`)
- **Concurrency**: goroutines + mutexes

This setup mimics a **V2V communication network** in an intersection, with each vehicle represented as a gRPC server instance.

## 4. How to Run

```bash
# 1. Install dependencies
make iprotoc-gen-go  
export PATH="$PATH:$(go env GOPATH)/bin"
make all

# 2. Run the simulation
make run
```

## 5. Relation to the Paper

This code is an **experimental prototype** of the consensus algorithm described in the paper. It is intended for:

- Illustrating how **Raft-style voting** can be adapted to CAV intersection management.
- Providing a **reproducible simulation** for measuring:
    - Consensus latency
    - Timeout frequency (vision fallback)
    - Behavior under varying CAV/HV ratios and intersection capacities.
    - `link`
