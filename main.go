package main

import (
	"os"
	"strings"

	"github.com/hylim9/scrappergo/scrapper"
	"github.com/labstack/echo/v4"
)

const fileName string = "jobs.csv"

func handleHome(c echo.Context) error {
	// return c.String(http.StatusOK, "Hello, World!")
	return c.File("home.html")
}

func handleScrape(c echo.Context) error {
	defer os.Remove(fileName) // 완료 후 파일 삭제
	searchWord := strings.ToLower(scrapper.CleanString(c.FormValue("searchWord")))
	scrapper.Scrape(searchWord)
	return c.Attachment(fileName, fileName) // 퍼알 저장 (파일명, 저장파일 명)
}

func main() {
	e := echo.New()
	e.GET("/", handleHome)
	e.POST("/scrape", handleScrape)
	e.Logger.Fatal(e.Start(":1323"))
}