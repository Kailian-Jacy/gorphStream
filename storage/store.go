// A pure lockless storage.

package storage

import "fmt"

type StorageMeta struct {
	// Total number of Storage Meta.
	numStatus     int
	commited      []int
	history       []map[int64]int
	latestVersion []int64
}

var s = StorageMeta{}

// The operation here will always be successful. Cuz the state schema is always pre-defined.
func (s *StorageMeta) Read(version int64, idx int) int {
	return s.history[idx][version]
}

func (s *StorageMeta) Write(version int64, idx int, v int) {
	s.history[idx][version] = v
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
	s.latestVersion = make([]int64, schema)
	s.numStatus = schema
	s.history = make([]map[int64]int, 0)
	for _, v := range s.commited {
		s.history = append(s.history, map[int64]int{int64(0): v})
	}
	return nil
}

func Commit() {
	for idx, v := range s.latestVersion {
		s.commited[idx] = s.history[idx][v]
	}
	// Clear the history.
	s.history = make([]map[int64]int, 0)
	for _, v := range s.commited {
		s.history = append(s.history, map[int64]int{int64(0): v})
	}
}

func LatestVersion(idx int) int64 {
	return s.latestVersion[idx]
}
