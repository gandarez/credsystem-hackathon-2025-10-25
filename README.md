# credsystem-hackathon-2025-10-25

## Setup Github

Esse desafio vai usar a sua imagem buildada para ser executado localmente, caso queira fazer o pull dela, siga esses passos:

- Gerar um novo Personal Access Token (classic) [aqui](https://github.com/settings/tokens)
 - Com pelo menos a permissão `read:packages`
- Realizar o login no docker registry do Github
 - `docker login ghcr.io`
 - Com seu usuário do Github como `Username`
 - Seu PAT como `Password`
- Rodar o comando alterando a tag da imagem pelo seu usuário do Github
 - `docker pull ghcr.io/gandarez/credsystem-hackathon-2025-10-25:<TAG>` 
