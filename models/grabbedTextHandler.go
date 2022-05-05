package models

import (
	"encoding/json"
	"errors"
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

type ParseURL struct {
	Url string `json:"url"`
}
type GrabbedText struct {
	Id       string    `json:"id"`
	Url      string    `json:"url"`
	GrabText string    `json:"grabtext"`
	GrabDate time.Time `json:"grab_date"`
}

var pagesite1 []GrabbedText

func PrintCurrentTime() time.Time {
	time, err := ntp.Time("ntp5.stratum2.ru")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Err: %v\n", err)
		os.Exit(1)
	}
	return time
}

func htmlExtractor(p ParseURL) (string, error) {
	g := goose.New()
	grabbing, err := g.ExtractFromURL(p.Url)
	if err != nil {
		return "", err
	}
	grabText := grabbing.CleanedText
	fmt.Println("text:", grabText)
	if grabText == "" {
		fmt.Println("text:", grabText)
		return "", errors.New("недоступная ссылка")
	}
	fmt.Println(grabText)
	builder := strings.Builder{}

	if len(grabbing.Links) != 0 {
		extraLinks := parsePages(grabbing.Links, g)
		builder.Grow(len(grabText) + len(extraLinks))
		builder.WriteString(grabText)
		builder.WriteString(extraLinks)
	}
	builder.Grow(len(grabText))
	builder.WriteString(grabText)
	text := builder.String()

	return text, nil
}

func parsePages(pages []string, g goose.Goose) string {
	var articlePage *goose.Article
	var err error
	for _, page := range pages {
		articlePage, err = g.ExtractFromURL(page)
		if err != nil {
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
	var p ParseURL
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	grabText, err := htmlExtractor(p)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
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

func PageSiteDELETEHandler(w http.ResponseWriter, r *http.Request) {
	db := OpenConnection()
	params := mux.Vars(r)
	idGet := params["id"]
	_, err := db.Exec("DELETE FROM grabbedtexts where id=$1", idGet)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
	defer db.Close()
}

func PageSitePUTHandler(w http.ResponseWriter, r *http.Request) {
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
	_, err = db.Exec(sqlStatement, p.Url, p.GrabText, p.GrabDate, idGet)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)

	defer db.Close()
}
