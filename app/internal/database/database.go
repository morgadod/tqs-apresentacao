package database

import (
	"context"
	"fmt"

	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	*pgxpool.Pool
}

func New() (*Database, error) {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	name := os.Getenv("DATABASE_NAME")

	// user := "tqs_user"
	// password := "tqs_password"
	// host := "localhost"
	// port := "5432"
	// name := "tqs_db"

	dbUrl := "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + name + "?sslmode=disable"

	config, err := pgxpool.ParseConfig(dbUrl)
	if err != nil {
		return nil, err
	}

	config.MaxConns = 10
	config.MinConns = 0
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30
	config.HealthCheckPeriod = time.Minute
	config.ConnConfig.ConnectTimeout = time.Second * 5

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	db := &Database{pool}

	err = db.runMigrations()
	if err != nil {
		return nil, err
	}

	err = db.runSeed()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *Database) Close() error {
	db.Pool.Close()
	return nil
}

func (db *Database) runMigrations() error {
	_, err := db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS surveys (
			id SERIAL PRIMARY KEY,
			question TEXT NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS survey_answers (
			id SERIAL PRIMARY KEY,
			survey_id INT NOT NULL,
			answer TEXT NOT NULL,

			CONSTRAINT fk_survey_id
				FOREIGN KEY (survey_id)
				REFERENCES surveys (id)
				ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	_, err = db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS survey_votes (
			id SERIAL PRIMARY KEY,
			survey_id INT NOT NULL,
			answer_id INT NOT NULL,

			CONSTRAINT fk_survey_id
				FOREIGN KEY (survey_id)
				REFERENCES surveys (id)
				ON DELETE CASCADE,

			CONSTRAINT fk_answer_id
				FOREIGN KEY (answer_id)
				REFERENCES survey_answers (id)
				ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

func (db *Database) runSeed() error {
	surveysHasData := false
	row := db.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM surveys)")
	err := row.Scan(&surveysHasData)
	if err != nil {
		return err
	}

	answersHasData := true
	row = db.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM survey_answers)")
	err = row.Scan(&answersHasData)
	if err != nil {
		return err
	}

	if !surveysHasData {
		_, err := db.Exec(context.Background(), `
			INSERT INTO surveys (id, question) VALUES
			(1, 'What is your favorite programming language?')
		`)
		if err != nil {
			return err
		}
	}

	if !answersHasData {
		_, err = db.Exec(context.Background(), `
			INSERT INTO survey_answers (id, survey_id, answer) VALUES
			(1, 1, 'Go'),
			(2, 1, 'Python'),
			(3, 1, 'JavaScript')
		`)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Database) IdExists(id int, objectName string) (bool, error) {
	var table string

	if objectName == "survey" {
		table = "surveys"
	} else {
		table = "survey_answers"
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = %d)", table, id)
	row := db.QueryRow(ctx, query)
	err := row.Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (db *Database) AddVote(surveyID, answerID int) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	args := pgx.NamedArgs{"surveyID": surveyID, "answerID": answerID}
	_, err := db.Exec(ctx, "INSERT INTO survey_votes (survey_id, answer_id) VALUES (@surveyID, @answerID)", args)
	if err != nil {
		return err
	}

	return nil
}
