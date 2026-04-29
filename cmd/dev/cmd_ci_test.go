package dev

import (
	core "dappco.re/go"
	"time"
)

func TestCmdCi_WorkflowRun_Good(t *core.T) {
	run := WorkflowRun{Name: "tests", Status: "completed", Conclusion: "success", CreatedAt: time.Unix(1, 0)}

	core.AssertEqual(t, "tests", run.Name)
	core.AssertEqual(t, "success", run.Conclusion)
}

func TestCmdCi_WorkflowRun_Bad(t *core.T) {
	run := WorkflowRun{Status: "completed", Conclusion: "failure"}

	core.AssertEqual(t, "failure", run.Conclusion)
	core.AssertEqual(t, "", run.Name)
}

func TestCmdCi_WorkflowRun_Ugly(t *core.T) {
	run := WorkflowRun{}

	core.AssertEqual(t, "", run.Status)
	core.AssertTrue(t, run.CreatedAt.IsZero())
}
