package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	id string
	title string
	condition string
}

var baseURL string = "https://www.saramin.co.kr/zf_user/search/recruit?searchword=django"

func main() {
	var jobs []extractedJob
	totalPages := getPages()
	
	for i := 0; i < totalPages; i++ {
		extractedJobs := getPage(i)
		jobs = append(jobs, extractedJobs...) // [x, x, x] not [[x], [x], [x]]

	}
	writeJobs(jobs)
	fmt.Println("파일 생성 완료")
}

func writeJobs(jobs []extractedJob) {
	file, err := os.Create("jobs.csv")
	checkErr(err)

	w := csv.NewWriter(file)
	defer w.Flush() //function 내 작업 완료 후 defer flush(파일생성)

	headers := []string{"ID", "Title", "Condition"}

	wErr := w.Write(headers)
	checkErr(wErr)

	for _, job := range jobs {
		jobSlice := []string{"https://www.saramin.co.kr/zf_user/jobs/relay/view?rec_idx=" + job.id, job.title, job.condition}
		jwErr := w.Write(jobSlice)
		checkErr(jwErr)
	}

}

func getPage(page int) []extractedJob {
	var jobs []extractedJob
	pageUrl := baseURL + "&recruitPage=" + strconv.Itoa(page+1)
	fmt.Println("Requesting: ", pageUrl)
	res, err := http.Get(pageUrl)
	checkErr(err)
	checkStatusCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".item_recruit")
	searchCards.Each(func(i int, card *goquery.Selection) {
		job := extractJob(card)
		jobs = append(jobs, job)
	})

	return jobs
}

func extractJob(card *goquery.Selection) extractedJob {
	id, _ := card.Attr("value")
	title := cleanString(card.Find(".job_tit>a").Text())
	condition := cleanString(card.Find(".job_condition").Text())
	// fmt.Println(id, title, condition)
	return extractedJob{
		id: id, 
		title: title, 
		condition: condition,
	}
}

func cleanString(str string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

func getPages() int {
	pages := 0
	res, err := http.Get(baseURL)
	checkErr(err)
	checkStatusCode(res)

	defer res.Body.Close() // 메모리 샘 방지
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)
    
	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})
	return pages
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func checkStatusCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
}