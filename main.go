package main

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

type Activity struct {
	ID           int       `json:"id"`
	Title        string    `json:"title" validate:"required"`
	Category     string    `json:"category" validate:"required,oneof= TASK EVENT"`
	Description  string    `json:"description" validate:"required"`
	ActivityDate time.Time `json:"activity_date" validate:"required"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

func initDB() (*pgx.Conn, error) {
	dsn := "postgres://postgres.geupriplpyktgjschjnj:@Bdw9488!!!@aws-1-ap-southeast-1.pooler.supabase.com:6543/postgres?sslmode=require"
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func main() {
	db, err := initDB()
	if err != nil {
		panic(err)
	}
	defer db.Close(context.Background())

	app := fiber.New()
	validate := validator.New()

	app.Get("/activities", func(c *fiber.Ctx) error {
		rows, err := db.Query(context.Background(), "SELECT * FROM activities")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}
		defer rows.Close()

		var activities []Activity
		for rows.Next() {
			var activity Activity
			err = rows.Scan(&activity.ID, &activity.Title, &activity.Category, &activity.Description, &activity.ActivityDate, &activity.Status, &activity.CreatedAt)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
			}
			activities = append(activities, activity)
		}

		return c.Status(fiber.StatusOK).JSON(activities)
	})

	app.Post("/activities", func(c *fiber.Ctx) error {
		var activity Activity
		err := c.BodyParser(&activity)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}

		if err = validate.Struct(&activity); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": err.Error()})
		}

		sqlStatement := `INSERT INTO activities(title, category, description, activity_date, status)
			VALUES($1, $2, $3, $4, $5) RETURNING id`
		err = db.QueryRow(context.Background(), sqlStatement, activity.Title, activity.Category, activity.Description, activity.ActivityDate, "NEW").Scan(&activity.ID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": err.Error()})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success"})
	})

	app.Listen(":8081")
}
