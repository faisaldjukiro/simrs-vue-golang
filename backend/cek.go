package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "simrs:@simrs2023@tcp(192.168.20.1:3306)/sik")
	if err != nil { log.Fatal(err) }
	
	rows, err := db.Query("SELECT kode_sidebar, nama FROM sidebar_pasien WHERE nama LIKE '%Awal Medis%'")
	if err != nil { log.Fatal(err) }
	defer rows.Close()

	for rows.Next() {
		var kode, nama sql.NullString
		if err := rows.Scan(&kode, &nama); err != nil { log.Fatal(err) }
		fmt.Printf("%s: %s\n", kode.String, nama.String)
	}
}
