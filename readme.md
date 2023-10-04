# GorphStream

This is a WIP program, for now it's a golang implementation of `morphStream`. 

## Overview

TODO.

Rich and detailed documents could be found in each part of this implementation:
- [TPG Construction and Traversal.](tpg/readme.md)
- [Multi-Version Storage engine.](storage/readme.md)

## Progress

- [x] TPG Construction.
- [ ] TPG Traversal.
  - [x] DFS.
  - [ ] Novel NotifyDFS.
    - [x] Coding.
    - [ ] Verification.
- [x] Multi-Version Storage.
- [ ] Fault tolerance and rolling back.
- [ ] Optimization and Benchmark.

## Quick Start

As example provided in `main.go`, fairly intuitive to build your own transactions.

To run the example, simply:

```
go mod init gorphStream && go mod tidy
go run main.go
```

## Credit

Credit to https://github.com/intellistream/MorphStream and Paper "MorphStream: Adaptive Scheduling for Scalable Transactional Stream Processing on Multicores".

Some pictures in this documents are captured from the public paper.
