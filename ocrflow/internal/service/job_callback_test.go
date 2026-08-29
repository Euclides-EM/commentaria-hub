package service

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/job"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/store"
	"github.com/Euclides-EM/commentaria-hub/ocrflow/pkg/cache"
)

func TestCompleteAnnotationRuleCallback(t *testing.T) {
	jobsStore := store.NewJobStore(cache.NewCache())
	svc := &Job{jobsStore: jobsStore}
	details, err := json.Marshal(annotationRuleDispatchResult{
		AnnotationID:      "ann_target",
		DatasetID:         "ds_target",
		State:             "gpu_farm_submitted_waiting_for_callback",
		Message:           "waiting",
		FollowLogs:        "ssh 'cca-ocr' 'tail -n 100 -F first'\nssh 'cca-ocr' 'tail -n 100 -F second'",
		CallbacksExpected: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	jb := &job.Job{
		Task:    job.AnnotationRuleApply,
		Target:  &job.Target{DatasetID: "ds_target", AnnotationID: "ann_target"},
		Status:  job.StatusRunning,
		Details: string(details),
	}
	jb.ID = "job_test"
	jobsStore.Create(jb)

	if err := svc.CompleteAnnotationRuleCallback("ds_target", "ann_target"); err != nil {
		t.Fatal(err)
	}
	if jb.Status != job.StatusRunning || jb.FinishedAt != nil {
		t.Fatalf("job completed before all callbacks: status=%s finished_at=%v", jb.Status, jb.FinishedAt)
	}

	if err := svc.CompleteAnnotationRuleCallback("ds_target", "ann_target"); err != nil {
		t.Fatal(err)
	}
	if jb.Status != job.StatusCompleted || jb.FinishedAt == nil {
		t.Fatalf("job not completed after final callback: status=%s finished_at=%v", jb.Status, jb.FinishedAt)
	}
	var completed annotationRuleDispatchResult
	if err := json.Unmarshal([]byte(jb.Details), &completed); err != nil {
		t.Fatal(err)
	}
	if completed.State != "callback_received" {
		t.Fatalf("unexpected callback state %q", completed.State)
	}
	if completed.FollowLogs != "" {
		t.Fatalf("follow logs retained after callback: %v", completed.FollowLogs)
	}
	if strings.Contains(jb.Details, "follow_logs") {
		t.Fatalf("follow_logs field was not removed: %s", jb.Details)
	}
	if completed.CallbacksReceived != 2 {
		t.Fatalf("received callbacks = %d, want 2", completed.CallbacksReceived)
	}
}
