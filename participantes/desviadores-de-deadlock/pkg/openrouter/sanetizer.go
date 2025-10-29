package openrouter

import (
	// ...
	"regexp"
	"strings"
)

var (
	codeFenceRe = regexp.MustCompile("(?s)```(?:json)?\\s*(\\{.*?\\})\\s*```")
	jsonObjRe   = regexp.MustCompile(`(?s)\{.*\}`)
)

func extractJSON(s string) string {
	// 1) tenta bloco cercado
	if m := codeFenceRe.FindStringSubmatch(s); len(m) == 2 {
		return m[1]
	}
	// 2) tenta primeiro objeto JSON "cru" no texto
	if m := jsonObjRe.FindString(s); m != "" {
		return m
	}
	// 3) fallback: trim cercas simples ou retorno direto
	return strings.TrimSpace(s)
}
