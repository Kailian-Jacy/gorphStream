package utils

import (
	"fmt"
	"morphGo/storage"
)

/*
	Operations are used to dispose some relationship. Dependency, etc.
*/

type Operation interface {
	Type() int
	// What must be fullfiled before operation.
	Dependencies() []int
	// Compared to Dependencies, parameters includes the variable himself.
	Parameters() []int
	// Mainly disposed var
	VarIdx() int
	// Txn
	Txn() *Txn
	SetTxn(*Txn)

	// Execute. The parameters are in the calling struct.
	Execute(*[]int64)

	// Used for debugging.
	Logger()
}

/*
	The basic model for writing is W(k, f(k1, k2, k3, .... kn))
*/

// W(Target, f(Params1, Params2, ...))
type W struct {
	transaction  *Txn
	Name         string
	Params       []int
	dependencies []int
	Target       int
	Do           func(storage.ParamView, storage.ParamView)
}

func (w *W) Execute(versions *[]int64) {
	targetParamView, paramView := storage.ParamView{
		Versions: &[]int64{w.transaction.Timestamp},
		Params:   &[]int{w.Target},
		ReadOnly: false,
	}, storage.ParamView{
		Versions: versions,
		Params:   &w.Params,
		ReadOnly: true,
	}
	w.Do(targetParamView, paramView)
	// New Record of target shall not rely on User set. If no new record, we need to set it to maintain correspondent.
	if storage.LatestVersion(w.Target) != w.transaction.Timestamp {
		// targetParamView.Set(LatestValue)
		panic("TODO")
	}
}

func (w *W) Type() int { return WRITE }

func (w *W) VarIdx() int { return w.Target }

func (w *W) Txn() *Txn { return w.transaction }

func (w *W) SetTxn(t *Txn) { w.transaction = t }

func (w *W) Logger() {
	fmt.Println("Executed: ", w.Name)
}

func (w *W) Parameters() []int { return w.Params }

// Dependencies is the other refered parameters. Excluding himself.
func (w *W) Dependencies() []int {
	if w.dependencies == nil {
		w.dependencies = []int{}
		for _, v := range w.Params {
			if v != w.Target {
				w.dependencies = append(w.dependencies, v)
			}
		}
	}
	return w.dependencies
}

// Proxy is used to link up Operation and its visited variable.
// type WProxy struct{}

// func (p WProxy) Execute() {}

// func (p WProxy) Type() int { return PROXY }

// func (p WProxy) VarIdx() int { return -1 }

// func (p WProxy) Dependencies() []int { return nil }
