package app

import (
	"fmt"
	"tqs/internal/cache"
	"tqs/internal/database"
	"tqs/internal/queue"

	"encoding/json"
	"net/http"
)

type VoteRequest struct {
	SurveyId int `json:"surveyId"`
	AnswerId int `json:"answerId"`
}

type App struct {
	database *database.Database
	cache    *cache.Cache
	queue    *queue.Queue
}

func New() (*App, error) {
	database, err := database.New()
	if err != nil {
		return nil, err
	}

	cache := cache.New()
	queue := queue.New(database)

	return &App{database, cache, queue}, nil
}

func (app *App) VoteSurvey(response http.ResponseWriter, request *http.Request) {
	var voteRequest VoteRequest
	err := json.NewDecoder(request.Body).Decode(&voteRequest)
	if err != nil {
		http.Error(response, "Invalid request body", http.StatusBadRequest)
		return
	}

	errCode, err := app.validId(voteRequest.SurveyId, "survey")
	if err != nil {
		http.Error(response, err.Error(), errCode)
		return
	}

	errCode, err = app.validId(voteRequest.AnswerId, "answer")
	if err != nil {
		http.Error(response, err.Error(), errCode)
		return
	}

	if app.queue.IsAvailable() {
		app.queue.AddJob(queue.Job{SurveyId: voteRequest.SurveyId, AnswerId: voteRequest.AnswerId})
		return
	}

	err = app.database.AddVote(voteRequest.SurveyId, voteRequest.AnswerId)
	if err != nil {
		http.Error(response, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (app *App) validId(id int, objectName string) (errCode int, err error) {
	if app.cache.IdExists(id, objectName) {
		return 0, nil
	}

	exists, err := app.database.IdExists(id, objectName)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if !exists {
		return http.StatusNotFound, fmt.Errorf("%s ID not found", objectName)
	}

	app.cache.AddId(id, objectName)

	return 0, nil
}
