package core

import (
	"context"
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

func (i *InferenceUseCase) Infer(ctx context.Context, input InferenceInput) InferenceResult {

	inference, err := i.inferer.InferService(ctx, input.Intent)
	if err != nil {
		return InferenceResult{
			Success: false,
			Error:   err.Error(),
		}
	}

	svcName, exists := ServiceByID[inference]
	if inference == notFoundId || !exists {
		return InferenceResult{
			Success: false,
			Error:   notFoundMessage,
		}
	}

	return InferenceResult{
		Success: true,
		Data: InferenceData{
			ServiceId:   inference,
			ServiceName: svcName,
		},
	}
}
