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

## Execução local

1. Crie o banco de dados: `createdb -U postgres registro_despesas`.
2. Aplique o esquema e os dados iniciais: `psql -U postgres -d registro_despesas -f postgres.sql`.
3. Se necessário, configure `DATABASE_URL` para as credenciais do seu PostgreSQL.
4. Baixe as dependências: `go mod download`.
5. Inicie a aplicação na raiz do projeto: `go run .`.
6. Acesse [http://localhost:5000/login.html](http://localhost:5000/login.html).


## Autenticação

A autenticação é feita através do token JWT. Para gerar um token, você precisa fornecer um usuário e uma senha. O usuário inicial é `admin@admin.com` e a senha é `123456`.

## Licença

GNU General Public License v3.0 (GPL-3.0). Você pode usar, estudar, modificar e redistribuir este projeto, inclusive versões modificadas, desde que mantenha a mesma licença e disponibilize o código-fonte correspondente. O projeto é fornecido sem garantia.
