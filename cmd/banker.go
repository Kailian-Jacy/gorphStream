package cmd

import (
	"errors"
	"gorphStream/events"
	"gorphStream/storage"
)

/*
	This is an example app with transactions. Build with naive api.
	- First define the Operations and provide corresponding callbacks.
	- Then assemble them to transactions and Commit them to gorphStream.
*/

// Define storage schema.
const (
	A = iota
	B
)

const BANKER_SCHEMA = 2

// The part of Callbacks providing `Operation.Do()`

var V1, V2 = 500, 200

// Just write some deposit.
var deposit = func(target storage.ParamView, _ storage.ParamView) error {
	target.Set(V1)
	return nil
}

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
var transferSend = func(target storage.ParamView, params storage.ParamView) error {
	deposit := params.Get(0)
	if deposit >= V2 {
		target.Set(deposit - V2)
		return nil
	} else {
		return errors.New("insufficient valance")
	}
}

// Assemble into transactions.
var DepositTxn = events.Txn{
	Ops: []events.Operation{
		&events.W{
			Name:   "depositA",
			Params: nil,
			Target: A,
			Do:     deposit,
		}},
	Timestamp: int64(111),
}

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

var TransferB2ATxn = events.Txn{
	Ops: []events.Operation{
		// B Add.
		&events.W{
			Name:   "A receive from B",
			Params: []int{B, A},
			Target: A,
			Do:     transferReceive,
		},
		// A decrease.
		&events.W{
			Name:   "B send to A",
			Params: []int{B},
			Target: B,
			Do:     transferSend,
		},
	},
	Timestamp: int64(333),
}

// Send 3 transactions to build the TPG. Shuffled.
