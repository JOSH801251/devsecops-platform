package jobs

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestPostgresRepoIntegration(t *testing.T) {
	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_TEST_DSN not set; skipping PostgreSQL integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	repo, err := NewPostgresRepo(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to initialize PostgreSQL repository: %v", err)
	}

	var insertedIDs []string

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(), 5*time.Second,
		)
		defer cleanupCancel()

		for _, id := range insertedIDs {
			_, _ = repo.Pool().Exec(
				cleanupCtx,
				`DELETE FROM jobs WHERE id = $1`,
				id,
			)
		}

		repo.Close()
	})

	if repo.Pool() == nil {
		t.Fatal("expected a valid PostgreSQL connection pool")
	}

	pendingBefore, err := repo.Count(ctx, StatusPending)
	if err != nil {
		t.Fatal(err)
	}

	doneBefore, err := repo.Count(ctx, StatusDone)
	if err != nil {
		t.Fatal(err)
	}

	failedBefore, err := repo.Count(ctx, StatusFailed)
	if err != nil {
		t.Fatal(err)
	}

	// Test inserting a pending job.
	pendingID, err := repo.Enqueue(ctx, []byte(`{"test":"pending"}`))
	if err != nil {
		t.Fatalf("failed to enqueue pending job: %v", err)
	}

	insertedIDs = append(insertedIDs, pendingID)

	if pendingID == "" {
		t.Fatal("expected a job ID")
	}

	pendingAfter, err := repo.Count(ctx, StatusPending)
	if err != nil {
		t.Fatal(err)
	}

	if pendingAfter != pendingBefore+1 {
		t.Fatalf("expected pending count %d, got %d",
			pendingBefore+1, pendingAfter)
	}

	// Move the job to done and verify the status count.
	_, err = repo.Pool().Exec(
		ctx,
		`UPDATE jobs SET status = $1 WHERE id = $2`,
		string(StatusDone),
		pendingID,
	)
	if err != nil {
		t.Fatalf("failed to update job status: %v", err)
	}

	doneAfter, err := repo.Count(ctx, StatusDone)
	if err != nil {
		t.Fatal(err)
	}

	if doneAfter != doneBefore+1 {
		t.Fatalf("expected done count %d, got %d",
			doneBefore+1, doneAfter)
	}

	// Test the failed status.
	failedID, err := repo.Enqueue(ctx, []byte(`{"test":"failed"}`))
	if err != nil {
		t.Fatal(err)
	}

	insertedIDs = append(insertedIDs, failedID)

	_, err = repo.Pool().Exec(
		ctx,
		`UPDATE jobs SET status = $1 WHERE id = $2`,
		string(StatusFailed),
		failedID,
	)
	if err != nil {
		t.Fatalf("failed to update job status: %v", err)
	}

	failedAfter, err := repo.Count(ctx, StatusFailed)
	if err != nil {
		t.Fatal(err)
	}

	if failedAfter != failedBefore+1 {
		t.Fatalf("expected failed count %d, got %d",
			failedBefore+1, failedAfter)
	}
}
