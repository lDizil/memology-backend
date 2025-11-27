package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"memology-backend/internal/config"
	"memology-backend/internal/repository"

	"github.com/google/uuid"
)

type TaskProcessor struct {
	memeRepo      repository.MemeRepository
	aiSvc         AIService
	minioSvc      MinIOService
	taskQueue     chan uuid.UUID
	workers       int
	pollInterval  time.Duration
	wg            sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.Mutex
	processingIDs map[uuid.UUID]bool
}

func NewTaskProcessor(
	cfg *config.Config,
	memeRepo repository.MemeRepository,
	aiSvc AIService,
	minioSvc MinIOService,
) *TaskProcessor {
	ctx, cancel := context.WithCancel(context.Background())

	return &TaskProcessor{
		memeRepo:      memeRepo,
		aiSvc:         aiSvc,
		minioSvc:      minioSvc,
		taskQueue:     make(chan uuid.UUID, cfg.TaskProcessor.QueueSize),
		workers:       cfg.TaskProcessor.Workers,
		pollInterval:  cfg.TaskProcessor.PollInterval,
		ctx:           ctx,
		cancel:        cancel,
		processingIDs: make(map[uuid.UUID]bool),
	}
}

func (tp *TaskProcessor) Start() {
	log.Printf("Starting Task Processor with %d workers, poll interval: %v", tp.workers, tp.pollInterval)

	for i := 0; i < tp.workers; i++ {
		tp.wg.Add(1)
		go tp.worker(i + 1)
	}

	tp.wg.Add(1)
	go tp.stuckTasksScanner()

	log.Println("Task Processor started successfully")
}

func (tp *TaskProcessor) Stop() {
	log.Println("Stopping Task Processor...")
	tp.cancel()
	close(tp.taskQueue)
	tp.wg.Wait()
	log.Println("Task Processor stopped")
}

func (tp *TaskProcessor) stuckTasksScanner() {
	defer tp.wg.Done()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	log.Println("Stuck tasks scanner started (checking every 1 hour)")

	tp.scanAndReschedule()

	for {
		select {
		case <-tp.ctx.Done():
			log.Println("Stuck tasks scanner stopped")
			return
		case <-ticker.C:
			tp.scanAndReschedule()
		}
	}
}

func (tp *TaskProcessor) scanAndReschedule() {
	log.Println("=== Scanning for stuck tasks ===")

	stuckMemes, err := tp.memeRepo.FindStuckMemes(tp.ctx, 30*time.Minute)
	if err != nil {
		log.Printf("Failed to find stuck memes: %v", err)
		return
	}

	log.Printf("Found %d stuck memes to reschedule", len(stuckMemes))

	for _, meme := range stuckMemes {
		tp.mu.Lock()
		isProcessing := tp.processingIDs[meme.ID]
		tp.mu.Unlock()

		if !isProcessing {
			log.Printf("Rescheduling stuck meme %s (status: %s, updated: %s)",
				meme.ID, meme.Status, meme.UpdatedAt.Format("2006-01-02 15:04:05"))

			meme.Status = "pending"
			if err := tp.memeRepo.Update(tp.ctx, meme); err != nil {
				log.Printf("Failed to reset status for meme %s: %v", meme.ID, err)
				continue
			}

			if err := tp.AddTask(meme.ID); err != nil {
				log.Printf("Failed to reschedule meme %s: %v", meme.ID, err)
			}
		} else {
			log.Printf("Meme %s is already in processing, skipping", meme.ID)
		}
	}
	log.Printf("Scan completed: rescheduled %d memes", len(stuckMemes))
}

func (tp *TaskProcessor) AddTask(memeID uuid.UUID) error {
	tp.mu.Lock()
	if tp.processingIDs[memeID] {
		tp.mu.Unlock()
		log.Printf("Task for meme %s is already being processed", memeID)
		return nil
	}
	tp.processingIDs[memeID] = true
	tp.mu.Unlock()

	select {
	case tp.taskQueue <- memeID:
		log.Printf("Task added to queue: meme_id=%s", memeID)
		return nil
	case <-tp.ctx.Done():
		tp.mu.Lock()
		delete(tp.processingIDs, memeID)
		tp.mu.Unlock()
		return fmt.Errorf("task processor is shutting down")
	default:
		tp.mu.Lock()
		delete(tp.processingIDs, memeID)
		tp.mu.Unlock()
		return fmt.Errorf("task queue is full")
	}
}

func (tp *TaskProcessor) worker(id int) {
	defer tp.wg.Done()
	log.Printf("Worker %d started", id)

	for {
		select {
		case <-tp.ctx.Done():
			log.Printf("Worker %d stopped", id)
			return
		case memeID, ok := <-tp.taskQueue:
			if !ok {
				log.Printf("Worker %d: task queue closed", id)
				return
			}
			tp.processTask(id, memeID)
		}
	}
}

func (tp *TaskProcessor) processTask(workerID int, memeID uuid.UUID) {
	defer func() {
		tp.mu.Lock()
		delete(tp.processingIDs, memeID)
		tp.mu.Unlock()
	}()

	log.Printf("Worker %d: processing task for meme %s", workerID, memeID)

	meme, err := tp.memeRepo.GetByID(tp.ctx, memeID)
	if err != nil {
		log.Printf("Worker %d: failed to get meme %s: %v", workerID, memeID, err)
		return
	}

	if meme.Status == "completed" {
		log.Printf("Worker %d: meme %s already completed", workerID, memeID)
		return
	}

	if meme.TaskID == "" {
		log.Printf("Worker %d: meme %s has no task_id", workerID, memeID)
		return
	}

	ticker := time.NewTicker(tp.pollInterval)
	defer ticker.Stop()

	maxAttempts := 120
	attempts := 0

	for {
		select {
		case <-tp.ctx.Done():
			log.Printf("Worker %d: context cancelled, stopping task for meme %s", workerID, memeID)
			return
		case <-ticker.C:
			attempts++
			if attempts > maxAttempts {
				log.Printf("Worker %d: max attempts reached for meme %s - will be retried by scanner", workerID, memeID)
				return
			}

			log.Printf("Worker %d: checking status for meme %s (attempt %d/%d)", workerID, memeID, attempts, maxAttempts)

			taskStatus, err := tp.aiSvc.GetTaskStatus(tp.ctx, meme.TaskID)
			if err != nil {
				log.Printf("Worker %d: failed to check task status for meme %s: %v", workerID, memeID, err)
				continue
			}

			log.Printf("Worker %d: meme %s AI status: %s", workerID, memeID, taskStatus.Status)

			meme.Status = taskStatus.Status
			if err := tp.memeRepo.Update(tp.ctx, meme); err != nil {
				log.Printf("Worker %d: failed to update meme status: %v", workerID, err)
			}

			if taskStatus.Status == "completed" || taskStatus.Status == "SUCCESS" || taskStatus.Status == "success" {
				log.Printf("Worker %d: task completed for meme %s, fetching result", workerID, memeID)
				if err := tp.processCompletedTask(memeID); err != nil {
					log.Printf("Worker %d: failed to process completed task: %v", workerID, err)
					tp.markAsFailed(memeID, fmt.Sprintf("failed to process result: %v", err))
				} else {
					log.Printf("Worker %d: successfully processed meme %s", workerID, memeID)
				}
				return
			}

			if taskStatus.Status == "failed" || taskStatus.Status == "FAILED" || taskStatus.Status == "error" || taskStatus.Status == "ERROR" {
				log.Printf("Worker %d: AI returned failed for meme %s", workerID, memeID)
				tp.markAsFailed(memeID, "AI task failed")
				return
			}
		}
	}
}

func (tp *TaskProcessor) processCompletedTask(memeID uuid.UUID) error {
	meme, err := tp.memeRepo.GetByID(tp.ctx, memeID)
	if err != nil {
		return fmt.Errorf("failed to get meme: %w", err)
	}

	imageData, err := tp.aiSvc.GetTaskResult(tp.ctx, meme.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task result: %w", err)
	}

	objectName := fmt.Sprintf("memes/%s.jpg", meme.ID.String())
	if err := tp.minioSvc.UploadBytes(tp.ctx, objectName, imageData); err != nil {
		return fmt.Errorf("failed to upload image to MinIO: %w", err)
	}

	meme.ImageURL = tp.minioSvc.GetMemeURL(objectName)
	meme.Status = "completed"

	if err := tp.memeRepo.Update(tp.ctx, meme); err != nil {
		tp.minioSvc.DeleteMeme(tp.ctx, objectName)
		return fmt.Errorf("failed to update meme: %w", err)
	}

	return nil
}

func (tp *TaskProcessor) markAsFailed(memeID uuid.UUID, reason string) {
	meme, err := tp.memeRepo.GetByID(tp.ctx, memeID)
	if err != nil {
		log.Printf("Failed to get meme %s for marking as failed: %v", memeID, err)
		return
	}

	meme.Status = "failed"
	if err := tp.memeRepo.Update(tp.ctx, meme); err != nil {
		log.Printf("Failed to mark meme %s as failed: %v", memeID, err)
	}

	log.Printf("Meme %s marked as failed: %s", memeID, reason)
}
