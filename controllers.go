package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

func AllHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rows, err := DB.Query("SELECT * FROM books;")
	if err != nil {
		log.Println("Database query error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var books []Book

	for rows.Next() {
		var b Book
		err := rows.Scan(&b.Isbn, &b.Title, &b.Author)
		if err != nil {
			log.Println("Row scan error:", err)
			http.Error(w, "Error processing data", http.StatusInternalServerError)
			return
		}
		books = append(books, b)
	}

	err = rows.Err(); 
	if err != nil {
		log.Println("Row iteration error:", err)
		http.Error(w, "Error reading data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(books); err != nil {
		log.Println("Encoding error:", err)
	}
}

func IsbnHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	isbn := mux.Vars(r)["isbn"]
	log.Println("Requested ISBN:", isbn)

	query := `
		SELECT * FROM books 
		WHERE isbn = $1;
	`
	var b Book

	err := DB.QueryRow(query, isbn).Scan(&b.Isbn, &b.Title, &b.Author)
	if err != nil {
		log.Println("Book not found:", err)
		http.Error(w, "Book not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(b);
	fmt.Println("HURRRAY")
	if err != nil {
		log.Println("Encoding error:", err)
	}
}

func NewHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var b Book
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		log.Println("JSON decode error:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO books (isbn, title, author)
		VALUES ($1, $2, $3)
		RETURNING isbn;
	`

	var returnedIsbn string
	err := DB.QueryRow(query, b.Isbn, b.Title, b.Author).Scan(&returnedIsbn)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			// 23505 is the code for unique_violation in PostgreSQL
			log.Println("Duplicate ISBN:", b.Isbn)
			http.Error(w, "Book with this ISBN already exists", http.StatusConflict)
			return
		}

		log.Println("Insert query error:", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	log.Println("Added book with ISBN:", returnedIsbn)

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message": "Book added successfully",
		"isbn":    returnedIsbn,
	}); err != nil {
		log.Println("Encoding error:", err)
	}
}



func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	isbn := mux.Vars(r)["isbn"]
	log.Println("Updating book with ISBN:", isbn)

	var updatedBook Book
	err := json.NewDecoder(r.Body).Decode(&updatedBook)

	if err != nil {
		log.Println("Failed to decode the request body:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	queryStatement := `
		UPDATE books
		SET title = $1, author = $2
		WHERE isbn = $3
		RETURNING isbn, title, author;
	`

	var returnedIsbn, returnedTitle, returnedAuthor string
	err = DB.QueryRow(queryStatement, updatedBook.Title, updatedBook.Author, isbn).Scan(&returnedIsbn, &returnedTitle, &returnedAuthor)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("No book found with the given ISBN")
			http.Error(w, "Book not found", http.StatusNotFound)
		} else {
			log.Println("Error executing query:", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"message": "Book updated successfully",
		"isbn":    returnedIsbn,
		"title":   returnedTitle,
		"author":  returnedAuthor,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println("Encoding error:", err)
	}
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	isbn := mux.Vars(r)["isbn"]
	log.Println("Deleting book with ISBN:", isbn)

	queryStatement := `
		DELETE FROM books
		WHERE isbn = $1
		RETURNING isbn;
	`

	var deletedIsbn string
	err := DB.QueryRow(queryStatement, isbn).Scan(&deletedIsbn)

	if err != nil {
		if err == sql.ErrNoRows {
			log.Println("No book found with the given ISBN")
			http.Error(w, "Book not found", http.StatusNotFound)
		} else {
			log.Println("Error executing query:", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	response := map[string]string{
		"message": "Book deleted successfully",
		"isbn":    deletedIsbn,
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println("Encoding error:", err)
	}
}