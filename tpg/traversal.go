package tpg

const (
	DFS = iota + 1
	DFSNotify
)

type Traversor interface {
	Traversal(*TpgNode)
	StopExecutor() chan bool
}

type DFSExecutor struct {
	Idx   int
	Tpg   *TpgMeta
	Stack []*TpgNode

	// Receive ending from the main thread.
	Ending chan bool
}

type DFSNotifyExecutor struct {
	Idx   int
	Tpg   *TpgMeta // Visit shared variables.
	Stack []*TpgNode

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
		node.Opt.Execute(&node.ParamVersions)
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

/*
A single thread execute to traverse through. Compared to Simply DFS traversal,
this handler shared Nodes through threads with channel to balance the load.

This traversal features:
 1. Each node would be traversed only once. BLK -> RDY -> Occupied -> EXE.
 2. Based on the current load, the executor would decide if to give the notifier. In low load (short stack length), it would
    give notifier channel in BLK nodes to get update notification.
*/
func (e *DFSNotifyExecutor) Traversal(node *TpgNode) {
	for {
		if node == nil {
			// Select from the waiting list Stack.
			// The node inside is guaranteed to be RDY(Occupied) but not BLK or EXE.
			if len(e.Stack) == 0 {
				// Stack emptied for this thread.
				e.Tpg.Finish <- e.Idx
				select {
				case node = <-e.Notifier:
					// Resend to cancel done report.
					e.Tpg.Finish <- e.Idx
				case <-e.StopExecutor():
					// Received signal from main suggesting all threads done.
					break
				}
			} else {
				node = e.Stack[len(e.Stack)-1]
				e.Stack = e.Stack[:len(e.Stack)-1] // Pop the last one.
			}
		}
		// Execute Current.
		node.Status = OCCUPIED
		node.Opt.Execute(&node.ParamVersions)
		node.Status = EXE
		// Resolve the dependencies to be ready.
		for _, d := range node.PD {
			// Finish this dependency.
			d.DCount -= 1
			if d.DCount == 0 {
				if d.Status == BLK {
					// Free now. Try to send to the registered free thread.
					if d.Notify != nil {
						select {
						case (*d.Notify) <- d:
						default:
							// Pass. Use stack.
						}
					}
					// The otherside is not waiting or no notifier registered. Push to stack.
					d.Status = OCCUPIED
					e.Stack = append(e.Stack, d)
				}
			} else {
				// Still BLK. There are other dependencies unfilled.
				if d.Notify == nil && len(e.Stack) < e.Threshold {
					d.Notify = &e.Notifier
				}
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

func (e *DFSNotifyExecutor) StopExecutor() chan bool {
	return e.Ending
}
