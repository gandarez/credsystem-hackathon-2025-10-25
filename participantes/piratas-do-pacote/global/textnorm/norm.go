package textnorm

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// NormalizeOptions permite customizar mapeamentos e ruídos.
type NormalizeOptions struct {
	Replacements map[string]string

	Synonyms map[string]string

	Noise map[string]bool
}

func DefaultOptions() NormalizeOptions {
	return NormalizeOptions{
		// Replacements: após lower+remover acentos, padroniza grafias/abreviações.
		Replacements: map[string]string{
			// "segunda via"
			"2a via": "segunda via",
			"2 via":  "segunda via",
			"2ª via": "segunda via", // será removido acento antes, mas não custa manter
			// "codigo de barras"
			"cod de barras": "codigo de barras",
			"cod barras":    "codigo de barras",
			"codbarras":     "codigo de barras",
			// variações úteis
			"cartao de credito": "cartao",
			"boleto bancario":   "boleto",
			"boleto da fatura":  "boleto fatura",
			"fatura do cartao":  "fatura",
			// termos próximos que queremos unificar
			"melhor dia para comprar": "melhor dia de compra",
			"melhor dia compra":       "melhor dia de compra",
		},

		// Synonyms: mapeia frases/palavras do domínio para formas canônicas (usa bordas de palavra).
		Synonyms: map[string]string{
			// ===== Consulta limite / vencimento / melhor dia de compra (id 1)
			"quanto tem disponivel para usar": "consulta limite",
			"valor para gastar":               "consulta limite",
			"quando fecha minha fatura":       "consulta vencimento",
			"quando vence meu cartao":         "consulta vencimento",
			"vencimento da fatura":            "consulta vencimento",
			"quando posso comprar":            "melhor dia de compra",
			"melhor dia de compra":            "melhor dia de compra",

			// ===== Segunda via de boleto de acordo (id 2)
			"segunda via boleto de acordo":       "boleto acordo",
			"boleto para pagar minha negociacao": "boleto acordo",
			"codigo de barras acordo":            "boleto acordo",
			"preciso pagar negociacao":           "boleto acordo",
			"enviar boleto acordo":               "boleto acordo",
			"boleto da negociacao":               "boleto acordo",

			// ===== Segunda via de fatura (id 3)
			"quero meu boleto":         "segunda via fatura",
			"segunda via de fatura":    "segunda via fatura",
			"codigo de barras fatura":  "segunda via fatura",
			"quero a fatura do cartao": "segunda via fatura",
			"enviar boleto da fatura":  "segunda via fatura",
			"fatura para pagamento":    "segunda via fatura",

			// ===== Solicitacao de cartao (id 4)
			"pedir cartao":              "solicitar cartao",
			"quero um cartao":           "solicitar cartao",
			"solicitar cartao":          "solicitar cartao",
			"cartao novo":               "solicitar cartao",
			"como faco para ter cartao": "solicitar cartao",
			"pedido de cartao":          "solicitar cartao",

			// ===== Aumento de limite (id 5)
			"aumentar limite":             "aumento de limite",
			"limite maior":                "aumento de limite",
			"elevar limite":               "aumento de limite",
			"reajuste de limite":          "aumento de limite",
			"solicitar aumento de limite": "aumento de limite",
			"pedir mais limite":           "aumento de limite",

			// ===== Parcelamento de fatura (id 6)
			"parcelar fatura":                  "parcelar fatura",
			"dividir fatura":                   "parcelar fatura",
			"parcelamento da fatura do cartao": "parcelar fatura",
			"quero parcelar a fatura":          "parcelar fatura",
			"parcela fatura":                   "parcelar fatura",
			"opcoes de parcelamento":           "parcelar fatura",

			// ===== Cancelamento de cartao (id 7)
			"cancelar cartao":               "cancelar cartao",
			"encerrar cartao":               "cancelar cartao",
			"bloquear definitivamente":      "cancelar cartao",
			"cancelamento do cartao":        "cancelar cartao",
			"cartao perdido quero cancelar": "cancelar cartao",
			"encerrar":                      "cancelar",
			"cancelamento":                  "cancelar",

			// ===== Bloqueio temporario (id 8)
			"bloquear cartao temporariamente": "bloquear temporario",
			"bloqueio temporario":             "bloquear temporario",
			"pausar cartao":                   "bloquear temporario",
			"bloquear por um tempo":           "bloquear temporario",
			"suspender uso do cartao":         "bloquear temporario",
			"bloqueio momentaneo":             "bloquear temporario",

			// ===== Desbloqueio de cartao (id 9)
			"desbloquear cartao":              "desbloquear cartao",
			"ativar cartao novo":              "desbloquear cartao",
			"cartao chegou quero desbloquear": "desbloquear cartao",
			"desbloqueio do cartao":           "desbloquear cartao",
			"liberar cartao":                  "desbloquear cartao",
			"ativacao de cartao":              "desbloqueio",

			// ===== Alteracao de limite (id 10)
			"alterar limite":    "alterar limite",
			"mudar limite":      "alterar limite",
			"ajustar limite":    "alterar limite",
			"reduzir limite":    "alterar limite",
			"diminuir limite":   "alterar limite",
			"configurar limite": "alterar limite",

			// ===== Contestacao de compra (id 11)
			"nao reconheco compra":      "contestacao de compra",
			"compra indevida":           "contestacao de compra",
			"contestar compra":          "contestacao de compra",
			"lancamento desconhecido":   "contestacao de compra",
			"transacao nao reconhecida": "contestacao de compra",
			"disputa de compra":         "contestacao de compra",

			// ===== Negociacao de divida (id 12)
			"negociar divida":                "negociacao de divida",
			"acordo de pagamento":            "negociacao de divida",
			"quitar divida com desconto":     "negociacao de divida",
			"proposta de negociacao":         "negociacao de divida",
			"negociacao de fatura em atraso": "negociacao de divida",
			"acordo":                         "negociacao de divida",

			// ===== Atualizacao cadastral (id 13)
			"trocar endereco":         "atualizacao cadastral",
			"alterar telefone":        "atualizacao cadastral",
			"atualizar cadastro":      "atualizacao cadastral",
			"mudar dados cadastrais":  "atualizacao cadastral",
			"atualizacao de cadastro": "atualizacao cadastral",
			"corrigir endereco":       "atualizacao cadastral",

			// ===== Reclamacoes (id 14)
			"quero reclamar":           "reclamacao",
			"reclamacao":               "reclamacao",
			"abrir reclamacao":         "reclamacao",
			"problema com atendimento": "reclamacao",
			"registrar reclamacao":     "reclamacao",
			"protocolo de reclamacao":  "reclamacao",

			// ===== Atendimento humano (id 15)
			"falar com uma pessoa":      "atendimento humano",
			"preciso de humano":         "atendimento humano",
			"transferir para atendente": "atendimento humano",
			"quero falar com atendente": "atendimento humano",
			"atendimento pessoal":       "atendimento humano",

			// ===== Token de proposta (id 16)
			"codigo para fazer meu cartao": "token de proposta",
			"token de proposta":            "token de proposta",
			"receber codigo do cartao":     "token de proposta",
			"proposta token":               "token de proposta",
			"numero de token":              "token de proposta",
			"codigo de token da proposta":  "token de proposta",
		},

		// Noise: marcas de cortesia/ruído que não carregam intenção.
		Noise: map[string]bool{
			"por favor":     true,
			"por gentileza": true,
			"pfv":           true,
			"urgente":       true,
			"obrigado":      true,
			"obrigada":      true,
			"bom dia":       true,
			"boa tarde":     true,
			"boa noite":     true,
		},
	}
}

func Normalize(input string, opt NormalizeOptions) string {
	if strings.TrimSpace(input) == "" {
		return ""
	}

	s := strings.ToLower(input)

	s = stripAccents(s)

	s = applyReplacements(s, opt.Replacements)

	s = applySynonymsWordBound(s, opt.Synonyms)

	s = removeNoiseWordBound(s, opt.Noise)

	s = normalizeSpaces(s)

	s = strings.TrimSpace(s)

	return s
}

// ---- helpers ----

func stripAccents(s string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(isNonSpacingMark), norm.NFC)
	res, _, _ := transform.String(t, s)
	return res
}

func isNonSpacingMark(r rune) bool {
	return unicode.Is(unicode.Mn, r)
}

func applyReplacements(s string, repl map[string]string) string {
	if len(repl) == 0 {
		return s
	}
	for from, to := range repl {
		s = strings.ReplaceAll(s, from, to)
	}
	return s
}

func applySynonymsWordBound(s string, syn map[string]string) string {
	if len(syn) == 0 {
		return s
	}
	for from, to := range syn {
		p := regexp.MustCompile(`\b` + regexp.QuoteMeta(from) + `\b`)
		s = p.ReplaceAllString(s, to)
	}
	return s
}

func removeNoiseWordBound(s string, noise map[string]bool) string {
	if len(noise) == 0 {
		return s
	}
	for token := range noise {
		p := regexp.MustCompile(`\b` + regexp.QuoteMeta(token) + `\b`)
		s = strings.TrimSpace(p.ReplaceAllString(s, " "))
	}
	return s
}

var multiSpace = regexp.MustCompile(`\s+`)

func normalizeSpaces(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = multiSpace.ReplaceAllString(s, " ")
	return s
}
