package main

import (
	"morphGo/cmd"
	"morphGo/storage"
	"morphGo/tpg"
	"morphGo/utils"
)

func main() {
	exampleWithNaiveApi()
}

func exampleWithNaiveApi() {
	storage.Init(cmd.BANKER_SCHEMA)
	tpg.Construct(
		// Shuffle input.
		[]*utils.Txn{
			&cmd.TransferA2BTxn,
			&cmd.TransferB2ATxn,
			&cmd.DepositTxn,
		},
	).Handle()
	storage.Dump()
}
