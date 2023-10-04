# Transaction Processing Graph
---

Using Graph to schedule the execution of multiple threads is a great idea. Since:
1. **Paths join like threads syncs**. Joint nodes (Synchronize) in some certain points while most of the time running async (Different paths).
2. **Expressive for dependency and possibility**. Branches are suggesting possibilities of different scheduling strategies; connected edges constraints and navigates.
3. **Markable Map**. Graph is suitable to make all kinds of marks to guide the traversal, which can be used to construct some kind of sequence.

TPG is a model using graph to guide scheduling operations.

## Construction

![Credit to paper "MorphStream: Adaptive Scheduling for Scalable Transactional Stream Processing on Multicores"](../assets/20231004203224.png)

The system assumes a series of transactions streaming in.
```golang
ReceiveStreaming <- []txn{
	[]Operations
	TimeStamp
}
```

## Traversal

After construction of TPG, we need to traverse in parallel and operates each node. These nodes (Operations) are connected by some edges suggesting the dependency. Including:
- TD (Transactional timestamp).
- PD (Parameter dependency. e.g. we need to perform `Write(B, f(A))`, then we need some version of A to get `f(A)` to update B. PD marks this dependency) .
- LD (Transactional dependency). Marking variables that need to be rolled back in protection of transactional semantics.

And we mark these nodes in four status:
- `RDY`: ready to be operated but not yet.
- `EXE`: Operation done.
- `BLK`: Dependency not fulfilled. It would be turned into `RDY` after all the dependency executed.
- `ABT`: Failure happens. Requires to be rolled back.

To avoid repeating node, we added the fifth status `OCCUPIED` to prevent the `RDY` nodes to be pushed to stack more than once. This is differ from original morhStream.

We are going to start from some points to traverse through this TPG, transitioning status of each node. 

After successful construction, there are three kinds of TPG nodes in the graph:
```
(TD, RDY). Status is Ready, and would be visited by TD;
(TD, RD, BLK). Status is blocked, can be visited by TD and RD;
(RD, BLK). Status is blocked, can be visited by TD and RD;
```

I'm introducing two methods of traversal. Before that, I need to prove some lemmas.
### Lemmas For TPG

They are quite intuitive:

1. The node we can start from features:
   - Must be the first node in the list of each status.
   - Dependency must be empty.
2. From the starter nodes above, each node can be visited through PD and TD with proper DFS.
	```
	IF [there are some isolated node of status X] then
	1. It's the first Operation of Status X with No dependency: It would be visited as Starter.
	1. It's the first Operation of Status X with dependency: 
	2. It's the later operation of status X: It would be visited as its TD parent would
	```
3. There could be more than one path from one node to another, but there is no circle if the transaction itself is executable.
	- Reason: TD and PD are both directed from the earliers to the laters. The only possibility of loop is multiple operations depends on each other in one single transaction, which is the user's responsibility to write logical transactions.
	- So we need to prevent repeated visit but no need to worry about dead lock or dead loop.

### Pure DFS

Quite intuitive. Explanation omitted.

Points:
- `PUSH`/`POP` could be expensive. Be sure to execute each node we `PUSH`ed.

Cost:
- Stack Space and `PUSH`/`POP` cost. The stack could not be avoided for the possibilities of graph traversal.
  - `TODO`: Reallocation of quite long stack may be optimized with linked list or something. Empirical size of stack could be predefined to avoid copy expansion.
- Visiting memory each movement to check the status. (According to morphStream paper)
  - `TODO`: Actually a little bit confused..

Drawbacks:
- Imbalance between threads. Length of status operations is unpredictable and varies a lot.

### Notified DFS

A modification of DFS that allieviate imbalance of threads through communication.

Idea: 
- When a thread is traversing to some `BLK` nodes, it could choose to register a channel to subscribe the later notification of the nodes' being ready. So, when the node is turned to `RDY`, the node could choose to be executed in which thread.
- Transfering one node could be transfering a series of later nodes. Theoretically, I believe it would work well in most cases.

Details:
- When to register the channel? 
  - Depends on the burden. i.e. the length of stack. Shorter stack means higher possibility to be idle and wait for transfering.
- When the node is ready, how to decide dispose locally or send to the subscriber?
  - Always try to send to the subscriber. If subscriber accept, it's making use of a idle thread. And cost is low.

Cost:
- Almost zero. Fractional cost including trying out send through channel and subscribe by putting the pointer of channel to the node on TPG.

## Limitations

### High latency introduced by Batch disposal. 

Now we construct TPG in punctuation style, waiting for some interval and then perform decomposing in **Batches**. 

This brought us some troubles like:
1. We need to manually maintain the hyperparameter `punctuation` .
2. Introduce latency for fixed punctuation. Despite its huge throughput, it's not so adaptable in network conditions with low latency demands.

However, as the fact below, we can't choose to build it `streamingly` if blueprint not changing significantly:
1. For some reason like network delay, thread asynchronicity, the txns are arriving in times. And we need to finish the disposal after the last one arrives, which can't be shortened.