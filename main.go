package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"net/http"

	"./conf"
	"./sronacional/modelo"

	r "github.com/dancannon/gorethink"
	_ "github.com/mattn/go-oci8"

)

func main() {

	err := conf.Load(`./conf/conf.json`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "teste")
	loga("Iniciando...")
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "teste1")

	session, err1 := r.Connect(r.ConnectOpts{
		Address:  conf.Conf.RethinDB.Address,
		Database: conf.Conf.RethinDB.Database,
	})

	if err1 != nil {
		log.Panic(err1.Error())
	}

	router := NewRouter(session)

	//modelo.RethinkDB(session)
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), session)

	db, err2 := sql.Open("oci8", conf.Conf.DnsOracle)
	if err2 != nil {
		erroFatal(err2.Error())
	}
	modelo.Db(db, session)

	ini, erri := time.Parse("2006-01-02 15:04:05 -0700", "2019-04-23 09:00:00 -0300")
	if erri != nil {
		erroFatal(erri.Error())
	}

	
	http.Handle("/", router)
	http.ListenAndServe(":4000", nil)

	fmt.Println("ini ",ini)

	iniE := ini


	go modelo.SelecionaUnitizadores(ini, session)
	if err != nil {
		erroFatal(err.Error())
	}

	go modelo.SelecionaEventos(iniE, session)
	if err != nil {
		erroFatal(err.Error())
	}

	ch := make(chan struct{})
	<-ch

}

func loga(msg string) {
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), msg)
}

func erroFatal(msg string) {
	log.Fatalln(time.Now().Format("2006-01-02 15:04:05"), msg)
}
