package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/galyym/fcm_push/internal/model"
	"github.com/galyym/fcm_push/internal/repository"
	"github.com/galyym/fcm_push/pkg/fcm"
)

type Config struct {
	WorkerCount    int
	PollInterval   time.Duration
	RetryIntervals []time.Duration
	CleanupAfter   time.Duration
}

type QueueWorker struct {
	repo          *repository.QueueRepository
	fcmClient     *fcm.Client
	config        Config
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	cleanupTicker *time.Ticker
}

func NewQueueWorker(repo *repository.QueueRepository, fcmClient *fcm.Client, config Config) *QueueWorker {
	ctx, cancel := context.WithCancel(context.Background())

	if len(config.RetryIntervals) == 0 {
		config.RetryIntervals = []time.Duration{
			1 * time.Minute,
			5 * time.Minute,
			15 * time.Minute,
		}
	}

	if config.CleanupAfter == 0 {
		config.CleanupAfter = 30 * 24 * time.Hour
	}

	return &QueueWorker{
		repo:      repo,
		fcmClient: fcmClient,
		config:    config,
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (w *QueueWorker) Start() {
	log.Printf("Starting queue worker with %d workers, poll interval: %s",
		w.config.WorkerCount, w.config.PollInterval)

	for i := 0; i < w.config.WorkerCount; i++ {
		w.wg.Add(1)
		go w.workerLoop(i)
	}

	w.wg.Add(1)
	go w.cleanupLoop()

	log.Println("Queue worker started successfully")
}
func (w *QueueWorker) Stop() {
	log.Println("Stopping queue worker...")
	w.cancel()

	if w.cleanupTicker != nil {
		w.cleanupTicker.Stop()
	}

	w.wg.Wait()
	log.Println("Queue worker stopped")
}

func (w *QueueWorker) workerLoop(workerID int) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.PollInterval)
	defer ticker.Stop()

	log.Printf("Worker %d started", workerID)

	for {
		select {
		case <-w.ctx.Done():
			log.Printf("Worker %d stopping", workerID)
			return
		case <-ticker.C:
			w.processBatch(workerID)
		}
	}
}

func (w *QueueWorker) processBatch(workerID int) {
	ctx, cancel := context.WithTimeout(w.ctx, 30*time.Second)
	defer cancel()
	tasks, err := w.repo.GetPendingTasks(ctx, 10)
	if err != nil {
		log.Printf("Worker %d: failed to get pending tasks: %v", workerID, err)
		return
	}

	if len(tasks) == 0 {
		return
	}

	log.Printf("Worker %d: processing %d tasks", workerID, len(tasks))

	for _, task := range tasks {
		w.processTask(ctx, workerID, task)
	}
}

func (w *QueueWorker) processTask(ctx context.Context, workerID int, task *model.PushQueueTask) {
	log.Printf("Worker %d: processing task %s (attempt %d/%d)",
		workerID, task.ID, task.Attempts+1, task.MaxAttempts)
	messageID, err := w.fcmClient.SendNotification(
		ctx,
		task.Token,
		task.Title,
		task.Body,
		task.Data,
		task.Priority,
	)

	if err != nil {
		w.handleTaskFailure(ctx, workerID, task, err)
		return
	}

	if err := w.repo.UpdateTaskSuccess(ctx, task.ID, messageID); err != nil {
		log.Printf("Worker %d: failed to update task success: %v", workerID, err)
		return
	}

	log.Printf("Worker %d: task %s completed successfully, FCM message ID: %s",
		workerID, task.ID, messageID)
}

func (w *QueueWorker) handleTaskFailure(ctx context.Context, workerID int, task *model.PushQueueTask, err error) {
	log.Printf("Worker %d: task %s failed: %v", workerID, task.ID, err)

	nextAttempt := task.Attempts + 1

	if nextAttempt < task.MaxAttempts {
		nextRetry := w.calculateNextRetry(task.Attempts)

		log.Printf("Worker %d: scheduling retry for task %s at %s (attempt %d/%d)",
			workerID, task.ID, nextRetry.Format(time.RFC3339), nextAttempt+1, task.MaxAttempts)

		if err := w.repo.UpdateTaskFailure(ctx, task.ID, err.Error(), &nextRetry); err != nil {
			log.Printf("Worker %d: failed to schedule retry: %v", workerID, err)
		}
	} else {
		log.Printf("Worker %d: task %s permanently failed after %d attempts",
			workerID, task.ID, task.MaxAttempts)

		if err := w.repo.UpdateTaskFailure(ctx, task.ID, err.Error(), nil); err != nil {
			log.Printf("Worker %d: failed to mark task as failed: %v", workerID, err)
		}
	}
}

func (w *QueueWorker) calculateNextRetry(currentAttempt int) time.Time {
	var delay time.Duration

	if currentAttempt < len(w.config.RetryIntervals) {
		delay = w.config.RetryIntervals[currentAttempt]
	} else {
		delay = w.config.RetryIntervals[len(w.config.RetryIntervals)-1]
	}

	return time.Now().Add(delay)
}

func (w *QueueWorker) cleanupLoop() {
	defer w.wg.Done()

	w.cleanupTicker = time.NewTicker(24 * time.Hour)
	defer w.cleanupTicker.Stop()

	w.runCleanup()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-w.cleanupTicker.C:
			w.runCleanup()
		}
	}
}

func (w *QueueWorker) runCleanup() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Printf("Running cleanup for tasks older than %s", w.config.CleanupAfter)

	deleted, err := w.repo.CleanupOldTasks(ctx, w.config.CleanupAfter)
	if err != nil {
		log.Printf("Cleanup failed: %v", err)
		return
	}

	if deleted > 0 {
		log.Printf("Cleanup completed: deleted %d old tasks", deleted)
	}
}
