package main

import (
	"database/sql"
	"encoding/csv"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

func connectDatabase() (*sql.DB, error) {
	db, err := sql.Open("mysql", "username:password@tcp(host:port)/dbname")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func exportToCSV(w http.ResponseWriter, query, fileName string) {
	db, err := connectDatabase()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "gagal menjalankan query database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		http.Error(w, "gagal mendapatkan kolom dari hasil query", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename="+fileName)

	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()
	csvWriter.Write(columns)

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			http.Error(w, "Gagal membaca baris dari database", http.StatusInternalServerError)
			return
		}

		record := make([]string, len(values))
		for i, value := range values {
			record[i] = string(value)
		}

		if err := csvWriter.Write(record); err != nil {
			http.Error(w, "Gagal menulis data ke CSV", http.StatusInternalServerError)
			return
		}
	}

	
	if err = rows.Err(); err != nil {
		http.Error(w, "Error saat iterasi baris", http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/export-users-csv", func(w http.ResponseWriter, r *http.Request) {
		exportToCSV(w, "SELECT * FROM users", "users.csv")
	})
	http.HandleFunc("/export-orders-csv", func(w http.ResponseWriter, r *http.Request) {
		exportToCSV(w, "SELECT * FROM user_orders", "orders.csv")
	})
	http.HandleFunc("/export-admins-csv", func(w http.ResponseWriter, r *http.Request) {
		exportToCSV(w, "SELECT * FROM admins", "admins.csv")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
