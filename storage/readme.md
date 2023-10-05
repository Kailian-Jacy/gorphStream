# Multi-Version Storage

A transaction oriented storage. Maintaining a window of multiple versions for each state.

You may want to read resources about TPG before this document.

## Organization

As extension of type is not the core idea, we are hereby implement it as fixed array storage with only integer type.

After commitment of transactions with TPG traversed done, the database looks like:
| VariableIndex | Last Commited Value | 
| - | - |
| 0 | 65536 |
| 1 | 0 |
| 2 | 533 |

However, to support the need of transaction rolling back and traversal of TPG, we provide multiple version before commitments. Halfway during traversal of TPG, it may looks like this:

| VariableIndex | Last Commited | Value after txn1 | Value after txn2 | Value after txn3 | ... |
| - | - | - | - | - | - |
| 0 | 65536 | 65535 | 63356 | 65535 | ... |
| 1 | 0 | - | - | - | - | - | - |
| 2 | 533 | 544 | 555 | - | - |

And when all the transactions are done and value commited, the other versions are discarded, shrinking the table to Commited status.

As in some traversal implementations, version update of different variables are not of the same pace, some states may hold only old versions.

## TPG and State version

There are some lemmas about the TPG and MV storage:
1. When PD of some TPG node is fulfilled, all variables of proper version have been in the storage engine. 
   - Parametic dependency fulfilled, means the related write operation is done, so the needed version is recorded;
   - But for DFS traversal, the required version may not be the latest version. So we need to get the required version from txn timeStamp.
     ```
	 DependencyValue = Storage[DependencyIdx][txnTimeStamp]
     ```
2. When updating a variable with reading itself, it only needs the latest version.  
  ```
   Obviously it won't require future version. 
   As "each state build adjacent next state with older version", if the updating requires older version, it's the same to say we get the future version first. It does not hold.
  ```

## View based Parameter visiting

As visiting value requires version and parameter index. This would incur:
- Multiple copy between operation parameters and storage messaging.
- Confusing indexes for user. User needs to index PD.Param to find out the index for variable in storage engine. 
- Unclean and dangerous visiting interface,

So we provided a delegation as `ParamView` to proxy the value visiting. It handles the Versioning and Index Redirection to provide user a clean view.

```golang
  // Defines the parameter
  events.W{
    ...
    Params: []int{A, B},
  }

  // Do method in operations
  func(target storage.ParamView, params storage.ParamView) {
    A, B := 0, 1
		target.Set(V1)
    AValue := params.Get(A)
    BValue := params.Get(B)
    ...
  }
```