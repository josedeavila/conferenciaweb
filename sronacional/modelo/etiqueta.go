package modelo

import (
	"regexp"
)

var regexEtiqueta *regexp.Regexp

func init() {
	regexEtiqueta = regexp.MustCompile(`[A-Z]{2}[0-9]{9}[A-Z]{2}`)
}

func EtiquetaValida(numero string) bool {
	return regexEtiqueta.MatchString(numero)
}