package models

import "sync"

type JobStatus string

const (
	JobStatusParsing JobStatus = "parsing"
	JobStatusReady   JobStatus = "ready"
	JobStatusError   JobStatus = "error"
)

type ParseJob struct {
	ID       string    `json:"id"`
	Status   JobStatus `json:"status"`
	Error    string    `json:"error,omitempty"`
	Progress float32   `json:"progress"`
}

// Simple in-memory store for tracking parse jobs.
type JobStore struct {
	mu   sync.RWMutex
	jobs map[string]*ParseJob
}

func NewJobStore() *JobStore {
	return &JobStore{jobs: make(map[string]*ParseJob)}
}

func (s *JobStore) Create(id string) *ParseJob {
	s.mu.Lock()
	defer s.mu.Unlock()

	job := &ParseJob{ID: id, Status: JobStatusParsing}
	s.jobs[id] = job
	return job
}

func (s *JobStore) Get(id string) (*ParseJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, ok := s.jobs[id]
	return job, ok
}

func (s *JobStore) SetProgress(id string, progress float32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if job, ok := s.jobs[id]; ok {
		job.Progress = progress
	}
}

func (s *JobStore) Complete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if job, ok := s.jobs[id]; ok {
		job.Status = JobStatusReady
		job.Progress = 1.0
	}
}

func (s *JobStore) Fail(id string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if job, ok := s.jobs[id]; ok {
		job.Status = JobStatusError
		job.Error = err.Error()
	}
}
