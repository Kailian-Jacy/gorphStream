package storage

/*
ParamView is a visiting controll that provide multi-version visiting for operations.
- pidx is the index of PD array in Operations.PD.
- translate to state index and version to visit.
*/
type ParamView struct {
	Versions *[]int64
	Params   *[]int
	ReadOnly bool
}

// Get provide index access to dependent parameters.
func (v *ParamView) Get(pidx int) int {
	return s.Read((*v.Versions)[pidx], (*v.Params)[pidx])
}

// Set only provide write to target.
func (v *ParamView) Set(value int) {
	if v.ReadOnly {
		panic("Not allowed to write.")
	}
	s.Write((*v.Versions)[0], (*v.Params)[0], value)
}
