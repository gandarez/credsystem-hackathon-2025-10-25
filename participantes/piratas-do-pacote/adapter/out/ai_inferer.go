package out

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/piratas-do-pacote/global/textnorm"
)

var ServiceByID = map[int]string{
	1:  "Consulta Limite / Vencimento do cartão / Melhor dia de compra",
	2:  "Segunda via de boleto de acordo",
	3:  "Segunda via de Fatura",
	4:  "Status de Entrega do Cartão",
	5:  "Status de cartão",
	6:  "Solicitação de aumento de limite",
	7:  "Cancelamento de cartão",
	8:  "Telefones de seguradoras",
	9:  "Desbloqueio de Cartão",
	10: "Esqueceu senha / Troca de senha",
	11: "Perda e roubo",
	12: "Consulta do Saldo Conta do Mais",
	13: "Pagamento de contas",
	14: "Reclamações",
	15: "Atendimento humano",
	16: "Token de proposta",
}

type KB struct {
	Version             string      `json:"version"`
	Domain              string      `json:"domain"`
	InstructionReminder string      `json:"instruction_reminder"`
	Services            []KBService `json:"services"`
}
type KBService struct {
	ID                  int      `json:"id"`
	Name                string   `json:"name"`
	Keywords            []string `json:"keywords"`
	PositiveExamples    []string `json:"positive_examples"`
	NegativeExamples    []string `json:"negative_examples"`
	DisambiguationNotes string   `json:"disambiguation_notes"`
}

func loadKB(path string) (*KB, error) {
	if path == "" {
		return nil, nil
	}
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ler kb.json: %w", err)
	}
	var kb KB
	if err := json.Unmarshal(f, &kb); err != nil {
	}
	return &kb, nil
}

func joinTop(ss []string, n int) string {
	if n > len(ss) {
		n = len(ss)
	}
	return strings.Join(ss[:n], " | ")
}
func trunc(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}

func (i *AiInferer) InferService(ctx context.Context, intent string) (int, error) {
	prompt := i.preparePrompt(intent)
	return i.request(ctx, prompt)
}

// ===== Prompt =====
func buildSystemPrompt(kb *KB) string {
	var b strings.Builder
	b.WriteString("Você é um classificador determinístico de intenções.\n")
	b.WriteString("Tarefa: Receba um texto do usuário e escolha O ÚNICO serviço que melhor corresponde, retornando APENAS o número do ID do serviço (um número inteiro) e nada mais.\n")
	b.WriteString("NUNCA invente serviços, NUNCA invente IDs e NUNCA escreva nomes. Apenas o número do ID.\n")
	b.WriteString("Lista fixa de serviços válidos (ID: Nome):\n")
	for id, name := range ServiceByID {
		fmt.Fprintf(&b, "%d: %s\n", id, name)
	}
	// (Opcional) lembrete da instrução vinda do KB
	if kb != nil && strings.TrimSpace(kb.InstructionReminder) != "" {
		b.WriteString("\nInstrução adicional (lembrete): ")
		b.WriteString(kb.InstructionReminder)
		b.WriteString("\n")
	}

	b.WriteString("- Se a entrada mencionar 'token' ou 'código' associado a 'proposta', 'fazer meu cartão', 'solicitação', 'cadastro' ou 'abertura', classifique como 16 (Token de proposta).\n")
	b.WriteString("- Use 9 (Desbloqueio) apenas para 'desbloquear/ativar cartão novo/primeiro uso', sem menção a token/código de proposta.\n")
	b.WriteString("- Se mencionar 'senha'/'reset de senha' (app/conta), classifique como 10 (Esqueceu/Troca de senha), não 16.\n")
	b.WriteString("- Se mencionar 'fatura' e NÃO mencionar 'acordo/negociação', classifique como 3 (Segunda via de Fatura), mesmo que o usuário fale 'pagar/pagamento'.\n")
	b.WriteString("- Se mencionar 'acordo/negociação', classifique como 2 (Segunda via de boleto de acordo), mesmo que o usuário fale 'pagar/pagamento'.\n")
	b.WriteString("- 'Pagamento de contas' (13) só quando for boleto/conta genérico sem 'fatura' nem 'acordo'.\n")
	b.WriteString("- Se o usuário disser 'meu boleto', 'boleto para pagamento', ou 'segunda via de boleto' SEM mencionar 'acordo/negociação', classifique como 3 (Segunda via de Fatura).\n")
	b.WriteString("- Se mencionar 'acordo/negociação/renegociação', classifique como 2 (Segunda via de boleto de acordo), mesmo que peça 'pagar' ou 'boleto'.\n")
	b.WriteString("- Use 13 (Pagamento de contas) apenas quando for boleto/conta genérico, sem 'fatura' e sem 'acordo', e NÃO houver indícios de que é a fatura do cartão.\n")

	// Base de conhecimento compacta (para reduzir alucinação).
	// Mantemos o formato simples, linha a linha: [ID] kws=... | pos=... | neg=... | note=...
	if kb != nil && len(kb.Services) > 0 {
		b.WriteString("\nBase de conhecimento (compacta; usar apenas como referência de desambiguação):\n")
		for _, s := range kb.Services {
			// Proteção: só aceita IDs válidos 1..16
			if _, ok := ServiceByID[s.ID]; !ok {
				continue
			}
			kws := joinTop(s.Keywords, 5)
			pos := joinTop(s.PositiveExamples, 3)
			neg := joinTop(s.NegativeExamples, 2)
			note := trunc(s.DisambiguationNotes, 160)
			fmt.Fprintf(&b, "[%d] kws=%s | pos=%s | neg=%s | note=%s\n",
				s.ID, kws, pos, neg, note)
		}
	}

	b.WriteString("\nRestrições:\n- Saída deve ser SÓ o número do ID (ex.: '4').\n- Temperature = 0.\n")
	return b.String()
}

func buildUserPrompt(intent string) string {
	return fmt.Sprintf("Entrada do usuário: %q\nRetorne apenas o ID (um inteiro).", intent)
}

func (i *AiInferer) preparePrompt(intent string) string {
	normIntent := textnorm.Normalize(intent, textnorm.DefaultOptions())
	return buildUserPrompt(normIntent)
}

func (i *AiInferer) request(ctx context.Context, prompt string) (int, error) {
	var result ChatResponse

	reqBody := ChatRequest{
		Model: defaultAiModel,
		Messages: []ChatMessage{
			{Role: systemRole, Content: buildSystemPrompt(i.kb)},
			{Role: userRole, Content: prompt},
		},
		Temperature: 0.0,                 // determinístico
		TopP:        1.0,                 // não restringe o espaço de amostragem
		MaxTokens:   4,                   // cabe um inteiro
		Stop:        []string{"\n", " "}, // corta qualquer extra após o dígito
		// Penalties em 0 para não “forçar” variação
		FrequencyPenalty: 0,
		PresencePenalty:  0,
	}

	_, err := i.client.PostJSON(ctx, openRouterDefaultPath, reqBody, &result)
	if err != nil {
		return 0, err
	}

	finalResponse, err := strconv.ParseInt(result.Choices[0].Message.Content, 10, 0)
	if err != nil {
		return 0, err
	}

	return int(finalResponse), nil

}
