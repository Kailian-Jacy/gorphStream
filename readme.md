# GorphStream

This is a WIP program, for now it's a golang implementation of the core part of `morphStream`[^1] with some modifications.

[^1]: [SIGMOD] Yancan Mao and Jianjun Zhao and Shuhao Zhang and Haikun Liu and Volker Markl. MorphStream: Adaptive Scheduling for Scalable Transactional Stream Processing on Multicores, SIGMOD, 2023


## Overview

GorphStream, inheriting the basic idea of morphStream, is a lockless **transactional stream processing engine (TSPE)** featuring high throughput.

The main idea is to organize atomic state visiting operations and their time and parameter dependency with a **directed multi-graph**, decomposing and assign operations to each working thread to achieve maximized parallelism with minimized synchronization.

`TODO`

Rich and detailed documents for each part could be found in each part of this implementation:
- [TPG Construction and Traversal.](tpg/readme.md)
- [Multi-Version Storage Engine.](storage/readme.md)


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

## Quick Start

As example provided in `main.go`, fairly intuitive to build your own transactions.

To run the example, simply:

```
go mod init gorphStream && go mod tidy
go run main.go
```
