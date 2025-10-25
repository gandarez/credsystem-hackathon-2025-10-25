# Credsystem & Golang SP - Hackathon 25/10/2025

A Credsystem em parceria com a Golang SP os convidam para o Hackathon em 25/10/2025.

## Pré-requisitos

- Conta no [GitHub](https://github.com)
- Ambiente de desenvolvimento Golang
- Docker instalado
- Conhecimento básico em Git
- Noções básicas de IA. Se não tiver, não se preocupe, vamos ajudar!
- Chave de API da [OpenRouter](https://openrouter.ai/) (**será fornecida durante o evento com limite de $3 de uso**)

## Descrição

O desafio é criar um serviço que expõe 2 endpoints RESTful:

1. `POST /api/find-service`: Retorna qual o serviço mais adequado para o tipo de solicitação.
2. `GET /api/healthz`: Verifica a saúde do serviço. Retorna 200 OK se o serviço estiver funcionando corretamente.

### Endpoints

- `POST /api/find-service`

  - Request Body:

    ```json
    {
      "intent": "string"
    }
    ```

  - Response Body:

    ```json
    {
      "success": "bool",
      "data": {
        "service_id": "int",
        "service_name": "string",
      },
      "error": "string"
    }
    ```

- `GET /api/healthz`

  - Response Body:

    ```json
    {
        "status": "ok"
    }
    ```

## O Desafio

Atualmente a URA da Credsystem direciona as ligações para o setor correto com base na solicitação do cliente. O objetivo deste desafio é automatizar esse processo utilizando IA para analisar a solicitação e direcioná-la ao serviço mais adequado.

Então você deve:

1. Implementar no endpoint `POST /api/find-service` a lógica para determinar o serviço mais adequado com base na solicitação recebida. Utilize técnicas de IA para analisar a solicitação e escolher o serviço correto.

2. Seu serviço deve ler a variável de ambiente `OPENROUTER_API_KEY` para autenticar as requisições à API da OpenRouter. Ele também deve ler a variável `PORT` para definir a porta em que o serviço irá rodar.

3. No arquivo `./assets/intents_pre_loaded.csv` você encontrará uma lista contendo as 93  intenções iniciais e seus respectivos serviços. Utilize essa lista para treinar seu modelo de IA.

4. **Não invente serviços, utilize um dos 16 serviços** listados:

- Consulta Limite / Vencimento do cartão / Melhor dia de compra (ID 1)
- Segunda via de boleto de acordo (ID 2)
- Segunda via de Fatura (ID 3)
- Status de Entrega do Cartão (ID 4)
- Status de cartão (ID 5)
- Solicitação de aumento de limite (ID 6)
- Cancelamento de cartão (ID 7)
- Telefones de seguradoras (ID 8)
- Desbloqueio de Cartão (ID 9)
- Esqueceu senha / Troca de senha (ID 10)
- Perda e roubo (ID 11)
- Consulta do Saldo (ID 12)
- Pagamento de contas (ID 13)
- Reclamações (ID 14)
- Atendimento humano (ID 15)
- Token de proposta (ID 16)

5. Faça um fork deste repositório criando uma pasta com o nome da sua dupla no diretório `participantes`.

6. O build será feito durante a execução da sua imagem Docker. Crie um Dockerfile conforme exemplo na pasta `./examples` para construir a imagem do seu serviço. Apenas preencher a imagem Docker criada por você.

7. Crie um arquivo docker-compose conforme exemplo na pasta `./examples` para facilitar a execução do seu serviço. Apenas preencher a imagem Docker criada por você.

8. Sua API só terá direito a 50% de uma CPU e 128MB de memória RAM. Deixe esse valor fixo no docker-compose conforme exemplo.

9. Seja cuidadoso na hora de escolher o modelo de IA a ser utilizado, lembre-se que o serviço terá recursos limitados. Para mais informaçoes sobre os modelos disponíveis na OpenRouter, consulte a [documentação oficial](https://openrouter.ai/models?o=pricing-high-to-low).

### Critérios de aprovação do PR

- O código fonte deve estar na pasta `participantes/NOME_DA_DUPLA/`.
- O Dockerfile deve estar na pasta `participantes/NOME_DA_DUPLA/`.
- O docker-compose.yml deve estar na pasta `participantes/NOME_DA_DUPLA/`.
- O serviço deve ler uma variável de ambiente: `OPENROUTER_API_KEY`.
- O serviço deve expor a porta *18020* conforme o exemplo do docker-compose.
- O serviço deve estar em conformidade com os limites de recursos especificados de 50% de CPU e 128MB de RAM definidos no docker-compose.

## Critérios de avaliação da entrega

A avaliação será baseada em **duas rodadas de testes**:

1. **Teste 93**: Retornar corretamente os serviços para as 93 intenções fornecidas no arquivo `./assets/intents_pre_loaded.csv`.
2. **Teste 80**: Retornar o serviço mais adequado para 80 intenções similares (*5 para cada serviço*). Neste caso essas intenções não estão no CSV fornecido e serão base para execução da segunda rodada de testes.

### Sistema de Pontuação

Cada participante receberá uma pontuação calculada da seguinte forma:

- **Sucessos**: +10.0 pontos por cada resposta correta (tanto no Teste 93 quanto no Teste 80)
- **Falhas**: -50.0 pontos por cada resposta incorreta ou erro
- **Tempo de Resposta**: -0.01 pontos por milissegundo de tempo médio de resposta (média dos dois testes)

**Fórmula**: `Score = (Total_Sucessos × 10.0) - (Total_Falhas × 50.0) - (Tempo_Médio_ms × 0.01)`

### Ranking Final

Os participantes serão ranqueados pela **maior pontuação total**. Quanto maior o score, melhor a colocação.

**Exemplo de cálculo**:

- 171 sucessos, 2 falhas, tempo médio de 3270ms
- Score = (171 × 10) - (2 × 50) - (3270 × 0.01) = 1710 - 100 - 32.7 = **1577.3 pontos**

## Modelos de IA

Recomendamos o uso dos seguintes modelos da OpenRouter, que oferecem um bom equilíbrio entre desempenho e custo:

- **Mistral 7B**: Um modelo eficiente e econômico, ideal para tarefas de compreensão de linguagem natural.
- **openai/gpt-4o-mini**: Um modelo compacto e eficiente, ideal para aplicações que requerem um bom desempenho em um pacote menor.

## Consultar créditos restantes na OpenRouter

Para consultar os créditos restantes na sua conta da OpenRouter, você pode utilizar o seguinte script Python:

> Não se esqueça de substituir `<seu_token_aqui>` pela sua chave de API real.

```shell
python ./utils/check_limit_openrouter.py
```

Made with :heart: by the Golang SP.
