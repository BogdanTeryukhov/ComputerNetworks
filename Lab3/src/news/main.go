package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mmcdole/gofeed"
)

const (
	tableName string = "iu9teryukhov"
	password  string = "Je2dTYr6"
	login     string = "iu9networkslabs"
	host      string = "students.yss.su"
	dbname    string = "iu9networkslabs"
	URL       string = "https://news.rambler.ru/rss/Namibia/"
)

func main() {
	db, err := sql.Open("mysql", login+":"+password+"@tcp("+host+")/"+dbname) //открывает sql соединение (подключаемся на сервер)
	if err != nil {
		fmt.Println("Database conn failed!")
		//panic(err)
	}
	defer db.Close()

	fp := gofeed.NewParser()
	feed, err1 := fp.ParseURL(URL) //парсим rss файл

	if err1 != nil {
		fmt.Println("Site parse failed!")
		//panic(err1)
	}

	var query = "INSERT INTO " + tableName + " (title, date, link) values "
	for _, item := range feed.Items { //обход массива новостей (item)
		query += "('" + item.Title + "', '" + item.Updated + "', '" + item.Link + "')," //ОДНА НОВОСТЬ
	}
	query = query[:len(query)-1]

	_, err2 := db.Exec(query + ";")

	if err2 != nil {
		fmt.Println("No")
	} else {
		fmt.Println("Data added in table: " + tableName)
	}
}
