# Documentação Validadção de Relatos

## Configurações

### PostgreSQL

As tabelas de relatos, inbox e outbox foram criadas conforme o arquivo /init-db/db.sql

A tabela relato no banco de dados local apresenta a seguinte estrutura:

- id: Id do relato
- usuario_id: O id do usuário que criou o relato.
- tipo: O tipo ou categoria do relato (ex: "trânsito", "tempestado", "alagamento").
- descricao: O texto detalhado explicando o que aconteceu.
- latitude: A coordenada geográfica norte/sul.
- longitude: A coordenada geográfica leste/oeste.
- votos_positivos: Quantidade de pessoas que confirmaram.
- votos_negativos: Quantidade de pessoas que rejeitaram.
- estado: A situação atual do relato (por padrão, começa como 'POSSIVEL_RELATO').
- criado_em: A data e hora exatas em que o relato foi registrado no banco de dados.
- proxima_avaliacao_em: O momento em que o relato deve ser revisado pelo sistema. Por padrão, é definido para 15 minutos após a criação.
- atualizado_em: A data e hora da última modificação feita no registro desse relato.

### Docker


## Rotas HTTP

Raiz /validacaorelatos

Rotas:

- Enviar uma cópia do relato:
    - POST /
    - Esta rota não precisa ser utilizada, uma vez que este microsserviço já recebe o evento EVENTS.REPORT.CREATED do microsserviço de relatos.
    - Esta rota espera receber o json:
        ```
        {
            "id": "relato_uuid",
            "usuario_id": "usuario_uuid",
            "tipo": "tipo",
            "descricao": "descricao",
            "latitude": 0.0,
            "longitude": 0.0
        }
        ```

- Buscar relato por id:
    - GET /:id
    - Exemplo de resposta em caso sucesso:
        ```
        {
            "id": "661211f1-921e-4100-9280-de1578f4dc50",
            "usuario_id": "661211f1-921e-4100-9280-de1578f4dc10",
            "tipo": "teste",
            "descricao": "descricao",
            "latitude": -15,
            "longitude": 30,
            "votos_positivos": 0,
            "votos_negativos": 0,
            "estado": "RELATO_VALIDADO",
            "criado_em": "2026-11-08T18:57:11.408232-03:00",
            "proxima_avaliacao_em": "2026-11-08T20:12:20.089912-03:00",
            "atualizado_em": "2026-11-08T19:12:20.089912-03:00"
        }
        ```

- Testar se usuário votou: 
    - GET /votou/:usuario_id/:relato_id
    - Exemplo de resposta:
        ```
        {
            "votou": true
        }
        ```

- Votar no relato:
    - POST /votar/:relato_id
    - Esta rota espera receber o json:
        ```
        {
            "positivo": true ou false
            "usuario_id": "usuario_uuid"
        }
        ```

## Eventos

A exchange do RabbitMQ onde este microsserviço opera é validacao_relatos_exchange

---

Eventos de inbox definidos no InboxConsumer.go. A fila de entrada é validacao_relatos_queue.

- validacaorelato.criar: Recebe um relato e armazena no banco de dados local. Estrutura do payload que deve receber:
{
    "id": "padrao_uuid",
    "usuario_id": "padrao_uuid",
    "tipo": "tipo",
    "descricao": "descricao",
    "latitude": 0.0,
    "longitude": 0.0
}

- validacaorelato.votar: Recebee o id do relato e se o voto é positivo ou negativo. Estrutura do payload que deve receber:
{
    "id": "padrao_uuid",
    "tipo": "positivo or negativo"
}

- validacaorelato.atualizar: Recebe os dados do relato para atualizar. Estrutura do payload que deve receber:
{
    "id": "padrao_uuid",
    "tipo": "tipo",
    "descricao": "descricao",
    "latitude": 0.0,
    "longitude": 0.0,
    "estado": "estado"
}

- validacaorelato.deletar: Recebe apenas o id do relato e apaga do banco de dados local. Estrutura do payload que deve receber:
{
    "id": "padrao_uuid"
}

- validacaorelato.buscar: Recebe o id do relato que deseja buscar e repondo com o relato na fila de resposta. Estrutura do payload que deve receber: 
{
    "id": "padrao_uuid"
}

---

Eventos de outbox definidos no RelatoService.go. A fila com as mensagem de resposta é validacao_relato_queue.

- validacaoresposta.criado: Disparado por CriarRelato

- validacaoresposta.votado: Disparado por VotarRelato

- validacaoresposta.atualizado: Disparado por AtualizarRelato

- validacaoresposta.deletado: Disparado por DeletarRelato

- validacaoresposta.relato: Disparado por BuscarRelatoPorId, responde com o json do registro do relato no formato:
{
  "id": "padrao_uuid",
  "usuario_id": "padrao_uuid",
  "tipo": "tipo",
  "descricao": "descricao",
  "latitude": 0.0,
  "longitude": 0.0,
  "votos_positivos": 0,
  "votos_negativos": 0,
  "estado": "estado",
  "criado_em": "data e hora com zona de tempo",
  "proxima_avaliacao_em": "data e hora com zona de tempo",
  "atualizado_em": "data e hora com zona de tempo"
}
