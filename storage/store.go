// A pure lockless storage.

package storage

import (
	"fmt"

	"github.com/huandu/skiplist"
)

type StorageMeta struct {
	// Total number of Storage Meta.
	numStatus int
	commited  []int
	// History shall be organized as skip list.
	history       []*skiplist.SkipList
	latestVersion []int64
}

var s = StorageMeta{}

// The operation here will always be successful. Cuz the state schema is always pre-defined.
func (s *StorageMeta) Read(version int64, idx int) int {
	return s.history[idx].Get(version).Value.(int)
}

func (s *StorageMeta) Write(version int64, idx int, v int) {
	s.history[idx].Set(version, v)
	if version > s.latestVersion[idx] {
		s.latestVersion[idx] = version
	}
}

func NumStatus() int { return s.numStatus }

func Dump() {
	Commit()
	for idx, v := range s.commited {
		fmt.Printf("Storage %d: %d\n", idx, v)
	}
}

func Init(schema int) error {
	s.commited = make([]int, schema)
	s.numStatus = schema
	s.latestVersion = make([]int64, schema)
	for idx, v := range s.commited {
		s.history = append(s.history, skiplist.New(skiplist.Int64))
		s.history[idx].Set(int64(0), v)
	}
	return nil
}

func Commit() {
	for idx, lv := range s.latestVersion {
		s.commited[idx] = s.history[idx].Get(lv).Value.(int)
	}
	// Clear the history.
	for idx, v := range s.commited {
		s.history = append(s.history, skiplist.New(skiplist.Int64))
		s.history[idx].Set(int64(0), v)
	}
}

func Revert(txnTS int64) {
	// No need to delete in history. Even no need to mark. Just move the lastTag away and
	// They would be discarded in the next commit.
	for idx, lsTS := range s.latestVersion {
		if txnTS == lsTS {
			s.latestVersion[idx] = s.history[idx].Get(s.latestVersion[idx]).Prev().Key().(int64)
		}
	}
}

func LatestVersion(idx int) int64 {
	return s.latestVersion[idx]
}
