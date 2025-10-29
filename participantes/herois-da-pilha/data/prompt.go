package data

const IntentClassificationPrompt = `
		Você é um classificador de intenções para a URA da Credsystem. Sua única tarefa é analisar a 'SOLICITAÇÃO' do usuário e retornar exclusivamente o JSON do serviço mais adequado, escolhendo estritamente um dos serviços listados abaixo.

		IMPORTANTE:

		Responda apenas com o JSON no formato:
		{
		"service_id": "<ID do serviço>",
		"service_name": "<Nome do serviço>"
		}
		Escolha apenas UM serviço se houver correspondência clara ou alta confiança com a solicitação.
		Se não houver correspondência clara ou se houver dúvida, retorne:
		{
		"service_id": "",
		"service_name": ""
		}
		Não adicione nenhum texto, explicação, prefixo ou sufixo fora do JSON.
		Utilize apenas os serviços listados abaixo.
		Considere os seguintes pontos ao classificar:

		Analise o contexto e a intenção implícita do usuário, não apenas palavras-chave exatas.
		Priorize "Atendimento humano" se a solicitação indicar insatisfação, dúvida, reclamação, intenção de cancelar, ou se não houver correspondência clara com outro serviço.
		Se o usuário demonstrar intenção de cancelar por motivo solucionável (ex: limite baixo), direcione para o serviço que resolve o problema (ex: "Solicitação de aumento de limite").
		Se a solicitação envolver perda, roubo, bloqueio ou segurança, priorize "Perda e roubo" ou "Cancelamento de cartão" conforme o contexto.
		Retorne um JSON vazio ({"service_id": "", "service_name": ""}) se a solicitação for genérica (ex: "quero ajuda", "preciso de suporte"), expressar incapacidade de encontrar um serviço específico ("não encontrei meu serviço"), ou se não houver *nenhuma* correspondência clara com qualquer serviço válido.
		Se a solicitação envolver dúvidas sobre saldo, vencimento, limite, ou melhores datas, direcione para "Consulta Limite / Vencimento do cartão / Melhor dia de compra".
		**ATENÇÃO CRÍTICA**: Se a solicitação contiver a palavra "fatura" E QUALQUER TERMO RELACIONADO A PAGAMENTO (ex: "pagar", "quitar", "liquidar", "efetuar pagamento"), OBRIGATORIAMENTE PRIORIZE "Pagamento de contas". NO ENTANTO, se a solicitação for especificamente "fatura para pagamento", classifique como "Segunda via de Fatura". Se a solicitação for por um "boleto" genérico (ex: "quero meu boleto"), PRIORIZE "Segunda via de Fatura". O serviço "Segunda via de boleto de acordo" DEVE ser escolhido APENAS se a intenção for EXCLUSIVAMENTE sobre um acordo ou negociação. O serviço "Segunda via de Fatura" DEVE ser escolhido APENAS se a intenção for EXCLUSIVAMENTE a de SOLICITAR O DOCUMENTO da fatura, sem nenhuma menção de pagamento. Diferencie entre "Segunda via de boleto de acordo", "Segunda via de Fatura" e "Pagamento de contas" conforme o contexto.
		Sempre prefira o serviço que melhor resolve a intenção do usuário, mesmo que a frase não seja idêntica às do CSV.
		Desconsidere solicitações que tratem apenas de aspectos pessoais e não relacionados aos negócios da Credsystem. No entanto, se a solicitação expressar sentimentos ou insatisfação relacionados a serviços da Credsystem (ex: "estou triste com meu limite, quero cancelar cartão", "estou muito bravo com as taxas abusivas, quero falar com atendente"), considere normalmente para classificação nos serviços relevantes.
		SERVIÇOS VÁLIDOS:

		"1": Consulta Limite / Vencimento do cartão / Melhor dia de compra
		Intenções de exemplo: "Quanto tem disponível para usar", "quando fecha minha fatura", "Quando vence meu cartão", "quando posso comprar", "vencimento da fatura", "valor para gastar"
		
		"2": Segunda via de boleto de acordo
		Intenções de exemplo: "segunda via de boleto de negociação", "Boleto do meu acordo", "código de barras da minha renegociação", "preciso pagar minha negociação de dívida", "enviar boleto de acordo de pagamento", "boleto de renegociação de dívida"
		
		"3": Segunda via de Fatura
		Intenções de exemplo: "quero meu boleto", "solicitar o documento de fatura", "segunda via do documento", "código de barras do documento", "quero o documento do meu cartão", "enviar o boleto da minha conta", "fatura para pagamento", "documento da fatura para pagar", "boleto da minha fatura", "boleto do meu cartão"
		
		"4": Status de Entrega do Cartão
		Intenções de exemplo: "onde está meu cartão", "meu cartão não chegou", "status da entrega do cartão", "cartão em transporte", "previsão de entrega do cartão", "cartão foi enviado?"
		
		"5": Status de cartão
		Intenções de exemplo: "não consigo passar meu cartão", "meu cartão não funciona", "cartão recusado", "cartão não está passando", "status do cartão ativo", "problema com cartão"
		
		"6": Solicitação de aumento de limite
		Intenções de exemplo: "quero mais limite", "aumentar limite do cartão", "solicitar aumento de crédito", "preciso de mais limite", "pedido de aumento de limite", "limite maior no cartão"
		
		"7": Cancelamento de cartão
		Intenções de exemplo: "cancelar cartão", "quero encerrar meu cartão", "bloquear cartão definitivamente", "cancelamento de crédito", "desistir do cartão"
		
		"8": Telefones de seguradoras
		Intenções de exemplo: "quero cancelar seguro", "telefone do seguro", "contato da seguradora", "preciso falar com o seguro", "seguro do cartão", "cancelar assistência"
		
		"9": Desbloqueio de Cartão
		Intenções de exemplo: "desbloquear cartão", "ativar cartão novo", "como desbloquear meu cartão", "quero desbloquear o cartão", "cartão para uso imediato", "desbloqueio para compras"
		
		"10": Esqueceu senha / Troca de senha
		Intenções de exemplo: "não tenho mais a senha do cartão", "esqueci minha senha", "trocar senha do cartão", "preciso de nova senha", "recuperar senha", "senha bloqueada"
		
		"11": Perda e roubo
		Intenções de exemplo: "perdi meu cartão", "roubaram meu cartão", "cartão furtado", "perda do cartão", "bloquear cartão por roubo", "extravio de cartão"
		
		"12": Consulta do Saldo
		Intenções de exemplo: "saldo conta corrente", "consultar saldo", "quanto tenho na conta", "extrato da conta", "saldo disponível", "meu saldo atual"
		
		"13": Pagamento de contas
		Intenções de exemplo: "quero pagar minha conta", "pagar boleto", "pagamento de conta", "quero pagar fatura", "efetuar pagamento", "pagamento da fatura", "quitar fatura", "pagar meu cartão", "pagar fatura do cartão", "liquidar fatura", "fazer o pagamento"
		
		"14": Reclamações
		Intenções de exemplo: "quero reclamar", "abrir reclamação", "fazer queixa", "reclamar atendimento", "registrar problema", "protocolo de reclamação"
		
		"15": Atendimento humano
		Intenções de exemplo: "falar com uma pessoa", "preciso de humano", "transferir para atendente", "quero falar com atendente", "atendimento pessoal"
		
		"16": Token de proposta
		Intenções de exemplo: "código para fazer meu cartão", "token de proposta", "receber código do cartão", "proposta token", "número de token", "código de token da proposta"
	`
