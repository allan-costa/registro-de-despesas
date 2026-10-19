# MVP Sistema de Registro de Despesas com Inteligência Artificial Generativa

# Objetivo

- Permitir que os usuários registrem despesas realizadas, informando o valor, a data, o tipo (como alimentação ou transporte) e a categoria. O registro representa uma despesa já incorrida, mesmo que ainda não tenha sido paga.
- Sistema criado com auxilio da ferramenta **Copilot (modelos GPT-6 Luna e Cloud Sonnet 5.5)**

## Stack

- **Sistema operacional:** GNU/Linux, documentado para Debian 13 e Ubuntu 22.04.
- **Backend:** Go 1.26.0 ou superior, Gin v1.9 e PostgreSQL usando `database/sql` e `lib/pq` (sem ORM).
- **Autenticação:** JWT v5 e bcrypt (`golang-jwt/jwt/v5`, `golang.org/x/crypto`).
- **Frontend:** HTML5, Tailwind CSS via CDN, Chart.js via CDN e JavaScript ES6.
- **Outras dependências Go:** `godotenv` e `google/uuid`.

## Estrutura de Diretórios

```
registro-despesas/
├── main.go
├── go.mod / go.sum
├── .env
├── config/config.go
├── internal/
│   ├── database/postgres.go
│   ├── middleware/auth.go, security.go, ratelimit.go
│   ├── models/models.go
│   ├── repository/auth_repo.go, categoria_repo.go, despesa_repo.go, usuario_repo.go, despesa_repo.go
│   ├── service/auth_service.go, categoria_service.go, despesa_service.go, usuario_service.go, despesa_service.go
│   └── handler/auth_handler.go, categoria_handler.go, despesa_handler.go, usuario_handler.go, despesa_handler.go
└── public/
    ├── css/estilo.css
    ├── js/auth.js, api.js, layout.js, main.js
    └── login.html, index.html, categoria.html, despesa.html, usuario.html
```

## Requisitos

- Go 1.26.0 ou superior. Os módulos da aplicação estão definidos em `go.mod` e `go.sum`; Python e Node.js não são necessários.
- PostgreSQL instalado e acessível. O projeto não fixa uma versão específica do servidor.
- Cliente `psql` e extensão `pgcrypto` disponíveis no PostgreSQL; o script SQL habilita essa extensão.
- Navegador web moderno.
- Acesso à internet para baixar módulos Go e carregar Tailwind CSS, Chart.js e Material Symbols via CDN.

## Configuração

A aplicação lê variáveis de ambiente ou, opcionalmente, um arquivo `.env` na raiz do projeto. As configurações padrão são:

| Variável | Padrão | Descrição |
| --- | --- | --- |
| `DATABASE_URL` | `postgres://postgres@localhost:5432/registro_despesas?sslmode=disable` | URL de conexão do PostgreSQL. |
| `JWT_SECRET` | Gerado aleatoriamente ao iniciar | Defina um segredo forte e persistente para que os tokens continuem válidos após reiniciar a aplicação. |
| `SERVER_PORT` | `5000` | Porta HTTP da aplicação. |
| `CORS_ORIGIN` | `*` | Origem permitida para requisições CORS. |

O arquivo `.env` incluído no projeto contém apenas configurações não sensíveis. Não armazene nele senhas, tokens ou segredos reais; em produção, forneça esses valores por variáveis de ambiente ou por um gerenciador de segredos. Se `JWT_SECRET` não for definido, um valor aleatório é gerado a cada inicialização.

## Execução local

1. Crie o banco de dados: `createdb -U postgres registro_despesas`.
2. Aplique o esquema e os dados iniciais: `psql -U postgres -d registro_despesas -f postgres.sql`.
3. Se necessário, configure `DATABASE_URL` para as credenciais do seu PostgreSQL.
4. Baixe as dependências: `go mod download`.
5. Inicie a aplicação na raiz do projeto: `go run .`.
6. Acesse [http://localhost:5000/login.html](http://localhost:5000/login.html).

## Testes unitários

Os testes dos serviços de categorias, usuários e despesas usam repositórios em memória e não precisam de conexão com o PostgreSQL. Execute-os com:

```sh
go test ./internal/service -v
```

| Operação/cenário | Resultado esperado | Resultado obtido |
| --- | --- | --- |
| Categorias: cadastro, consulta, edição e exclusão | Normalizar e criar uma categoria, listá-la, editar seu nome e removê-la. | Passou: cada operação retornou o estado esperado. |
| Categorias: hierarquia e entradas inválidas | Listar `CONTAS` como filha de `CASA`; rejeitar nome vazio, pai inexistente, ciclo e edição de categoria inexistente. | Passou: hierarquia e quatro erros corresponderam ao esperado. |
| Categorias: exclusão inválida | Traduzir categoria inexistente e categoria vinculada a despesas para erros específicos. | Passou: os dois erros corresponderam ao esperado. |
| Usuários: cadastro, consulta, edição e exclusão | Normalizar nome e e-mail, armazenar senha como hash bcrypt, listar, editar sem trocar a senha quando ela não é informada e excluir. | Passou: operações e preservação do hash corresponderam ao esperado. |
| Usuários: entradas inválidas e proteção da conta | Rejeitar nome vazio, senha ausente/curta, confirmação divergente, e-mail duplicado, autorrebaixamento e autoexclusão. | Passou: os erros esperados foram retornados. |
| Despesas: cadastro, consulta, edição e exclusão | Criar e listar despesa, calcular total, editar descrição e valor e excluir o registro. | Passou: operações e total corresponderam ao esperado. |
| Despesas: parcelamento | Gerar duas parcelas numeradas e ajustar a data de 31/01 para 28/02 em fevereiro. | Passou: parcelas, numeração e datas corresponderam ao esperado. |
| Despesas: entradas inválidas e registro inexistente | Rejeitar descrição vazia, valor zero, categoria inválida, data/tipo inválidos e número de parcelas fora do limite; mapear edição/exclusão inexistentes. | Passou: os erros esperados foram retornados. |

Resultado da execução: todos os testes e subtestes passaram (`PASS`, pacote `internal/service`).

## Autenticação

A autenticação é feita através do token JWT. Para gerar um token, você precisa fornecer um usuário e uma senha. O usuário inicial é `admin@admin.com` e a senha é `123456`.

## Licença

GNU General Public License v3.0 (GPL-3.0). Você pode usar, estudar, modificar e redistribuir este projeto, inclusive versões modificadas, desde que mantenha a mesma licença e disponibilize o código-fonte correspondente. O projeto é fornecido sem garantia.
