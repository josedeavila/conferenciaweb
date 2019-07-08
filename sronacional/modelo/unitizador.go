package modelo

import (
	"fmt"
	"time"
	"log"

	
	r "github.com/dancannon/gorethink"
)

type Unitizador struct {
	Id               string    `gorethink:"id"`
	DataCriacao      time.Time `gorethink:"datacriacao"`
	DataPrazo        time.Time `gorethink:"dataprazo"`
	Evento           string    `gorethink:"evento"`
	IdUnidadeOrigem  string    `gorethink:"idorigem"`
	IdUnidadeDestino string    `gorethink:"iddestino"`
	McmcuUnidadeDestino            string    `gorethink:"mcmcu"`
	CepUnidadeDestino            string    `gorethink:"cep"`
	NomeUnidadeDestino            string    `gorethink:"nome"`
	NumeroServico        string    `gorethink:"numero_servico"`
	NumeroUnitizador     string    `gorethink:"numero_unitizador"`
	PrazoUrg            int       `gorethink:"prazo_urg"`
	PrazoNurg            int       `gorethink:"prazo_nao_urg"`
	QuantidadeItens  int       `gorethink:"quantidade_itens"`
	QuantidadeItensPendentes  int       `gorethink:"quantidade_pendentes"`
	Itens            []Item    `gorethink:"itens"`
	IdSe            string    `gorethink:"idse"`
	SiglaSe            string    `gorethink:"siglase"`
}


//Função para inserção dos itens no unitizador
func (u *Unitizador) InserirItemNoUnitizador(item Item) {
	item.NumeroUnitizador = u.NumeroUnitizador
	u.Itens = append(u.Itens, item)
}

func SelecionaUnitizadores(ini time.Time, s *r.Session) {
	
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

		unitizadores := make([]Unitizador, 0)
		
		query := `SELECT t.tae_code, t.TAE_UNICEPORI, 
					t.TAE_UNICEPDES, t.TAE_RECEPTACLE, 
					t.TAE_CREATETIME, t.TAE_ALERTTYPE, 
					t.TEC_NU_PRAZO_URGENTE, t.TEC_NU_PRAZO_NAO_URGENTE  
			FROM SRO.TECALERTEXPEDITION t 
			WHERE t.TAE_INSERTTIME between to_date(:a,'ddmmyyyyhh24miss')                
			and to_date(:b,'ddmmyyyyhh24miss')
		`
		rows, err := db.Query(query, ini.Format("02012006150405"), fim.Format("02012006150405"))
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		

		for rows.Next() {
			u := Unitizador{}
			err = rows.Scan(
				&u.Id,
				&u.IdUnidadeOrigem,
				&u.IdUnidadeDestino,
				&u.NumeroUnitizador,
				&u.DataCriacao,
				&u.Evento,
				&u.PrazoUrg,
				&u.PrazoNurg,
			)
			if err != nil {
				log.Fatal(err)
			}

			if EtiquetaValida(u.NumeroUnitizador) {
				unitizadores = append(unitizadores, u)
				
			}
			unitizadores = append(unitizadores, u)
			
		}

		
		if len(unitizadores) > 0 {
						
			fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " unitizadores ", unitizadores, s)
			
			InsereUnitizador(unitizadores, s)
		
		}else{
            fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " não tem unitizadores ", unitizadores)
		}
		

		if time.Now().Unix()-fim.Unix() > 300 {
			time.Sleep(time.Second * 1)
		} else {
			time.Sleep(time.Second * 300)
		}
		ini = fim
    }

}

