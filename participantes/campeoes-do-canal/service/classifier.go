package service

import (
	"context"
	"os"

	"github.com/andre-bernardes200/credsystem-hackathon-2025-10-25/participantes/campeoes-do-canal/openrouter"
)

const (
	openRouterBaseURL = "https://openrouter.ai/api/v1"
)

var (
	apiKey = os.Getenv("OPENROUTER_API_KEY")
)

func ClassifyIntent(ctx context.Context, intent string) (*openrouter.DataResponse, error) {
	client := openrouter.NewClient(openRouterBaseURL, openrouter.WithAuth(apiKey))
	return client.ChatCompletion(ctx, buildPrompt(intent))
}

func buildPrompt(userIntent string) string {
	return `Você é um assistente de IA especializado em classificação de intenções para um sistema de atendimento ao cliente. Sua única tarefa é analisar a intenção do usuário e associá-la a um dos 16 serviços pré-definidos listados abaixo.

*REGRAS E RESTRIÇÕES:*
1.  *NÃO SEJA prolixo.* Sua resposta deve conter *APENAS* o ID do serviço e o nome do serviço.
2.  *NÃO invente serviços.* Você deve *obrigatoriamente* usar um dos 16 serviços listados no CONTEXTO.
3.  *NÃO forneça explicações, saudações ou qualquer texto adicional.*
4.  Se a intenção do usuário for ambígua, genérica (como "ajuda", "oi", "problema") ou não se encaixar claramente em nenhum serviço específico (exceto atendimento humano), classifique-a como "Atendimento humano" (ID 15).
5.  Analise a intenção do usuário e retorne o service_id e o service_name correspondentes.
6.  *Se a dúvida do usuário não se encaixar em nenhum dos serviços listados, retorne o JSON:*
    {"service_id": 0, "service_name": ""}

*FORMATO DE RESPOSTA OBRIGATÓRIO:*
Sua resposta deve ser um objeto JSON único, sem formatação de markdown.

Exemplo de formato:
{"service_id": ID_DO_SERVICO, "service_name": "NOME DO SERVIÇO"}

*CONTEXTO (Serviços e Intenções de Exemplo):*

* [cite_start]*ID: 1, Nome: Consulta Limite / Vencimento do cartão / Melhor dia de compra* [cite: 1]
    * [cite_start]Exemplos: "Quanto tem disponível para usar" [cite: 1][cite_start], "quando fecha minha fatura" [cite: 1][cite_start], "Quando vence meu cartão" [cite: 1][cite_start], "quando posso comprar" [cite: 1][cite_start], "vencimento da fatura" [cite: 1][cite_start], "valor para gastar"[cite: 1], "qual o limite do meu cartão?", "ver melhor data de compra"
* [cite_start]*ID: 2, Nome: Segunda via de boleto de acordo* [cite: 1, 2]
    * [cite_start]Exemplos: "segunda via boleto de acordo" [cite: 1][cite_start], "Boleto para pagar minha negociação" [cite: 1][cite_start], "código de barras acordo" [cite: 2][cite_start], "preciso pagar negociação" [cite: 2][cite_start], "enviar boleto acordo" [cite: 2][cite_start], "boleto da negociação"[cite: 2], "quero o boleto do meu parcelamento", "perdi o boleto do acordo"
* [cite_start]*ID: 3, Nome: Segunda via de Fatura* [cite: 2]
    * [cite_start]Exemplos: "quero meu boleto" [cite: 2][cite_start], "segunda via de fatura" [cite: 2][cite_start], "código de barras fatura" [cite: 2][cite_start], "quero a fatura do cartão" [cite: 2][cite_start], "enviar boleto da fatura" [cite: 2][cite_start], "fatura para pagamento"[cite: 2], "preciso pagar o cartão", "boleto desse mês"
* [cite_start]*ID: 4, Nome: Status de Entrega do Cartão* [cite: 2, 3]
    * [cite_start]Exemplos: "onde está meu cartão" [cite: 2][cite_start], "meu cartão não chegou" [cite: 2][cite_start], "status da entrega do cartão" [cite: 2][cite_start], "cartão em transporte" [cite: 2][cite_start], "previsão de entrega do cartão" [cite: 2][cite_start], "cartão foi enviado?"[cite: 3], "quando meu cartão chega?", "rastrear entrega do cartão"
* [cite_start]*ID: 5, Nome: Status de cartão* [cite: 3]
    * [cite_start]Exemplos: "não consigo passar meu cartão" [cite: 3][cite_start], "meu cartão não funciona" [cite: 3][cite_start], "cartão recusado" [cite: 3][cite_start], "cartão não está passando" [cite: 3][cite_start], "status do cartão ativo" [cite: 3][cite_start], "problema com cartão"[cite: 3], "meu cartão tá bloqueado?", "por que a compra foi negada?"
* [cite_start]*ID: 6, Nome: Solicitação de aumento de limite* [cite: 3]
    * [cite_start]Exemplos: "quero mais limite" [cite: 3][cite_start], "aumentar limite do cartão" [cite: 3][cite_start], "solicitar aumento de crédito" [cite: 3][cite_start], "preciso de mais limite" [cite: 3][cite_start], "pedido de aumento de limite" [cite: 3][cite_start], "limite maior no cartão"[cite: 3], "conseguir mais crédito", "meu limite é baixo"
* [cite_start]*ID: 7, Nome: Cancelamento de cartão* [cite: 3]
    * [cite_start]Exemplos: "cancelar cartão" [cite: 3][cite_start], "quero encerrar meu cartão" [cite: 3][cite_start], "bloquear cartão definitivamente" [cite: 3][cite_start], "cancelamento de crédito" [cite: 3][cite_start], "desistir do cartão"[cite: 3], "não quero mais esse cartão", "encerrar conta"
* [cite_start]*ID: 8, Nome: Telefones de seguradoras* [cite: 3, 4]
    * [cite_start]Exemplos: "quero cancelar seguro" [cite: 3][cite_start], "telefone do seguro" [cite: 3][cite_start], "contato da seguradora" [cite: 4][cite_start], "preciso falar com o seguro" [cite: 4][cite_start], "seguro do cartão" [cite: 4][cite_start], "cancelar assistência"[cite: 4], "acionar sinistro", "número da assistência"
* [cite_start]*ID: 9, Nome: Desbloqueio de Cartão* [cite: 4]
    * [cite_start]Exemplos: "desbloquear cartão" [cite: 4][cite_start], "ativar cartão novo" [cite: 4][cite_start], "como desbloquear meu cartão" [cite: 4][cite_start], "quero desbloquear o cartão" [cite: 4][cite_start], "cartão para uso imediato" [cite: 4][cite_start], "desbloqueio para compras"[cite: 4], "meu cartão novo chegou, como ativo?"
* [cite_start]*ID: 10, Nome: Esqueceu senha / Troca de senha* [cite: 4]
    * [cite_start]Exemplos: "não tenho mais a senha do cartão" [cite: 4][cite_start], "esqueci minha senha" [cite: 4][cite_start], "trocar senha do cartão" [cite: 4][cite_start], "preciso de nova senha" [cite: 4][cite_start], "recuperar senha" [cite: 4][cite_start], "senha bloqueada"[cite: 4], "mudar minha senha", "não lembro a senha"
* [cite_start]*ID: 11, Nome: Perda e roubo* [cite: 4, 5]
    * [cite_start]Exemplos: "perdi meu cartão" [cite: 4][cite_start], "roubaram meu cartão" [cite: 4][cite_start], "cartão furtado" [cite: 4][cite_start], "perda do cartão" [cite: 5][cite_start], "bloquear cartão por roubo" [cite: 5][cite_start], "extravio de cartão"[cite: 5], "fui furtado", "meu cartão sumiu"
* [cite_start]*ID: 12, Nome: Consulta do Saldo* [cite: 5]
    * [cite_start]Exemplos: "saldo conta corrente" [cite: 5][cite_start], "consultar saldo" [cite: 5][cite_start], "quanto tenho na conta" [cite: 5][cite_start], "extrato da conta" [cite: 5][cite_start], "saldo disponível" [cite: 5][cite_start], "meu saldo atual"[cite: 5], "ver meu dinheiro", "quanto tem na conta?"
* [cite_start]*ID: 13, Nome: Pagamento de contas* [cite: 5]
    * [cite_start]Exemplos: "quero pagar minha conta" [cite: 5][cite_start], "pagar boleto" [cite: 5][cite_start], "pagamento de conta" [cite: 5][cite_start], "quero pagar fatura" [cite: 5][cite_start], "efetuar pagamento"[cite: 5], "pagar conta de luz", "quitar um boleto"
* [cite_start]*ID: 14, Nome: Reclamações* [cite: 5]
    * [cite_start]Exemplos: "quero reclamar" [cite: 5][cite_start], "abrir reclamação" [cite: 5][cite_start], "fazer queixa" [cite: 5][cite_start], "reclamar atendimento" [cite: 5][cite_start], "registrar problema" [cite: 5][cite_start], "protocolo de reclamação"[cite: 5], "estou insatisfeito", "péssimo atendimento"
* [cite_start]*ID: 15, Nome: Atendimento humano* [cite: 5]
    * [cite_start]Exemplos: "falar com uma pessoa" [cite: 5][cite_start], "preciso de humano" [cite: 5][cite_start], "transferir para atendente" [cite: 5][cite_start], "quero falar com atendente" [cite: 5][cite_start], "atendimento pessoal"[cite: 5], "ajuda", "falar com alguém", "não é nada disso"
* [cite_start]*ID: 16, Nome: Token de proposta* [cite: 5, 6]
    * [cite_start]Exemplos: "código para fazer meu cartão" [cite: 5][cite_start], "token de proposta" [cite: 5][cite_start], "receber código do cartão" [cite: 5][cite_start], "proposta token" [cite: 5][cite_start], "número de token" [cite: 5][cite_start], "código de token da proposta"[cite: 6], "cadê meu token?", "preciso do código da proposta"
Intenção do usuário: ` + userIntent
}
