package middleware

import (
	"aiven-connect-to-pg/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"os"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
)

var db *sql.DB

func InitDB() {
    var err = godotenv.Load(".env")
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    dbURI := os.Getenv("DB_URI")

    // Parse the database URI
    conn, _ := url.Parse(dbURI)
    conn.RawQuery = "sslmode=verify-ca;sslrootcert=ca.pem"

    var e error
    db, e = sql.Open("postgres", conn.String())
    if e != nil {
        log.Fatal(e)
    }

    e = db.Ping()
    if err != nil {
        log.Fatal(e)
    }

    fmt.Println("Successfully connected to the database")
    createTableQuery := `
    CREATE TABLE IF NOT EXISTS stocks (
        stockid SERIAL PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        price INT NOT NULL,
        company VARCHAR(100) NOT NULL
    );`

    _, err = db.Exec(createTableQuery)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Table 'stocks' created or already exists")
}
func GetStock(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid stock ID", http.StatusBadRequest)
		return
	}

	var stock models.Stock
	err = db.QueryRow("SELECT stockid, name, price, company FROM stocks WHERE stockid = $1", id).Scan(&stock.StockID, &stock.Name, &stock.Price, &stock.Company)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Stock not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(stock)
}

func GetAllStock(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT stockid, name, price, company FROM stocks")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var stocks []models.Stock
	for rows.Next() {
		var stock models.Stock
		err := rows.Scan(&stock.StockID, &stock.Name, &stock.Price, &stock.Company)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		stocks = append(stocks, stock)
	}

	json.NewEncoder(w).Encode(stocks)
}

func CreateStock(w http.ResponseWriter, r *http.Request) {
	var stock models.Stock
	err := json.NewDecoder(r.Body).Decode(&stock)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	err = db.QueryRow("INSERT INTO stocks (name, price, company) VALUES ($1, $2, $3) RETURNING stockid", stock.Name, stock.Price, stock.Company).Scan(&stock.StockID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stock)
}

func UpdateStock(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid stock ID", http.StatusBadRequest)
		return
	}

	var stock models.Stock
	err = json.NewDecoder(r.Body).Decode(&stock)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE stocks SET name = $1, price = $2, company = $3 WHERE stockid = $4", stock.Name, stock.Price, stock.Company, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func DeleteStock(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid stock ID", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("DELETE FROM stocks WHERE stockid = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
