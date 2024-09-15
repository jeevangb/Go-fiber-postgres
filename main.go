package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/jeevangb/go-fiber-postgres/models"
	"github.com/jeevangb/go-fiber-postgres/storage"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type Book struct {
	Author    string `json:"author"`
	Title     string `json:"title"`
	Publisher string `json:"publisher"`
}

type Repository struct {
	DB *gorm.DB
}

func (r *Repository) CreateBook(context *fiber.Ctx) error {
	book := Book{}
	err := context.BodyParser(&book)
	if err != nil {
		context.Status((http.StatusUnprocessableEntity)).JSON(&fiber.Map{"message": "request failed"})
		return err
	}
	err = r.DB.Create(&book).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "could not create book"})
		return err
	}
	context.Status((http.StatusOK)).JSON(fiber.Map{"message": "book has been added successfully"})
	return nil
}

func (r *Repository) GetBooks(context *fiber.Ctx) error {
	bookModels := &[]models.Books{}
	err := r.DB.Find(&bookModels).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "could not get books"})
		return err
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{"message": "Books fetched successfully",
		"data": bookModels,
	})
	return nil

}

func (r *Repository) DeleteBook(context *fiber.Ctx) error {
	bookModels := models.Books{}
	id := context.Params("id")
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": "id cannot be empty"})
		return nil
	}

	err := r.DB.Delete(bookModels, id).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "could not delete book"})
		return err
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{"message": "book has been deleted successfully"})
	return nil
}

func (r *Repository) GetBookById(context *fiber.Ctx) error {
	id := context.Params("id")
	bookModel := &models.Books{}
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": "id cannot be empty"})
		return nil
	}
	fmt.Println("the id is: ", id)
	err := r.DB.Where("id =?", id).First(&bookModel).Error
	if err != nil {
		context.Status(http.StatusNotFound).JSON(&fiber.Map{"message": "book not found"})
		return err
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{"message": "book id fetched successfully",
		"data": bookModel,
	})
	return nil
}

func (r *Repository) SetupRotes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/create_books", r.CreateBook)
	api.Delete("delete_book/:id", r.DeleteBook)
	api.Get("/get_books/:id", r.GetBookById)
	api.Get("/get_all_books", r.GetBooks)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASSWD"),
		User:     os.Getenv("DB_USER"),
		SSLMode:  os.Getenv("DB_SSL"),
		DBName:   os.Getenv("DB_NAME"),
	}
	db, err := storage.NewConnection(config)
	if err != nil {
		log.Fatal("could not load the database")
	}
	err = models.MigrateBooks(db)
	if err != nil {
		log.Fatal("could not migrate the database")
	}
	r := Repository{
		DB: db,
	}
	app := fiber.New()
	r.SetupRotes(app)
	app.Listen(":8080")
}
