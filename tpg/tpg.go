package tpg

import "gorphStream/events"

const (
	BLK = iota + 1
	RDY
	EXE
	ABT
	OCCUPIED // Another version of RDY. Used in runtime.
)

type TpgMeta struct {
	ExeMode int
	// Travers auxiliary.
	Starter  []*TpgNode
	Executor []Traversor
	// Ending Sync
	Finish chan int // Threads report finish.
	Done   []bool
	// Transactions
	Txns    []*events.Txn
	LDHeads map[*events.Txn]*TpgNode
}

type TpgNode struct {
	// Operations registered for this node.
	Opt events.Operation
	// Time next nodes.
	TD *TpgNode
	// Parameter next nodes.
	PD []*TpgNode
	// Required version for each PD.
	ParamVersions []int64
	// Transational linked nodes.
	LD *TpgNode

	/*
		Run time parameters.
	*/
	// Status.
	Status int
	// Count for dependencies
	DCount int
	// Notifier that waits for this node to be ready.
	Notify *chan *TpgNode
}
