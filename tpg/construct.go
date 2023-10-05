package tpg

import (
	events "gorphStream/events"
	"gorphStream/storage"

	"golang.org/x/exp/slices"
)

var (
	mode      = DFS
	threshold = 100
)

func Construct(txns []*events.Txn) *TpgMeta {
	skipLists := decompose(txns)
	// Construct into TPG meta.
	meta := TpgMeta{}
	meta.ExeMode = mode
	for idx, l := range skipLists {
		if len(l) != 0 && len(l[0].Opt.Dependencies()) == 0 {
			meta.Starter = append(meta.Starter, l[0])
			switch mode {
			case DFS:
				meta.Executor = append(meta.Executor, &DFSExecutor{
					Idx:    idx,
					Tpg:    &meta,
					Ending: make(chan bool),
				})
			case DFSNotify:
				meta.Executor = append(meta.Executor, &DFSNotifyExecutor{
					Idx:       idx,
					Tpg:       &meta,
					Ending:    make(chan bool),
					Threshold: threshold,
				})
			default:
				panic("not implemented")
			}
		}
	}
	meta.Finish = make(chan int)
	meta.Done = make([]bool, len(meta.Starter))

	return &meta
}

// Preprocess for operations. Sort them with proxy.
func decompose(txns []*events.Txn) [][]*TpgNode {
	// Split these Operations into State-partitioned arrays.
	slices.SortFunc(txns, func(a, b *events.Txn) int {
		if a.Timestamp < b.Timestamp {
			return -1
		} else {
			return 1
		}
	})

	skipLists := make([][]*TpgNode, storage.NumStatus())

	// To Construct the TD, we need to hold the last tpgNode in each status.
	lastNodes := make([]*TpgNode, storage.NumStatus())
	// To Construct the PD, we need to hold the last Write tpgNode in each status.
	lastWriteNodes := make([]*TpgNode, storage.NumStatus())

	/*
		Splitting txn operations.
		1. Construct LDs for operations in the same transaction.
		2. Leave proxy operations for those who written and going to be read later.
	*/
	for _, txn := range txns {
		// To Construct the LD, we need to hold the last tpgNode in each transaction.
		var lastInTxn *TpgNode = nil
		for _, op := range txn.Ops {
			op.SetTxn(txn)
			newNode := TpgNode{
				Opt:           op,
				TD:            nil,          // TD is assigned through "lastNodes" in the next round.
				PD:            []*TpgNode{}, // PD is assigned each time a new Write Operation is found.
				ParamVersions: make([]int64, len(op.Parameters())),
				LD:            lastInTxn, // LD links to the last operations in the transaction.

				DCount: 0, // Runtime parameters remains zero.
				Status: RDY,
				Notify: nil,
			}
			// If it's write operation, add PD from dependencies.
			if op.Type() == events.WRITE {
				newNode.DCount = len(op.Dependencies())
				if newNode.DCount > 0 {
					newNode.Status = BLK
				} else {
					newNode.Status = RDY
				}
				dp := op.Parameters()
				for _, d := range dp {
					if d == newNode.Opt.VarIdx() {
						if lastNodes[d] != nil {
							for idx, v := range dp {
								if d == v {
									newNode.ParamVersions[idx] = lastNodes[d].Opt.Txn().Timestamp
									break
								}
							}
						} else {
							// If no last node, it shall use the commited version. The time stamp for this is 0. No need to set.
						}
					} else {
						// The corresponding writting nodes links to new.
						if lastWriteNodes[d] != nil {
							lastWriteNodes[d].PD = append(lastWriteNodes[d].PD, &newNode)
							for idx, v := range dp {
								if d == v {
									newNode.ParamVersions[idx] = lastWriteNodes[d].Opt.Txn().Timestamp
									break
								}
							}
						} else {
							// If no one ever written, skip this.
						}
					}
				}
				// Update the last write.
				lastWriteNodes[op.VarIdx()] = &newNode
			} else {
				newNode.Status = RDY
			}
			if lastNodes[op.VarIdx()] != nil {
				lastNodes[op.VarIdx()].TD = &newNode
			}
			lastNodes[op.VarIdx()] = &newNode // Linked list move forward.
			lastInTxn = &newNode
			// Append the new node to the overall list.
			skipLists[op.VarIdx()] = append(skipLists[op.VarIdx()], &newNode)
		}
	}

	return skipLists
}
