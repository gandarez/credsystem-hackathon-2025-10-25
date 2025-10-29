package util

// Define os serviços válidos em um mapa (ID -> Nome do Serviço).
var ValidServices = map[int]string{
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
	12: "Consulta do Saldo",
	13: "Pagamento de contas",
	14: "Reclamações",
	15: "Atendimento humano",
	16: "Token de proposta",
}

// FindServiceRequest é o corpo da requisição POST /api/find-service
type FindServiceRequest struct {
	Intent string `json:"intent" binding:"required"`
}

// ServiceData é a estrutura de dados retornada para o serviço encontrado
type ServiceData struct {
	ServiceID   int    `json:"service_id"`
	ServiceName string `json:"service_name"`
}

// FindServiceResponse é o corpo da resposta POST /api/find-service
type FindServiceResponse struct {
	Success bool        `json:"success"`
	Data    ServiceData `json:"data"`
	Error   string      `json:"error,omitempty"`
}

// HealthzResponse é o corpo da resposta GET /api/healthz
type HealthzResponse struct {
	Status string `json:"status"`
}

// AIResponse é a estrutura esperada (e forçada) do modelo de IA
type AIResponse struct {
	ServiceID   string `json:"service_id"`
	ServiceName string `json:"service_name"`
}

// JobRequest empacota a intenção e um canal de resposta para a solicitação.
type JobRequest struct {
	Intent       string
	ResponseChan chan FindServiceResponse
}
