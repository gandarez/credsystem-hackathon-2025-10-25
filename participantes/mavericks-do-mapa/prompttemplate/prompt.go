package prompttemplate

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"
)

const userPromptTemplate = `Você é um atendente da credsystem e esta atendendo a um chamado de um cliente que chegou com um problema sua resposta deve ser restrita em um formato json no seguinte formato 
{
  "success": "bool",
  "data": {
    "service_id": "int",
    "service_name": "string",
  },
  "error": "string"
}

vou te passar as informações do serviço, se não encontrar nenhum serviço que atenda aos requisitos do chamado do cliente retorne

{
  "success": false,
  "data": {},
  "error": "nenhum serviço encontrado"
}

caso encontre retorne o o objeto que mandei de exemplo populado com sucesso, o id do serviço, o nome do serviço e nenhuma mensagem de erro

<serviços disponiveis>
%s
</serviços disponiveis>

o chamado do usuario é 
<chamado do usuario>
%s
</chamado do usuario>
`

//go:embed servicos_mcp.md
var servicesCatalog string

// BuildPrompt returns the formatted prompt with the service catalog and user intent.
func BuildPrompt(intent string) (string, error) {
	trimmedIntent := strings.TrimSpace(intent)
	if trimmedIntent == "" {
		return "", errors.New("intent cannot be empty")
	}

	catalog := strings.TrimSpace(servicesCatalog)
	if catalog == "" {
		return "", errors.New("services catalog is empty")
	}

	return fmt.Sprintf(userPromptTemplate, catalog, trimmedIntent), nil
}
