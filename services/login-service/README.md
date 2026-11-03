# Serviço de Login e Autenticação

## Responsabilidade

O serviço gerencia o cadastro de usuários e a autenticação do EcoPulse.
Ele utiliza PostgreSQL para persistência, JWT para autenticação e
RabbitMQ para publicar eventos relacionados ao ciclo de vida dos
usuários.

**Porta:** `8080`

## Como executar

Na raiz do projeto, execute:

``` bash
docker compose up -d --build
```

Para verificar os containers:

``` bash
docker compose ps
```

Para acompanhar os logs do serviço:

``` bash
docker compose logs -f login-service
```

Para parar os serviços:

``` bash
docker compose down
```

O `docker-compose.yml` deve fornecer as seguintes variáveis ao serviço:

``` env
DB_HOST=postgres
DB_PORT=5432
DB_USERNAME=ecopulse
DB_PASSWORD=ecopulse_dev
DB_NAME=ecopulse_auth
JWT_SECRET=ecopulse-segredo-local-123456789012345
RABBITMQ_URL=amqp://ecopulse:ecopulse_dev@rabbitmq:5672/
```

Os valores acima são de desenvolvimento. É importante que o JWT_SECRET seja o mesmo informado acima em todos os testes.

O serviço cria automaticamente as tabelas `users` e `outbox_events` ao
iniciar. A execução por Docker Compose depende de o `Dockerfile` estar
configurado corretamente na raiz do projeto.

## Endpoints

URL base local: `http://localhost:8080`

### Autenticação

#### `POST /auth/login`

Realiza o login e retorna um JWT.

**Requisição:**

``` json
{
  "username": "usuario",
  "password": "senha"
}
```

**Resposta `200 OK`:**

``` json
{
  "token": "<jwt>",
  "token_type": "Bearer",
  "expires_in": 3600
}
```

O token utiliza o algoritmo HS256 e expira em uma hora. É ele que será usado como Bearer na requisição abaixo (GET /auth/me).

Credenciais inválidas retornam `401 Unauthorized`.

#### `GET /auth/me`

Retorna os dados do usuário associado ao token enviado.

**Header obrigatório:**

``` http
Authorization: Bearer <jwt>
```

**Resposta `200 OK`:**

``` json
{
  "id": 1,
  "username": "usuario"
}
```

### Usuários

> **Atenção:** atualmente, somente `/auth/me` está protegido pelo
> middleware JWT. Os endpoints `/users` não exigem token na
> implementação atual.

#### `POST /users`

Cadastra um usuário.

**Requisição:**

``` json
{
  "username": "usuario",
  "password": "senha"
}
```

**Resposta `201 Created`:**

``` json
{
  "id": 1,
  "username": "usuario"
}
```

O username deve ser único. A senha é armazenada com hash bcrypt e deve
ter no máximo 72 bytes. Se o username já estiver cadastrado, a resposta
é `409 Conflict`.

#### `GET /users`

Lista os usuários. Retorna apenas `id` e `username`.

#### `GET /users/{id}`

Consulta um usuário pelo ID. Retorna `404 Not Found` se ele não existir.

#### `PUT /users/{id}`

Atualiza o username e, opcionalmente, a senha.

**Requisição:**

``` json
{
  "username": "novo_usuario",
  "password": "nova_senha"
}
```

O campo `password` é opcional. Se enviado, não pode ser vazio e deve ter
no máximo 72 bytes.

**Resposta `200 OK`:**

``` json
{
  "id": 1,
  "username": "novo_usuario"
}
```

#### `DELETE /users/{id}`

Exclui o usuário. Retorna `204 No Content` em caso de sucesso ou
`404 Not Found` se ele não existir.

## Eventos publicados

O serviço publica eventos na exchange RabbitMQ:

``` text
ecopulse.events
```

**Tipo da exchange:** `topic`

Eventos publicados:

  Ação                     `event_type`       `routing_key`
  ------------------------ ------------------ -------------------
  Cadastro de usuário      `UserRegistered`   `user.registered`
  Atualização de usuário   `UserUpdated`      `user.updated`
  Exclusão de usuário      `UserDeleted`      `user.deleted`

### Estrutura da mensagem

Todas as mensagens seguem o mesmo padrão:

``` json
{
  "event_id": "uuid",
  "event_type": "UserRegistered",
  "version": 1,
  "occurred_at": "2026-01-01T12:00:00Z",
  "data": {
    "user_id": 1,
    "username": "usuario"
  }
}
```

O conteúdo de `data` varia conforme o evento:

-   `UserRegistered`: `user_id`, `username`.
-   `UserUpdated`: `user_id`, `username`, `password_changed` (booleano).
-   `UserDeleted`: `user_id`, `username`.

**Importante para os consumidores:**

-   O serviço registra o evento na tabela Outbox dentro da mesma
    transação da operação do usuário.
-   Um publicador em segundo plano envia os eventos pendentes ao
    RabbitMQ. A publicação é tentada a cada 3 segundos, em lotes de até
    50 eventos.
-   O serviço não cria as filas dos consumidores. Cada microsserviço
    consumidor deve declarar sua própria fila durável e vinculá-la à
    exchange `ecopulse.events` usando a routing key correspondente.
-   A entrega pode ocorrer mais de uma vez. Consumidores devem
    implementar idempotência usando `event_id` para evitar processar o
    mesmo evento repetidamente.
-   Se não houver fila vinculada à routing key, a mensagem não é
    considerada publicada e será tentada novamente.

## Observações de integração

-   O ID do usuário é disponibilizado nos eventos como `data.user_id`.
-   A senha nunca é incluída nas respostas HTTP nem nos eventos.
-   O JWT contém `user_id` e `username` nas claims. Os outros serviços
    podem validar o token usando o mesmo segredo e o algoritmo HS256, ou
    consultar o serviço de autenticação.
-   O segredo JWT precisa ser o mesmo entre o serviço que emite o token
    e qualquer serviço que faça sua validação.
