package modelo

import (
	"fmt"
	"time"
	"log"

	
	r "github.com/dancannon/gorethink"
)

type Evento struct {
	Id               string    `gorethink:"id"`
	NumeroObjeto        string    `gorethink:"numero_objeto"`
	DataCriacao      time.Time `gorethink:"datacriacao"`
	DataInsercao        time.Time `gorethink:"datainsercao"`
	Tipo           string    `gorethink:"tipo"`
	IdUnidadeOrigem  string    `gorethink:"idorigem"`	
	NumeroUnitizador     string    `gorethink:"numero_unitizador"`
	Situacao            int       `gorethink:"situacao"`
}

func SelecionaEventos(ini time.Time, s *r.Session) {
	
	for {
		

		t0 := time.Now()
		fim := ini.Add(time.Minute * 5)
		now := time.Now()

		fmt.Println("datas ", ini, fim,t0,s)

		if ini.Unix() >= now.Unix() {
			ini = now.Add(time.Duration(-1) * 5)
			fim = now.Add(time.Duration(-1) * 1)
		} else if fim.Unix() > now.Unix() {
		fim = now
		}

		eventos := make([]Evento, 0)
		
		query := `select evt_itemcode || evt_hitunitcep || evt_recordid,
		                 evt_itemcode, EVT_CREATETIME, EVT_INSERTTIME,
		                 evt_code, evt_hitunitcep, EVT_RECEPTACLE,
		                 EVT_RECORDTYPE
				from sro.event 
				where EVT_INSERTTIME between to_date(:a,'ddmmyyyyhh24miss')                
				and to_date(:b,'ddmmyyyyhh24miss') 
				and evt_code not in('PO')
		`
		rows, err := db.Query(query, ini.Format("02012006150405"), fim.Format("02012006150405"))
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()		

		for rows.Next() {
			u := Evento{}
			err = rows.Scan(
				&u.Id,
				&u.NumeroObjeto,
				&u.DataCriacao,
				&u.DataInsercao,
				&u.Tipo,
				&u.IdUnidadeOrigem,				
				&u.NumeroUnitizador,
				&u.Situacao,
			)
			if err != nil {
				log.Fatal(err)
			}

			if EtiquetaValida(u.NumeroObjeto) {
				eventos = append(eventos, u)
				
			}
			
		}

		
		if len(eventos) > 0 {
						
			fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " unitizadores ", eventos, s)
		    SelecionaObjetos(eventos, s)
		}else{
            fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " não tem unitizadores ", eventos)
		}
		

		if time.Now().Unix()-fim.Unix() > 300 {
			time.Sleep(time.Second * 1)
		} else {
			time.Sleep(time.Second * 300)
		}
		ini = fim
    }

}

func SelecionaObjetos(eventos []Evento,session *r.Session) {
	//ch := make(chan int, 10)
	
    for _, e := range eventos {	
		
		rows, err := r.Table("objeto").GetAllByIndex("numero",e.NumeroObjeto).Run(session)
		if err != nil {
			log.Fatal(err)
		}
		//fmt.Println("rows ", rows)
		var i []Item
		rows.All(&i)
		if len(i) > 0 {
			//fmt.Println("Item ", i)

			for _, i1 := range i {

                if e.DataCriacao.Unix() > i1.DataCriacao.Unix() {
					if e.IdUnidadeOrigem == i1.IdUnidadeDestino {
						i1.Status = "L"
					}else{
						i1.Status = "P"
					}

					fmt.Println("Item a ", i1)

                    err = r.Table("objeto_conferido").
					Insert(i1).
					Exec(session)
					if err != nil {
						log.Fatal(err)
					}

					_, err = r.Table("objeto").
					Get(i1.Id).
					Delete().
					RunWrite(session)
					if err != nil {
						log.Fatal(err)
					}
				}				

			
			}


		}
		

		
	}
	
	
}

