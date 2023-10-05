package tpg

import (
	"fmt"
	"gorphStream/storage"
)

const (
	DFS = iota + 1
	DFSNotify
)

type Traversor interface {
	Traversal(*TpgNode)
	StopExecutor() chan bool
}

type DFSExecutor struct {
	Idx      int
	Tpg      *TpgMeta
	Stack    []*TpgNode
	ErrStack []*TpgNode

	// Receive ending from the main thread.
	Ending chan bool
}

type DFSNotifyExecutor struct {
	Idx      int
	Tpg      *TpgMeta // Visit shared variables.
	Stack    []*TpgNode
	ErrStack []*TpgNode

	// Receive ending from the main thread.
	Ending chan bool

	// The threshold length of stack.
	Threshold int
	Notifier  chan *TpgNode
}

// Launch threads to perform running.
func (m *TpgMeta) Handle() {
	// For single thread, start traversal and dispose.
	for idx, s := range m.Starter {
		go m.Executor[idx].Traversal(s)
	}
	for {
		DoneIdx := <-m.Finish
		m.Done[DoneIdx] = !m.Done[DoneIdx]
		// Check if all done.
		allDone := true
		for i := 0; i < len(m.Starter); i++ {
			if !m.Done[i] {
				allDone = false
				break
			}
		}
		if allDone {
			for _, e := range m.Executor {
				e.StopExecutor() <- true
			}
			break
		}
	}
}

/*
A single thread execute to traverse through.

This traversal features:
1. Each node would be traversed only once. BLK -> RDY -> Occupied -> EXE.
*/
func (e *DFSExecutor) Traversal(node *TpgNode) {
	for {
		if node == nil {
			// Select from the waiting list Stack.
			// The node inside is guaranteed to be RDY(Occupied) but not BLK or EXE.
			if len(e.Stack) == 0 {
				// Traversal done for this thread.
				e.Tpg.Finish <- e.Idx
				<-e.StopExecutor()
				break
			} else {
				node = e.Stack[len(e.Stack)-1]
				e.Stack = e.Stack[:len(e.Stack)-1] // Pop the last one.
			}
		}
		// Execute Current. The parameter version is provided by node context.
		if err := node.Opt.Execute(&node.ParamVersions); err != nil {
			// TODO: May report the error through TpgMeta.
			fmt.Println("Error thrown: ", err)
			// Switch to Marking ABT Operations.
			e.ErrStack = []*TpgNode{node}
			node = nil
			for {
				if node == nil {
					// Pop ABT node.
					if len(e.ErrStack) == 0 {
						// Finish Marking ABT. Out with node == nil.
						break
					} else {
						node = e.ErrStack[len(e.ErrStack)-1]
						e.ErrStack = e.ErrStack[:len(e.ErrStack)-1]
						// Move to the head of the LD chain.
						node = e.Tpg.LDHeads[node.Opt.Txn()]
						// Transasctions that has been marked. Skip.
						if node.Status == ABT {
							node = nil
							continue
						}
					}
				} else {
					// Has been moved forward on the LD chain and is not nil
				}
				node.Status = ABT
				// Revert all the states version of this transaction.
				storage.Revert(node.Opt.Txn().Timestamp)
				// Push all the new PDs into the ErrStack.
				for _, d := range node.PD {
					if d.Status != ABT {
						e.ErrStack = append(e.ErrStack, d)
					}
				}
				node = node.LD
			}
		}
		if node == nil {
			// Has just finished ABT marking.
			continue
		}
		node.Opt.Logger()
		node.Status = EXE
		// Resolve the dependencies to be ready.
		for _, d := range node.PD {
			// Finish this dependency.
			d.DCount -= 1
			if d.DCount == 0 {
				if d.Status == BLK {
					// Push to stack.
					d.Status = OCCUPIED
					e.Stack = append(e.Stack, d)
				}
			} else {
				// Still BLK. There are other dependencies unfilled.
				// They would be disposed when dependencies are are done.
			}
		}
		// Move to the next along TD. The TD node could be [RDY, OCCUPIED, BLK, EXE].
		if node.TD == nil {
			node = nil
		} else if node.TD.Status == RDY {
			node = node.TD
		} else {
			// Ignore them. Other case they are going to be handled later or has been handled. Switch to nil.
			node = nil
		}
	}
}

func (e *DFSExecutor) StopExecutor() chan bool {
	return e.Ending
}
