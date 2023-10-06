# gorphStream

It's a golang implementation of the core part of `morphStream`[^1] with some modifications.

[^1]: [SIGMOD] Yancan Mao and Jianjun Zhao and Shuhao Zhang and Haikun Liu and Volker Markl. MorphStream: Adaptive Scheduling for Scalable Transactional Stream Processing on Multicores, SIGMOD, 2023


## Overview

GorphStream, inheriting the basic idea of morphStream, is a lockless **Transactional Stream Processing Engine (TSPE)** featuring high throughput.

The main idea is to organize atomic state visiting operations and their time/parameter dependency with a **directed multi-graph**, decomposing and assigning operations to each working thread to achieve maximized parallelism with minimized synchronization.


There are several parts in directories:
- **Events** under `gorphStream/events`. This defines the basic allowed operations and transactions to provide user logic coding. When `gorphStream` gets some events, it would decompose them into operations for later building TPG.
- **Task Precedence Graph (TPG)** under `gorphStream/tpg`. With a bunch of events arrival, `gorphStream` map the dependency of operations to this TPG. Analysis and scheduling would be done when constructing and traversing this TPG.
  - [TPG Construction and Traversal.](tpg/readme.md)
- **Multi-version Storage Engine (MV-Store)** under `gorphStream/storage`. This is a  storage engine providing log-based history of states and rolling back. This helps to provide transactional semantics with commitment after successful transactions and rolling back after failures.
  - [Multi-Version Storage Engine.](storage/readme.md)

Rich and detailed documents for each part could be found in each part of this implementation.


## Quick Start

As example provided in `main.go`, fairly intuitive to build your own transactions. 

This example demonstrates transferring funds between accounts A and B. What coder does is:
1. **Define state schema**: We defined two states as A and B's deposit to init the storage.
2. **Define the operations**: Register callbacks in operations to define transfering and deposit.
3. **Define the transactions**: Combine the operations into transactions.
4. **Trigger start**: Send the shuffled transaction batch to `gorphStream` and execute.

```golang
const BANKER_SCHEMA = 2

// The part of Callbacks providing `Operation.Do()`

// Check the Balance in the other account and write to this one.
var transferReceive = func(target storage.ParamView, params storage.ParamView) error {
	if params.Get(0) >= V2 {
		total := params.Get(1) + V2
		target.Set(total)
		return nil
	} else {
		return errors.New("insufficient balance")
	}
}

// Other operation callbacks.
...

// Assemble into transactions.
var TransferA2BTxn = events.Txn{
	Ops: []events.Operation{
		// B Add.
		&events.W{
			Name:   "B receive from A",
			Params: []int{A, B},
			Target: B,
			Do:     transferReceive,
		},
		// A decrease.
		&events.W{
			Name:   "A send to B",
			Params: []int{A},
			Target: A,
			Do:     transferSend,
		},
	},
	Timestamp: int64(222),
}

// Other transactions.
...

func main(){
	storage.Init(cmd.BANKER_SCHEMA)
	tpg.Construct(
		// Shuffled input.
		[]*events.Txn{
			&cmd.TransferA2BTxn,
			&cmd.TransferB2ATxn,
			&cmd.DepositTxn,
		},
	).Handle()
	storage.Dump()
}

```

To run the example, simply:

```bash
go mod init gorphStream && go mod tidy
go run main.go
```

The output shall be:

```bash
Storage 0:500
Storage 1:0
```

## Progress

- [ ] TPG Construction.
  - [x] Coding and testing.
  - [ ] (Optional) Parallelizing.
- [ ] TPG Traversal.
  - [x] DFS.
  - [ ] Novel NotifyDFS.
    - [x] Coding.
    - [ ] Verification.
- [x] Multi-Version Storage.
- [ ] Transaction abortion and rolling back.
- [ ] Optimization and Benchmark.
