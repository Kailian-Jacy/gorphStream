package utils

const (
	READ = iota + 1
	WRITE
	COMPAREANDSET
	PROXY
)

type Txn struct {
	Ops       []Operation
	Timestamp int64
}
