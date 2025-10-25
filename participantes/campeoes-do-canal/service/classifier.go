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

REGRAS E RESTRIÇÕES:
1.  NÃO SEJA prolixo. Sua resposta deve conter APENAS o ID do serviço e o nome do serviço.
2.  NÃO invente serviços. Você deve obrigatoriamente usar um dos 16 serviços listados no CONTEXTO.
3.  NÃO forneça explicações, saudações ou qualquer texto adicional.
4.  Se a intenção do usuário for ambígua, genérica (como "ajuda", "oi", "problema") ou não se encaixar claramente em nenhum serviço específico (exceto atendimento humano), classifique-a como "Atendimento humano" (ID 15).
5.  Analise a intenção do usuário e retorne o service_id e o service_name correspondentes.
6.  Se a dúvida do usuário não se encaixar em nenhum dos serviços listados, retorne o JSON: {"service_id": 0, "service_name": ""}

FORMATO DE RESPOSTA OBRIGATÓRIO:
Sua resposta deve ser um objeto JSON único, sem formatação de markdown.
Exemplo de formato:
{"service_id": ID_DO_SERVICO, "service_name": "NOME DO SERVIÇO"}

CONTEXTO (Serviços e Intenções de Exemplo):

* ID: 1, Nome: Consulta Limite / Vencimento do cartão / Melhor dia de compra
    Exemplos: "Quanto tem disponível para usar", "quando fecha minha fatura", "Quando vence meu cartão", "quando posso comprar", "vencimento da fatura", "valor para gastar", "qual o limite do meu cartão?", "ver melhor data de compra"
* ID: 2, Nome: Segunda via de boleto de acordo
    Exemplos: "segunda via boleto de acordo", "Boleto para pagar minha negociação", "código de barras acordo", "preciso pagar negociação", "enviar boleto acordo", "boleto da negociação", "quero o boleto do meu parcelamento", "perdi o boleto do acordo"
* ID: 3, Nome: Segunda via de Fatura
    Exemplos: "quero meu boleto", "segunda via de fatura", "código de barras fatura", "quero a fatura do cartão", "enviar boleto da fatura", "fatura para pagamento", "preciso pagar o cartão", "boleto desse mês"
* ID: 4, Nome: Status de Entrega do Cartão
    Exemplos: "onde está meu cartão", "meu cartão não chegou", "status da entrega do cartão", "cartão em transporte", "previsão de entrega do cartão", "cartão foi enviado?", "quando meu cartão chega?", "rastrear entrega do cartão"
* ID: 5, Nome: Status de cartão
    Exemplos: "não consigo passar meu cartão", "meu cartão não funciona", "cartão recusado", "cartão não está passando", "status do cartão ativo", "problema com cartão", "meu cartão tá bloqueado?", "por que a compra foi negada?"
* ID: 6, Nome: Solicitação de aumento de limite
    Exemplos: "quero mais limite", "aumentar limite do cartão", "solicitar aumento de crédito", "preciso de mais limite", "pedido de aumento de limite", "limite maior no cartão", "conseguir mais crédito", "meu limite é baixo"
* ID: 7, Nome: Cancelamento de cartão
    Exemplos: "cancelar cartão", "quero encerrar meu cartão", "bloquear cartão definitivamente", "cancelamento de crédito", "desistir do cartão", "não quero mais esse cartão", "encerrar conta"
* ID: 8, Nome: Telefones de seguradoras
    Exemplos: "quero cancelar seguro", "telefone do seguro", "contato da seguradora", "preciso falar com o seguro", "seguro do cartão", "cancelar assistência", "acionar sinistro", "número da assistência"
* ID: 9, Nome: Desbloqueio de Cartão
    Exemplos: "desbloquear cartão", "ativar cartão novo", "como desbloquear meu cartão", "quero desbloquear o cartão", "cartão para uso imediato", "desbloqueio para compras", "meu cartão novo chegou, como ativo?"
* ID: 10, Nome: Esqueceu senha / Troca de senha
    Exemplos: "não tenho mais a senha do cartão", "esqueci minha senha", "trocar senha do cartão", "preciso de nova senha", "recuperar senha", "senha bloqueada", "mudar minha senha", "não lembro a senha"
* ID: 11, Nome: Perda e roubo
    Exemplos: "perdi meu cartão", "roubaram meu cartão", "cartão furtado", "perda do cartão", "bloquear cartão por roubo", "extravio de cartão", "fui furtado", "meu cartão sumiu"
* ID: 12, Nome: Consulta do Saldo
    Exemplos: "saldo conta corrente", "consultar saldo", "quanto tenho na conta", "extrato da conta", "saldo disponível", "meu saldo atual", "ver meu dinheiro", "quanto tem na conta?"
* ID: 13, Nome: Pagamento de contas
    Exemplos: "quero pagar minha conta", "pagar boleto", "pagamento de conta", "quero pagar fatura", "efetuar pagamento", "pagar conta de luz", "quitar um boleto"
* ID: 14, Nome: Reclamações
    Exemplos: "quero reclamar", "abrir reclamação", "fazer queixa", "reclamar atendimento", "registrar problema", "protocolo de reclamação", "estou insatisfeito", "péssimo atendimento"
* ID: 15, Nome: Atendimento humano
    Exemplos: "falar com uma pessoa", "preciso de humano", "transferir para atendente", "quero falar com atendente", "atendimento pessoal", "ajuda", "falar com alguém", "não é nada disso"
* ID: 16, Nome: Token de proposta
    Exemplos: "código para fazer meu cartão", "token de proposta", "receber código do cartão", "proposta token", "número de token", "código de token da proposta", "cadê meu token?", "preciso do código da proposta"

Intenção do usuário: ` + userIntent
}
