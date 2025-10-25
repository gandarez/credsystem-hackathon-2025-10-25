package validator

// IsValidService verifica se o service_id está entre os 16 serviços válidos
func IsValidService(serviceID int) bool {
	return serviceID >= 1 && serviceID <= 16
}

// GetServiceName retorna o nome do serviço baseado no ID
func GetServiceName(serviceID int) string {
	services := map[int]string{
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
	
	if name, ok := services[serviceID]; ok {
		return name
	}
	return ""
}

// ValidateResponse verifica se a resposta do agente é válida
func ValidateResponse(serviceID int, serviceName string) bool {
	if !IsValidService(serviceID) {
		return false
	}
	
	expectedName := GetServiceName(serviceID)
	return serviceName == expectedName
}
