package orchestrator

import (
	"testing"

	"cloudperf/backend/internal/db"
	"cloudperf/backend/internal/model"
)

func TestValidatePairs(t *testing.T) {
	ok := []db.PairSpec{{SourceNodeID: "a", TargetNodeID: "b"}, {SourceNodeID: "a", TargetNodeID: "c"}}
	if err := ValidatePairs(ok); err != nil {
		t.Fatalf("expected valid pairs, got %v", err)
	}

	dup := []db.PairSpec{{SourceNodeID: "a", TargetNodeID: "b"}, {SourceNodeID: "a", TargetNodeID: "b"}}
	if err := ValidatePairs(dup); err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestAggregateStatus(t *testing.T) {
	if got := aggregateStatus([]model.RunPair{{Status: "success"}, {Status: "success"}}); got != "success" {
		t.Fatalf("want success, got %s", got)
	}
	if got := aggregateStatus([]model.RunPair{{Status: "success"}, {Status: "failed"}}); got != "partial_failed" {
		t.Fatalf("want partial_failed, got %s", got)
	}
	if got := aggregateStatus([]model.RunPair{{Status: "failed"}}); got != "failed" {
		t.Fatalf("want failed, got %s", got)
	}
}
