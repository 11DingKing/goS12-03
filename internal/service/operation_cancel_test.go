package service

import (
	"testing"
	"time"

	"batteryops/internal/domain"
)

// TestCancelledOperationIsNotExecutable describes the expected public behaviour
// of the grid-switching workflow after a cancellation.
//
// Input: the duty dispatcher initiates a grid-connect operation, the supply
// station confirms it, and then the dispatcher cancels it before it is carried
// out.  A black-start operation is exercised the same way.
//
// Expected output: the cancelled operation stays cancelled.  The execute call
// must be rejected, no execution timestamp is recorded, and the audit trail must
// not contain an execution entry.
func TestCancelledOperationIsNotExecutable(t *testing.T) {
	for _, opType := range []domain.OperationType{domain.OperationGridConnect, domain.OperationBlackStart} {
		t.Run(string(opType), func(t *testing.T) {
			svc, _, _ := newTestService(30 * time.Minute)

			op, err := svc.InitiateOperation(opType, "dispatcher-1")
			if err != nil {
				t.Fatalf("initiate: %v", err)
			}
			if op, err = svc.ConfirmOperation(op.ID, "station-1"); err != nil {
				t.Fatalf("station confirm: %v", err)
			}
			if op.Status != domain.OperationConfirmed {
				t.Fatalf("expected status %s, got %s", domain.OperationConfirmed, op.Status)
			}

			op, err = svc.CancelOperation(op.ID, "dispatcher-1")
			if err != nil {
				t.Fatalf("cancel: %v", err)
			}
			if op.Status != domain.OperationCancelled {
				t.Fatalf("expected status %s, got %s", domain.OperationCancelled, op.Status)
			}
			executed, err := svc.ExecuteOperation(op.ID)
			if err == nil {
				t.Fatalf("expected the execute call on cancelled operation %s to be rejected, got status %s",
					op.ID, executed.Status)
			}

			stored := findOperation(t, svc, op.ID)
			if stored.Status != domain.OperationCancelled {
				t.Fatalf("expected the operation to stay %s, got %s", domain.OperationCancelled, stored.Status)
			}
			if stored.ExecutedAt != nil {
				t.Fatalf("expected no execution timestamp on a cancelled operation, got %v", stored.ExecutedAt)
			}
			for _, entry := range stored.AuditTrail {
				if entry.Action == "operation_executed" {
					t.Fatalf("audit trail of a cancelled operation must not contain operation_executed: %+v",
						stored.AuditTrail)
				}
			}
		})
	}
}

// TestDualConfirmedOperationStillExecutes keeps the ordinary path covered: an
// operation confirmed by both the dispatcher and the supply station and never
// cancelled must still execute exactly once.
func TestDualConfirmedOperationStillExecutes(t *testing.T) {
	svc, _, _ := newTestService(30 * time.Minute)

	op, err := svc.InitiateOperation(domain.OperationGridConnect, "dispatcher-1")
	if err != nil {
		t.Fatalf("initiate: %v", err)
	}
	if _, err := svc.ConfirmOperation(op.ID, "station-1"); err != nil {
		t.Fatalf("confirm: %v", err)
	}

	executed, err := svc.ExecuteOperation(op.ID)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if executed.Status != domain.OperationExecuted {
		t.Fatalf("expected status %s, got %s", domain.OperationExecuted, executed.Status)
	}
	if executed.ExecutedAt == nil {
		t.Fatalf("expected an execution timestamp")
	}
	if _, err := svc.ExecuteOperation(op.ID); err == nil {
		t.Fatalf("expected the second execute call to be rejected")
	}
	if _, err := svc.CancelOperation(op.ID, "dispatcher-1"); err == nil {
		t.Fatalf("expected cancelling an executed operation to be rejected")
	}
}

func findOperation(t *testing.T, svc *Service, id string) *domain.Operation {
	t.Helper()
	for _, op := range svc.ListOperations() {
		if op.ID == id {
			return op
		}
	}
	t.Fatalf("operation %s not found", id)
	return nil
}
