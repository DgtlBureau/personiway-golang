package main

import (
	"DgtlBureau/personiway-golang/internal/controllers"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	e.POST("/convert-pdf", func(c echo.Context) error {
		form, err := c.MultipartForm()

		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid form data"})
		}

		files := form.File["pdf"]

		if len(files) == 0 {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "No files uploaded"})
		}

		controllers.RunConvert(files)

		return c.JSON(200, map[string]bool{"success": true})
	})

	e.Logger.Fatal(e.Start(":1323"))
}
