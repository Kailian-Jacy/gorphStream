package main

import (
	"morphGo/storage"
	"morphGo/tpg"
	"morphGo/utils"
)

const (
	A = iota
	B
)

func main() {

	V1, V2 := 500, 200

	// Just write some deposit.
	deposit := func(target storage.ParamView, _ storage.ParamView) {
		target.Set(V1)
	}
	// Check the Balance in the other account and write to this one.
	transferReceive := func(target storage.ParamView, params storage.ParamView) {
		if params.Get(0) >= V2 {
			total := params.Get(1) + V2
			target.Set(total)
		}
	}
	transferSend := func(target storage.ParamView, params storage.ParamView) {
		deposit := params.Get(0)
		if deposit >= V2 {
			target.Set(deposit - V2)
		}
	}

	depositTxn := utils.Txn{
		Ops: []utils.Operation{
			&utils.W{
				Name:   "depositA",
				Params: nil,
				Target: A,
				Do:     deposit,
			}},
		Timestamp: int64(111),
	}

	transferA2B := utils.Txn{
		Ops: []utils.Operation{
			// B Add.
			&utils.W{
				Name:   "B receive from A",
				Params: []int{A, B},
				Target: B,
				Do:     transferReceive,
			},
			// A decrease.
			&utils.W{
				Name:   "A send to B",
				Params: []int{A},
				Target: A,
				Do:     transferSend,
			},
		},
		Timestamp: int64(222),
	}

	transferB2A := utils.Txn{
		Ops: []utils.Operation{
			// B Add.
			&utils.W{
				Name:   "A receive from B",
				Params: []int{B, A},
				Target: A,
				Do:     transferReceive,
			},
			// A decrease.
			&utils.W{
				Name:   "B send to A",
				Params: []int{B},
				Target: B,
				Do:     transferSend,
			},
		},
		Timestamp: int64(333),
	}

	// Send 3 transactions to build the TPG. Shuffled.

	storage.InitStorage(2)
	g := tpg.Construct(
		[]*utils.Txn{
			&transferA2B,
			&transferB2A,
			&depositTxn,
		},
	)

	g.Handle()

	storage.Dump()
}
