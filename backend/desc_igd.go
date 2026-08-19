package main
import (
	"database/sql"
	"fmt"
	"log"
	_ "github.com/go-sql-driver/mysql"
)
func main() {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/simrs-golang")
	if err != nil { log.Fatal(err) }
	defer db.Close()
	rows, err := db.Query("DESCRIBE penilaian_medis_igd")
	if err != nil { log.Fatal(err) }
	defer rows.Close()
	for rows.Next() {
		var field, typ, null, key, extra string
		var def sql.NullString
		if err := rows.Scan(&field, &typ, &null, &key, &def, &extra); err != nil { log.Fatal(err) }
		fmt.Printf("%s %s\n", field, typ)
	}
}
