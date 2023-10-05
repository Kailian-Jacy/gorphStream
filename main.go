package main

import (
	"gorphStream/cmd"
	"gorphStream/events"
	"gorphStream/storage"
	"gorphStream/tpg"
)

func main() {
	BankerApp()
}

func BankerApp() {
	storage.Init(cmd.BANKER_SCHEMA)
	tpg.Construct(
		// Shuffle input.
		[]*events.Txn{
			&cmd.TransferA2BTxn,
			&cmd.TransferB2ATxn,
			&cmd.DepositTxn,
		},
	).Handle()
	storage.Dump()
}
