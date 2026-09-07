package service

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Euclides-EM/commentaria-hub/ocrflow/internal/model/annotation"
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

func TestFailAnnotationRuleCallback(t *testing.T) {
	jobsStore := store.NewJobStore(cache.NewCache())
	svc := &Job{jobsStore: jobsStore}
	details, err := json.Marshal(annotationRuleDispatchResult{
		AnnotationID:      "ann_target",
		DatasetID:         "ds_target",
		State:             "gpu_farm_submitted_waiting_for_callback",
		Message:           "waiting",
		FollowLogs:        "ssh cca-ocr tail remote.log",
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

	if err := svc.FailAnnotationRuleCallback("ds_target", "ann_target", annotation.DetectionModeLines, "Document is empty"); err != nil {
		t.Fatal(err)
	}
	if jb.Status != job.StatusFailed || jb.FinishedAt == nil {
		t.Fatalf("job not failed: status=%s finished_at=%v", jb.Status, jb.FinishedAt)
	}
	var failed annotationRuleDispatchResult
	if err := json.Unmarshal([]byte(jb.Details), &failed); err != nil {
		t.Fatal(err)
	}
	if failed.State != "gpu_farm_failed" {
		t.Fatalf("unexpected failure state %q", failed.State)
	}
	if !strings.Contains(failed.Message, "lines detection failed: Document is empty") {
		t.Fatalf("unexpected failure message %q", failed.Message)
	}
	if failed.FollowLogs == "" {
		t.Fatal("failure removed the remote log command")
	}

	if err := svc.CompleteAnnotationRuleCallback("ds_target", "ann_target"); err != nil {
		t.Fatal(err)
	}
	if jb.Status != job.StatusFailed {
		t.Fatalf("late success callback changed failed job to %s", jb.Status)
	}
}

func TestFailAnnotationRuleCallbackAcceptsEarlyCallbackDuringProgressUpdate(t *testing.T) {
	jobsStore := store.NewJobStore(cache.NewCache())
	svc := &Job{jobsStore: jobsStore}
	jb := &job.Job{
		Task:   job.AnnotationRuleApply,
		Target: &job.Target{DatasetID: "ds_target", AnnotationID: "ann_target"},
		Status: job.StatusRunning,
		Details: "GPU farm job submitted; waiting for detection result callback: " +
			"follow logs with: ssh cca-ocr tail remote.log",
	}
	jb.ID = "job_test"
	jobsStore.Create(jb)

	if err := svc.FailAnnotationRuleCallback("ds_target", "ann_target", annotation.DetectionModeLines, "startup failed"); err != nil {
		t.Fatal(err)
	}
	if jb.Status != job.StatusFailed {
		t.Fatalf("early callback left job in status %s", jb.Status)
	}
	var failed annotationRuleDispatchResult
	if err := json.Unmarshal([]byte(jb.Details), &failed); err != nil {
		t.Fatal(err)
	}
	if failed.FollowLogs == "" {
		t.Fatal("early failure callback did not preserve progress details")
	}
}
