package models

import (
	"encoding/json"
	"fmt"
	goose "github.com/advancedlogic/GoOse"
	"github.com/beevik/ntp"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type GrabbedText struct {
	Id       	string 		`json:"id"`
	Url 		string		`json:"url"`
	GrabText	string 		`json:"grabtext""`
	GrabDate	time.Time 	`json:"grab_date"`
}

var pagesite1 []GrabbedText

func PrintCurrentTime() time.Time{
	//получаем данные о текущем времени с NTP сервера
	time, err := ntp.Time("ntp5.stratum2.ru")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err: %v\n", err)
		os.Exit(1)
	}
	return time
	//Форматировано выводим текущее время
	//fmt.Println(time.Format("Сейчас: 15:04:05 MST"))
}

func htmlExtractor(p GrabbedText) string {
	g := goose.New()

	grabbing, _ := g.ExtractFromURL(p.Url)
	grabText := grabbing.CleanedText
	extraLinks := parsePages(grabbing.Links, g)

	builder := strings.Builder{}
	builder.Grow(len(grabText) + len(extraLinks))
	builder.WriteString(grabText)
	builder.WriteString(extraLinks)

	text := builder.String()

	return text
}


func parsePages(pages []string, g goose.Goose) string{
	var articlePage *goose.Article
	var err error
	for _, page := range pages{
		articlePage, err = g.ExtractFromURL(page)
		if err != nil{
			fmt.Println(page, ": ошибка обработки страницы")
			continue
		}
	}
	return articlePage.CleanedText
}

func PageSitesGETHandler(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()
	pagesite1 = nil

	rows, err := db.Query("SELECT * FROM grabbedtexts ORDER BY id")
	if err != nil {
		log.Fatal(err)
	}
	var pageBytes []byte
	for rows.Next() {
		var page GrabbedText
		err = rows.Scan(&page.Id, &page.Url, &page.GrabText, &page.GrabDate)
		if err != nil {
			log.Fatal(err)
		}

		pagesite1 = append(pagesite1, GrabbedText{page.Id, page.Url, page.GrabText, page.GrabDate})
	}
	pageBytes, _ = json.MarshalIndent(pagesite1, "", "\t")

	w.Header().Set("Content-Type", "application/json")
	w.Write(pageBytes)

	defer rows.Close()
	defer db.Close()

}


func PageSitePOSTHandler(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()

	var p GrabbedText
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	grabText := htmlExtractor(p)
	date := PrintCurrentTime().Format("2006-01-02 15:04:05")

	sqlStatement := `INSERT INTO grabbedtexts (url, grabtext, grabdate) VALUES ($1, $2, $3)`

	_, err = db.Exec(sqlStatement, p.Url, grabText, date)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	defer db.Close()
}

func PageSiteGETHandler(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()

	params := mux.Vars(r)
	idGet := params["id"]
	row, err := db.Query("SELECT * FROM grabbedtexts where id=$1", idGet)
	if err != nil {
		log.Fatal(err)
	}

	var page GrabbedText
	if row.Next() {
		err = row.Scan(&page.Id, &page.Url, &page.GrabText, &page.GrabDate)
		if err != nil {
			log.Fatal(err)
		}
		userId := GrabbedText{page.Id, page.Url, page.GrabText, page.GrabDate}

		pageBytes, _ := json.Marshal(userId)
		w.Header().Set("Content-Type", "application/json")
		w.Write(pageBytes)
	}

	defer row.Close()
	defer db.Close()

}

func PageSiteDELETEHandler(w http.ResponseWriter, r *http.Request)  {
	db := OpenConnection()
	params := mux.Vars(r)
	idGet := params["id"]
	fmt.Println(idGet)
	_, err := db.Exec("DELETE FROM grabbedtexts where id=$1", idGet)

	if err != nil{
		w.WriteHeader(http.StatusBadRequest)
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	defer db.Close()
}

func PageSitePUTHandler(w http.ResponseWriter, r *http.Request)  {
	fmt.Print(1)
	db := OpenConnection()
	params := mux.Vars(r)
	idGet := params["id"]

	var p GrabbedText

	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sqlStatement := `UPDATE grabbedtexts SET url = $1, grabtext = $2, grabdate = $3 WHERE id = $4`
	_, err = db.Exec(sqlStatement, p.Id, p.Url, p.GrabText, p.GrabDate, idGet)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)

	defer db.Close()
}
