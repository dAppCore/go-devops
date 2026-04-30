package dev

import core "dappco.re/go"

func TestCmdReviews_ForgePR_Good(t *core.T) {
	pr := ForgePR{Number: 7, Title: "Update CI", ReviewDecision: "APPROVED"}

	core.AssertEqual(t, int64(7), pr.Number)
	core.AssertEqual(t, "APPROVED", pr.ReviewDecision)
}

func TestCmdReviews_ForgePR_Bad(t *core.T) {
	pr := ForgePR{Draft: true}

	core.AssertTrue(t, pr.Draft)
	core.AssertEqual(t, "", pr.Title)
}

func TestCmdReviews_ForgePR_Ugly(t *core.T) {
	pr := ForgePR{}

	core.AssertEqual(t, int64(0), pr.Number)
	core.AssertTrue(t, pr.CreatedAt.IsZero())
}
