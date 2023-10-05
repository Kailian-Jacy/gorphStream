package events

const (
	READ = iota + 1
	WRITE
	COMPAREANDSET
)

type Txn struct {
	Ops       []Operation
	Timestamp int64
}
