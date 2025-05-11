package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
    "github.com/rs/cors"
    _"github.com/lib/pq"
)

var DB *sql.DB

const (
    HOST = "localhost"
    PORT = 5432
    USER = "tahsin"
    PASSWORD = "password"
    DBNAME = "bookstoreDB"
)
func homeHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Welcome to the Bookstore API")
}

func main() {
    connString := fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        HOST, PORT, USER, PASSWORD, DBNAME,
    )

    var err error
    DB, err = sql.Open("postgres", connString)

    if err != nil {
        log.Fatal(err)
    }
    defer DB.Close()

    r := mux.NewRouter()

    r.HandleFunc("/", homeHandler)

    booksSubR := r.PathPrefix("/books").Subrouter()

    booksSubR.HandleFunc("/all", AllHandler).Methods(http.MethodGet)
    booksSubR.HandleFunc("/{isbn:[0-9]{13}}", IsbnHandler).Methods(http.MethodGet)
    booksSubR.HandleFunc("/new", NewHandler).Methods(http.MethodPost)
    booksSubR.HandleFunc("/update/{isbn:[0-9]{13}}", UpdateHandler).Methods(http.MethodPut)
    booksSubR.HandleFunc("/delete/{isbn:[0-9]{13}}", DeleteHandler).Methods(http.MethodDelete)


    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:5173"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders:   []string{"Content-Type"},
        AllowCredentials: true,
    })

    fmt.Println("Server running on port 8080....")
    handler := c.Handler(r)
    log.Fatal(http.ListenAndServe(":8080", handler))
}

