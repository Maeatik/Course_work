package main

import (
	"MortyGRAB/models"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"log"
	"net/http"
)

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/parse", models.PageSitesGETHandler).Methods("GET")
	r.HandleFunc("/parse", models.PageSitePOSTHandler).Methods("POST")
	r.HandleFunc("/parse/{id}", models.PageSiteGETHandler).Methods("GET")
	r.HandleFunc("/parse/{id}", models.PageSiteDELETEHandler).Methods("DELETE")
	r.HandleFunc("/parse/{id}", models.PageSitePUTHandler).Methods("PUT")


	handler := cors.AllowAll().Handler(r)
	log.Fatal(http.ListenAndServe(":3000", handler))


}

