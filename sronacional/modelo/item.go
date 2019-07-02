package modelo

import (
	"fmt"
	"log"
	"time"

	 r "github.com/dancannon/gorethink"

)

type Item struct {
	Id           string `gorethink:"id"`
	IdUnitizador           string `gorethink:"id_unitizador"`
	Numero       string `gorethink:"numero"`	
	Status       string `gorethink:"status"`
	DataCriacao      time.Time `gorethink:"datacriacao"`
	NumeroUnitizador string `gorethink:"numero_unitizador"`
	IdUnidadeDestino string    `gorethink:"iddestino"`
}

func InsereUnitizador(unitizadores []Unitizador,session *r.Session) {
	ch := make(chan int, 10)
	
    for _, u := range unitizadores {		

		err := r.Table("unitizador").
			Insert(u).
			Exec(session)
		if err != nil {
			log.Fatal(err)
		}

		go InsereItens(u, ch, session)
		
	}
	
	
}

func InsereItens(u Unitizador, ch chan int, session *r.Session) {
	
	ch <- len(ch)
	fmt.Println("go inserwitens init ", len(ch))
	   itens, err := SelecionaItensDoUnitizador(u.Id)
	   if err != nil {
		  erroFatal(err.Error())
	   }

	   u.QuantidadeItens = len(itens)
		
	   if len(itens) > 0 {
		_, err = r.Table("unitizador").
		Get(u.Id).
		Update(u).
		RunWrite(session)
		if err != nil {
			log.Fatal(err)
		}

		for _, i := range itens {
			err = r.Table("objeto").
			Insert(i).
			Exec(session)
		if err != nil {
			log.Fatal(err)
		}

		}

	   } else {
		_, err = r.Table("unitizador").
			Get(u.Id).
			Delete().
			Run(session)
			if err != nil {
				log.Fatal(err)
			}

	   }
	   
	   fmt.Println("go inserwitens end ", len(ch), <- ch)
	  
}

func SelecionaItensDoUnitizador(idunitizador string) ([]Item, error) {
	
	query := `SELECT p.PEN_OBJECTNUMBER || '-' || p.TAE_CODE AS id,
	                 p.tae_code, 
					 p.PEN_OBJECTNUMBER, 
					 p.PEN_IN_TRATADO,
					 t.TAE_CREATETIME, 
					 t.TAE_RECEPTACLE, 
					 t.TAE_UNICEPDES
			 FROM SRO.EVENTPENDENT p, SRO.TECALERTEXPEDITION t
			 WHERE p.TAE_CODE = :a 
			       and p.TAE_CODE = t.TAE_CODE
	`
	rows, err := db.Query(query, idunitizador)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	itens := make([]Item, 0)

	for rows.Next() {
		i := Item{}
		err = rows.Scan(
			&i.Id,
			&i.IdUnitizador,
			&i.Numero,
			&i.Status,
			&i.DataCriacao,
			&i.NumeroUnitizador,
			&i.IdUnidadeDestino,
		)

		if err != nil {
			return nil, err
		}
		
		if EtiquetaValida(i.Numero) {
			itens = append(itens, i)			
		}
	    
	}

	fmt.Println("itens ", idunitizador, len(itens))
	return itens, nil
}


func erroFatal(msg string) {
	log.Fatalln(time.Now().Format("2006-01-02 15:04:05"), msg)
}

func InitItem() {
	fmt.Println("item init", session)
}