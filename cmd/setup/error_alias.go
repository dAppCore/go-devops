package setup

import core "dappco.re/go"

type coreFailure = error

func commandResultError(r core.Result) coreFailure {
	if r.OK {
		return nil
	}
	if err, ok := r.Value.(error); ok {
		return err
	}
	return core.Errorf("%v", r.Value)
}
