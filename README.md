# Order Fulfillment System

## Overview
This project is a lightweight yet high-concurrency-capable emulation of an order fulfillment system designed to manage and optimize the placement of orders across different types of storages(cooler, heater, ambient shelf).

The system dynamically reallocates orders in real time to maintain freshness and operational efficiency, even under high load.

## Architecture
```
.
├── README.md
├── client
│   └── client.go
├── config
│   ├── config.go
│   ├── constant.go
│   └── init.json
├── entity
│   ├── storage.go
│   └── storage_group.go
├── go.mod
├── logic
│   └── fulfilment.go
├── main.go
└── test
    └── fulfillment_test.go
```
    
The system is designed to be modular and extensible. 

The `main` package integrates the fulfillment system with the challenge client, handling command-line arguments and submitting actions to the server.

The `entity` package defines the core data structures, such as `Order`, `Storage`, and `StorageGroup`. 

The `logic` package contains the core logic for processing orders and managing storage.

The `test` package contains the tests for the system.

The `config` package contains the configuration for the system.

The `client` package contains the challenge client.


## How to Build and Run

### Prerequisites
- Go 1.18 or later installed on your machine.

### Building the Program
To build the program, navigate to the root directory of the project and run the following command:

```bash
$ go build -o order-fulfillment
```

This will compile the program and create an executable named `order-fulfillment`.

### Running the Program
To run the program, use the following command:

```bash
$ ./order-fulfillment --auth=<token> --rate=<rate in ms> --min=<min pickup delay in seconds> --max=<max pickup delay in seconds> --seed=<seed>
```
Replace `<token>` with your authentication token.

## How to Run Tests
To run the tests, use the following command:

```bash
$ cd go/test 
$ go test -v
# Run all tests
$ go test -run <test_name:e.g. TestMultipleOrderReallocation>
# Run a specific test
```

## Order Discarding Criteria

When the shelf is full, the system needs to discard an order to make room for new ones. The criteria for selecting an order to discard is based on the remaining freshness of the orders. The system identifies the order with the least remaining freshness and discards it. This approach ensures that the freshest possible orders are retained, maximizing the quality of the orders that are eventually delivered.

### Rationale
The decision to discard the least fresh order is driven by the goal of maintaining the highest possible quality of service. By prioritizing the retention of fresher orders, the system can ensure that customers receive their orders in optimal condition. This strategy also aligns with the dynamic reallocation approach, which aims to optimize the overall freshness of the orders in the system.

By following this approach, the system can adapt to varying order volumes and storage constraints while maintaining a high standard of service quality.

