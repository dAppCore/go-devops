package dev

import core "dappco.re/go"

func ExampleForgePR() {
	pr := ForgePR{Number: 7, Title: "Update CI", ReviewDecision: "APPROVED"}
	core.Println(pr.Number, pr.ReviewDecision)
	// Output: 7 APPROVED
}
