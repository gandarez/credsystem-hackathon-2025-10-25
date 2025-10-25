# Mapa de Serviços do MCP

Este documento consolida os serviços disponíveis, com ID, nome e uma descrição detalhada para auxiliar o roteamento inteligente de intenções (intents) para a ferramenta correta. Inclui objetivos, quando usar, entradas, retornos esperados, diferenciações e exemplos reais de intenções extraídos do CSV.

---

## Serviço 1 — Consulta Limite / Vencimento do cartão / Melhor dia de compra
- ID: 1
- Descrição detalhada: Centraliza consultas sobre ciclo de fatura e capacidade de compra do cartão de crédito. Permite verificar limite total e disponível, data de fechamento da fatura (quando novas compras passam ao próximo ciclo), data de vencimento, e o melhor dia de compra (janela ideal para que a compra entre somente na fatura seguinte). Apoia planejamento de gastos, evita atrasos e otimiza compras próximas ao fechamento.
- Quando usar: Dúvidas como “quanto posso gastar”, “quando fecha/vencem fatura/cartão”, “qual o melhor dia para comprar”, “posso comprar hoje sem cair na fatura atual?”.
- Entradas típicas: Identificador do cliente (ex.: CPF), confirmações de segurança, cartão (últimos dígitos) se necessário.
- Retornos esperados: Limite total, limite disponível, data de fechamento, data de vencimento, melhor dia de compra, observações de uso do ciclo (ex.: compras parceladas, autorizações pendentes).
- Diferenciações: Não cobre emissão de fatura/boletos (ver Serviço 3) nem saldo de conta corrente (ver Serviço 12).
- Exemplos de intenções (CSV):
  - Quanto tem disponível para usar
  - quando fecha minha fatura
  - Quando vence meu cartão
  - quando posso comprar
  - vencimento da fatura
  - valor para gastar

---

## Serviço 2 — Segunda via de boleto de acordo
- ID: 2
- Descrição detalhada: Emite ou reenvia a 2ª via do boleto de um acordo/renegociação (parcelamento de dívida/negociação). Fornece linha digitável/código de barras atualizado, valor e vencimento corretos, reduzindo risco de pagamento com documento desatualizado. Pode entregar PDF/URL e enviar por e-mail/SMS.
- Quando usar: Cliente menciona acordo/renegociação e pede “boleto”, “segunda via”, “código de barras do acordo”, “pagar minha negociação”.
- Entradas típicas: Identificação do acordo (nº ou CPF/CNPJ), parcela desejada, canal de entrega (e-mail/SMS), validações de segurança.
- Retornos esperados: Linha digitável, PDF/URL, valor/vencimento atualizados, instruções de pagamento.
- Diferenciações: Se o pedido for da fatura do cartão (mensal), usar Serviço 3 (Segunda via de Fatura). Este serviço é específico de “acordo/renegociação”.
- Exemplos de intenções (CSV):
  - segunda via boleto de acordo
  - Boleto para pagar minha negociação
  - código de barras acordo
  - preciso pagar negociação
  - enviar boleto acordo
  - boleto da negociação

---

## Serviço 3 — Segunda via de Fatura
- ID: 3
- Descrição detalhada: Gera ou reenvia a 2ª via da fatura do cartão. Entrega boleto/código de barras para pagamento, com valores e encargos atualizados. Atende fatura atual, vencida ou, conforme política, antecipação.
- Quando usar: Pedidos de “quero meu boleto”, “segunda via de fatura”, “código de barras da fatura”, “enviar boleto da fatura”, “fatura para pagamento”.
- Entradas típicas: Identificador do cliente, referência da fatura (competência/mês), canal de entrega (e-mail/SMS/app), validações.
- Retornos esperados: Linha digitável, PDF/URL, valor atualizado, vencimento, eventuais encargos/juros, instruções.
- Diferenciações: Se mencionar “acordo/renegociação”, usar Serviço 2. Se for apenas consultar limites/datas da fatura, usar Serviço 1.
- Exemplos de intenções (CSV):
  - quero meu boleto
  - segunda via de fatura
  - código de barras fatura
  - quero a fatura do cartão
  - enviar boleto da fatura
  - fatura para pagamento

---

## Serviço 4 — Status de Entrega do Cartão
- ID: 4
- Descrição detalhada: Consulta status logístico do cartão físico (produção, expedição, em transporte, tentativa de entrega, entregue, devolvido). Quando integrado a transportadoras, exibe tracking, previsão de entrega e orientações em caso de insucesso.
- Quando usar: Dúvidas “onde está meu cartão”, “não chegou”, “status/previsão de entrega”, “foi enviado?”.
- Entradas típicas: Identificador do cliente, endereço cadastrado para conferência, dados de envio.
- Retornos esperados: Status atual, histórico/tracking, previsão, última atualização, ações recomendadas (ex.: reentrega, atualização de endereço).
- Diferenciações: Não cobre desbloqueio/ativação do cartão (ver Serviço 9).
- Exemplos de intenções (CSV):
  - onde está meu cartão
  - meu cartão não chegou
  - status da entrega do cartão
  - cartão em transporte
  - previsão de entrega do cartão
  - cartão foi enviado?

---

## Serviço 5 — Status de cartão
- ID: 5
- Descrição detalhada: Verifica a situação operacional do cartão (ativo, bloqueado temporariamente, bloqueado por segurança, vencido, com restrição). Explica recusas em transações, orienta desbloqueio/reativação e verifica pendências (senha bloqueada, atualização cadastral).
- Quando usar: “cartão recusado/não passa”, “meu cartão não funciona”, “status do cartão ativo?”, “problema com cartão”.
- Entradas típicas: Identificação do cliente/portador, últimos dígitos do cartão, validações de segurança.
- Retornos esperados: Status, motivo (quando disponível), próximos passos (desbloqueio, troca de senha, regularizações), dicas de uso.
- Diferenciações: Bloqueio por perda/roubo é tratado no Serviço 11; desbloqueio/ativação no Serviço 9.
- Exemplos de intenções (CSV):
  - não consigo passar meu cartão
  - meu cartão não funciona
  - cartão recusado
  - cartão não está passando
  - status do cartão ativo
  - problema com cartão

---

## Serviço 6 — Solicitação de aumento de limite
- ID: 6
- Descrição detalhada: Recebe e avalia pedido de aumento de limite (temporário ou permanente). Apura elegibilidade (histórico, renda, relacionamento) e encaminha para análise automática ou manual, informando prazo e resultado.
- Quando usar: “quero mais limite”, “aumentar limite”, “solicitar aumento de crédito”, “preciso de mais limite”.
- Entradas típicas: Identificador do cliente, renda/atualização cadastral (se necessário), valor pretendido ou sugestão automática.
- Retornos esperados: Status (pré-aprovado, em análise, aprovado, recusado), novo limite (se aprovado), validade (se temporário), justificativas.
- Diferenciações: Para consultar limite atual, usar Serviço 1. Para problemas de recusa por limite disponível insuficiente, combinar com Serviço 1.
- Exemplos de intenções (CSV):
  - quero mais limite
  - aumentar limite do cartão
  - solicitar aumento de crédito
  - preciso de mais limite
  - pedido de aumento de limite
  - limite maior no cartão

---

## Serviço 7 — Cancelamento de cartão
- ID: 7
- Descrição detalhada: Processa encerramento definitivo do cartão. Explica impactos (parcelas abertas, faturas pendentes), alternativas (bloqueio temporário) e coleta consentimento. Pode acionar reemissão em casos de substituição.
- Quando usar: “cancelar cartão”, “encerrar meu cartão”, “bloquear definitivamente”, “cancelamento de crédito”.
- Entradas típicas: Autenticação reforçada, motivo do cancelamento, confirmação sobre pendências.
- Retornos esperados: Confirmação e protocolo, instruções sobre faturas/parcelas remanescentes, prazo de efetivação.
- Diferenciações: Se for perda/roubo, usar Serviço 11 (bloqueio imediato e reemissão). Para bloqueio temporário, avaliar Serviço 5/9 conforme o fluxo disponível.
- Exemplos de intenções (CSV):
  - cancelar cartão
  - quero encerrar meu cartão
  - bloquear cartão definitivamente
  - cancelamento de crédito
  - desistir do cartão

---

## Serviço 8 — Telefones de seguradoras
- ID: 8
- Descrição detalhada: Fornece contatos oficiais (telefone, horários, canais) das seguradoras e assistências vinculadas aos produtos do cartão (seguros, assistência residencial/auto, etc.). Útil para cancelamento, sinistro ou suporte técnico.
- Quando usar: “telefone do seguro”, “cancelar seguro/assistência”, “contato da seguradora”, “preciso falar com o seguro”.
- Entradas típicas: Produto/seguro contratado (se houver), identificação do cliente, preferência de canal.
- Retornos esperados: Números de contato, horários, canais alternativos, passos iniciais (nº de apólice, documentos necessários).
- Diferenciações: Cancelamento de seguro pode exigir contato direto com a seguradora; este serviço provê os meios.
- Exemplos de intenções (CSV):
  - quero cancelar seguro
  - telefone do seguro
  - contato da seguradora
  - preciso falar com o seguro
  - seguro do cartão
  - cancelar assistência

---

## Serviço 9 — Desbloqueio de Cartão
- ID: 9
- Descrição detalhada: Realiza ativação do cartão recém-chegado ou desbloqueio após bloqueio preventivo. Confirma identidade com múltiplos fatores (token, dados cadastrais) e habilita o cartão para uso imediato.
- Quando usar: “desbloquear/ativar cartão”, “como desbloquear”, “cartão para uso imediato”, “desbloqueio para compras”.
- Entradas típicas: Identificação do cliente, últimos dígitos do cartão, validações (token/senha provisória), confirmação de posse.
- Retornos esperados: Confirmação de desbloqueio, status final do cartão, orientações de primeira utilização e segurança.
- Diferenciações: Para redefinição de senha/PIN, usar Serviço 10.
- Exemplos de intenções (CSV):
  - desbloquear cartão
  - ativar cartão novo
  - como desbloquear meu cartão
  - quero desbloquear o cartão
  - cartão para uso imediato
  - desbloqueio para compras

---

## Serviço 10 — Esqueceu senha / Troca de senha
- ID: 10
- Descrição detalhada: Suporta recuperação/definição de nova senha (PIN) do cartão ou troca voluntária. Aplica políticas de segurança (verificação de identidade, tentativas, pressão de tempo) e orienta criação segura.
- Quando usar: “esqueci minha senha”, “trocar senha do cartão”, “preciso de nova senha”, “senha bloqueada”.
- Entradas típicas: Identificação do cliente, validações (token, dados cadastrais), canal de confirmação (SMS/e-mail/app).
- Retornos esperados: Confirmação do processo (nova senha definida, desbloqueio), prazo de propagação, instruções de uso.
- Diferenciações: Se o problema é o cartão estar bloqueado por segurança, combinar com Serviços 5 e 9. Para fatura, ver Serviços 1/3.
- Exemplos de intenções (CSV):
  - não tenho mais a senha do cartão
  - esqueci minha senha
  - trocar senha do cartão
  - preciso de nova senha
  - recuperar senha
  - senha bloqueada

---

## Serviço 11 — Perda e roubo
- ID: 11
- Descrição detalhada: Atendimento emergencial para perda, roubo ou furto. Realiza bloqueio imediato, orienta contestação de compras indevidas, e inicia reemissão/envio de novo cartão com segurança reforçada.
- Quando usar: “perdi meu cartão”, “roubaram/furtado”, “extravio”, “bloquear cartão por roubo”.
- Entradas típicas: Identificação do cliente, confirmação do evento (perda/roubo), local e data aproximados, opção de reemissão e endereço de entrega.
- Retornos esperados: Bloqueio confirmado, protocolo, início e tracking da reemissão, instruções de contestação e prazos.
- Diferenciações: Não é cancelamento definitivo (Serviço 7). Se o cartão foi encontrado e está em mãos do cliente, avaliar desbloqueio (Serviço 9) conforme política.
- Exemplos de intenções (CSV):
  - perdi meu cartão
  - roubaram meu cartão
  - cartão furtado
  - perda do cartão
  - bloquear cartão por roubo
  - extravio de cartão

---

## Serviço 12 — Consulta do Saldo
- ID: 12
- Descrição detalhada: Consulta saldo de conta corrente (e quando aplicável, extrato simplificado/lançamentos recentes). Ajuda a entender saldo disponível, bloqueado e movimentações.
- Quando usar: “saldo conta corrente”, “consultar saldo”, “quanto tenho na conta”, “extrato da conta”, “saldo disponível”.
- Entradas típicas: Identificação da conta/cliente, período (para extrato), filtros básicos.
- Retornos esperados: Saldo atual/disponível/bloqueado, extrato recente, data/hora da última atualização.
- Diferenciações: Não confundir com limite do cartão (Serviço 1). Para pagamento de contas, usar Serviço 13.
- Exemplos de intenções (CSV):
  - saldo conta corrente
  - consultar saldo
  - quanto tenho na conta
  - extrato da conta
  - saldo disponível
  - meu saldo atual

---

## Serviço 13 — Pagamento de contas
- ID: 13
- Descrição detalhada: Efetua pagamentos de boletos/contas (incluindo, quando aplicável, fatura do cartão). Lê linha digitável, valida valor e vencimento, permite agendamento, confirma protocolo e comprovante.
- Quando usar: “pagar boleto/conta”, “efetuar pagamento”, “pagar fatura”.
- Entradas típicas: Linha digitável/código de barras, valor (se permitido alterar), data de pagamento/agendamento, conta de débito, autenticação.
- Retornos esperados: Protocolo/ID da transação, comprovante (PDF/URL), status (agendado/efetuado), eventuais taxas/limites.
- Diferenciações: Para obter a 2ª via da fatura, usar Serviço 3. Para 2ª via de acordo, usar Serviço 2.
- Exemplos de intenções (CSV):
  - quero pagar minha conta
  - pagar boleto
  - pagamento de conta
  - quero pagar fatura
  - efetuar pagamento

---

## Serviço 14 — Reclamações
- ID: 14
- Descrição detalhada: Recebe manifestações sobre experiências com os serviços do cartão e seus canais (fatura, pagamento, entrega do cartão, limites, acordos, atendimento e app/site). Registra protocolo, organiza o relato por tema, encaminha à área responsável e acompanha a devolutiva para promover melhorias contínuas na jornada do cliente.
- Quando usar: Para compartilhar percepções construtivas sobre fatura (clareza de informações e organização dos lançamentos), jornada de pagamento (confirmação e comprovantes), entrega do cartão (visibilidade de status e comunicação), limites e acordos (compreensão de prazos e etapas), atendimento (cordialidade, clareza, tempo de resposta) e app/site (facilidade de uso e acessibilidade).
- Entradas típicas: Identificação do cliente, tema, descrição objetiva do relato, período/ocorrência, canal preferido para contato e anexos úteis (ex.: comprovantes, prints).
- Retornos esperados: Protocolo, prazo estimado de retorno, status de acompanhamento (aberto/em análise/concluído) e próximos passos.
- Diferenciações: Para resolver rapidamente necessidades operacionais (ex.: 2ª via, desbloqueio, limites), utilize o serviço dedicado correspondente; para registrar e acompanhar a experiência vivida, utilize este canal de manifestação.
- Exemplos de intenções (CSV):
  - quero reclamar
  - abrir reclamação
  - fazer queixa
  - reclamar atendimento
  - registrar problema
  - protocolo de reclamação

---

## Serviço 15 — Atendimento humano
- ID: 15
- Descrição detalhada: Transfere para um agente humano quando necessário (casos complexos, falhas na automação, preferência do cliente). Mantém o contexto da conversa para evitar reexplicações.
- Quando usar: “falar com uma pessoa”, “preciso de humano”, “transferir para atendente”, “quero falar com atendente”.
- Entradas típicas: Identificação do cliente, breve resumo/motivo do contato, canal preferido.
- Retornos esperados: Confirmação de transferência, posição na fila/tempo estimado, continuidade do atendimento humano.
- Diferenciações: Antes de transferir, avaliar se há um serviço específico que resolve a demanda. Se o usuário insistir, priorizar transferência.
- Exemplos de intenções (CSV):
  - falar com uma pessoa
  - preciso de humano
  - transferir para atendente
  - quero falar com atendente
  - atendimento pessoal

---

## Serviço 16 — Token de proposta
- ID: 16
- Descrição detalhada: Gera, valida ou reenvia o token/código usado para iniciar, continuar ou validar uma proposta de cartão/serviço financeiro. Autentica a jornada de proposta, evita fraudes, e permite retomar processos interrompidos.
- Quando usar: “código para fazer meu cartão”, “token de proposta”, “receber código do cartão”, “número/código da proposta”.
- Entradas típicas: Identificação do proponente, canal de envio (SMS/e-mail), verificação de contato, vínculo com ID de proposta (se houver).
- Retornos esperados: Token válido/novo, validade/expiração, instruções de uso, status de associação à proposta.
- Diferenciações: Não substitui validações de identidade completas. Para desbloqueio/ativação de cartão físico, ver Serviço 9.
- Exemplos de intenções (CSV):
  - código para fazer meu cartão
  - token de proposta
  - receber código do cartão
  - proposta token
  - número de token
  - código de token da proposta

---

Observação: Em todos os serviços, respeitar políticas de segurança, LGPD e melhores práticas de confirmação de identidade antes de expor dados sensíveis ou realizar operações críticas.

---

## Anexo — Política de Roteamento e Calibração do Classificador

Objetivo: orientar o LLM/classificador a escolher corretamente os serviços e retornar “nenhum serviço encontrado” quando o texto não tiver relação com o domínio financeiro/cartões.

### Política de Roteamento
1) Selecione um serviço apenas se houver menção clara a termos do domínio: ["cartão", "fatura", "boleto", "pagamento", "limite", "acordo", "senha", "desbloqueio", "entrega", "seguro", "atendimento", "app", "site", "proposta", "token", "saldo"].
2) Não use “Reclamações” (Serviço 14) como fallback. Utilize-o somente se a manifestação tratar da experiência com algum serviço/canal listado e houver ao menos uma âncora do domínio.
3) Se não houver correspondência clara, retorne exatamente:
```
{"success": false, "data": {}, "error": "nenhum serviço encontrado"}
```
4) A resposta deve ser apenas JSON válido (sem explicações ou raciocínio livre).

### Regras de Qualificação — Serviço 14 (Reclamações)
- Use o Serviço 14 apenas quando o texto indicar manifestação sobre a experiência com: fatura (clareza de lançamentos), pagamento (confirmação/comprovantes), entrega do cartão (status/comunicação), limites e acordos (prazos/etapas), atendimento (cordialidade/clareza/tempo de resposta), app/site (usabilidade/acessibilidade) ou demais serviços listados (senha, desbloqueio, seguro, etc.).
- Exija presença de pelo menos uma âncora do domínio (ex.: "cartão", "fatura", "pagamento", "app", "site", "atendimento", "limite", "acordo", "senha", "desbloqueio", "seguro", "entrega", "protocolo").
- Não promova o Serviço 14 apenas por conter verbos genéricos como “reclamar” sem relação ao domínio acima.

### Few-shot de calibração
- Input: “Quero registrar uma manifestação sobre a clareza dos lançamentos na minha fatura.”
  - Output:
```
{"success": true, "data": {"service_id": 14, "service_name": "Reclamações"}, "error": ""}
```
- Input: “O status de entrega do meu cartão não atualiza no app.”
  - Output:
```
{"success": true, "data": {"service_id": 4, "service_name": "Status de Entrega do Cartão"}, "error": ""}
```
- Input: “Fui pagar uma pizza e fiquei com dor de barriga.”
  - Output:
```
{"success": false, "data": {}, "error": "nenhum serviço encontrado"}
```

### Pré-filtro léxico (opcional)
Antes de chamar o LLM, utilize um checape simples. Se nenhuma âncora global for encontrada, retorne diretamente “nenhum serviço encontrado”. Para o Serviço 14, só considere se houver ao menos uma âncora de experiência/canal.

Exemplo em Go:
```go
var globalAnchors = []string{"cartão","fatura","boleto","pagamento","limite","acordo","senha","desbloqueio","entrega","seguro","atendimento","app","site","proposta","token","saldo"}
var reclamacoesAnchors = []string{"reclama","manifestação","protocolo","atendimento","app","site","fatura","pagamento","entrega","limite","acordo","seguro","senha","desbloqueio","cartão"}

func containsAny(s string, terms []string) bool {
    s = strings.ToLower(s)
    for _, t := range terms {
        if strings.Contains(s, t) { return true }
    }
    return false
}

func prefilter(text string) (ok bool, isReclamacoesEligible bool) {
    hasDomain := containsAny(text, globalAnchors)
    if !hasDomain { return false, false }
    isRecl := containsAny(text, reclamacoesAnchors)
    return true, isRecl
}
```

Uso do pré-filtro:
- Se `ok == false`, retorne `{ "success": false, "data": {}, "error": "nenhum serviço encontrado" }`.
- Se `ok == true` e `isReclamacoesEligible == false`, não permita escolher o Serviço 14 apenas por similaridade vaga; avalie os demais serviços pela descrição.
