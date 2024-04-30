package queue

import (
	"log"
	"tqs/internal/database"
)

const (
	MaxQueueSize = 500
	MaxWorkers   = 25
)

type Job struct {
	SurveyId int
	AnswerId int
}

type Queue struct {
	isAvailable bool
	database    *database.Database

	jobs chan Job
}

func New(database *database.Database) *Queue {
	// useQueue := os.Getenv("USE_QUEUE") == "true"
	useQueue := true

	if !useQueue {
		return &Queue{isAvailable: false}
	}

	jobs := make(chan Job, MaxQueueSize)

	for i := 0; i < MaxWorkers; i++ {
		go func() {
			for job := range jobs {
				err := database.AddVote(job.SurveyId, job.AnswerId)
				if err != nil {
					log.Println(err)
				}
			}
		}()
	}

	return &Queue{true, database, jobs}
}

func (queue *Queue) IsAvailable() bool {
	return queue.isAvailable
}

func (queue *Queue) AddJob(job Job) {
	go func() {
		queue.jobs <- job
	}()
}
