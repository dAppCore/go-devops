package dev

import core "dappco.re/go"

func ExampleWorkflowRun() {
	run := WorkflowRun{Name: "tests", Conclusion: "success"}
	core.Println(run.Name, run.Conclusion)
	// Output: tests success
}
