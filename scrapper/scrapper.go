package scrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	ccsv "github.com/tsak/concurrent-csv-writer"
)

type extractedJob struct {
	id string
	title string
	condition string
}

// Scrape 
func Scrape(searchWord string) {
	var baseURL string = "https://www.saramin.co.kr/zf_user/search/recruit?searchword=" + searchWord
	start := time.Now()
	var jobs []extractedJob
	c := make(chan []extractedJob)
	totalPages := getPages(baseURL)
	
	for i := 0; i < totalPages; i++ {
		go getPage(baseURL, i, c)
	}

	for i := 0; i < totalPages; i++ {
		extractedJobs := <-c
		jobs = append(jobs, extractedJobs...) // [x, x, x] not [[x], [x], [x]]
	}

	// 개수가 적으면 크게 차이가 나지 않음? 대량으로 테스트 필요
	// writeJobs(jobs)
	writeJobsConcurrent(jobs)

	elapsed := time.Since(start)
	fmt.Println(elapsed)
	fmt.Println("파일 생성 완료, " + strconv.Itoa(len(jobs)) + " 개 포지션 수집 됨")
}

func getPage(url string, page int, mainC chan<- []extractedJob) {
	var jobs []extractedJob
	c := make(chan extractedJob)
	pageUrl := url + "&recruitPage=" + strconv.Itoa(page+1)
	fmt.Println("Requesting: ", pageUrl)
	res, err := http.Get(pageUrl)
	checkErr(err)
	checkStatusCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	searchCards := doc.Find(".item_recruit")
	searchCards.Each(func(i int, card *goquery.Selection) {
		// job := extractJob(card, c)
		// jobs = append(jobs, job)
		go extractJob(card, c)
	})

	for i := 0; i < searchCards.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}

	mainC <- jobs
}

func extractJob(card *goquery.Selection, c chan<- extractedJob) {
	id, _ := card.Attr("value")
	title := CleanString(card.Find(".job_tit>a").Text())
	condition := CleanString(card.Find(".job_condition").Text())
	// fmt.Println(id, title, condition)
	c <- extractedJob{
		id: id, 
		title: title, 
		condition: condition,
	}
}

// CleanString 문자열 정리
func CleanString(str string) string {
	// Fields 문자열을 분리 , TrimSpace 양쪽 끝에 공백을 제거, Join 배열을 separater 기준 join
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

func getPages(url string) int {
	pages := 0
	res, err := http.Get(url)
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


func writeJobsConcurrent(jobs []extractedJob) {
	csv, err := ccsv.NewCsvWriter("jobs.csv")
	checkErr(err)

	defer csv.Close()

	headers := []string{"ID", "Title", "Condition"}

	wErr := csv.Write(headers)
	checkErr(wErr)

	done := make(chan bool)

	for _, job := range jobs {
		go func(job extractedJob) {
			csv.Write([]string{"https://www.saramin.co.kr/zf_user/jobs/relay/view?rec_idx=" + job.id, job.title, job.condition})
			done <- true
		}(job)
	}

	for i := 0; i < len(jobs); i++ {
		<-done
	}
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