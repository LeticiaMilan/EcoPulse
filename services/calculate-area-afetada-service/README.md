# Serviço de Cálculo de Área Afetada

## Responsabilidade

O serviço agrupa relatos validados que estão próximos no espaço e no
tempo e calcula a área afetada por eles. Quando relatos de usuários
diferentes confirmam a mesma ocorrência, a área passa a ser publicada
para os outros serviços, que podem exibi-la no mapa. O serviço também
pergunta periodicamente se a ocorrência continua, desmarca a área
quando os usuários respondem que ela acabou e expira áreas sem
atividade.

Ele utiliza PostgreSQL para persistência e RabbitMQ para receber e
publicar eventos, com os padrões Inbox e Outbox.

**Porta:** nenhuma. O serviço não expõe HTTP; toda a comunicação é feita
por mensagens.

## Como executar

Na pasta `services/calculate-area-afetada-service`, execute:

``` bash
docker compose up -d --build
```

O `docker-compose.yml` dessa pasta sobe o PostgreSQL, o RabbitMQ e o
serviço. Para verificar os containers:

``` bash
docker compose ps
```

Para acompanhar os logs do serviço:

``` bash
docker compose logs -f calculate-area-service
```

Para parar os serviços:

``` bash
docker compose down
```

O painel do RabbitMQ fica em `http://localhost:15672` (usuário e senha:
`ecopulse`). Por ele é possível publicar mensagens de teste na exchange
e ver as filas.

Variáveis de ambiente:

``` env
DATABASE_URL=postgresql://ecopulse:ecopulse@postgres:5432/calculate_area
RABBITMQ_URL=amqp://ecopulse:ecopulse@rabbitmq:5672/
RABBITMQ_EXCHANGE=exchange.ecopulse
RABBITMQ_QUEUE=queue.service.affected-area
POLL_INTERVAL_SECONDS=1
```

Os valores acima são de desenvolvimento. Só `DATABASE_URL` e
`RABBITMQ_URL` são definidas no `docker-compose.yml`; as outras usam o
valor padrão mostrado. Sem nenhuma variável, o serviço tenta se conectar
em `localhost`, o que permite rodá-lo fora do Docker com
`python src/main.py`.

O serviço cria automaticamente as tabelas `affected_areas`,
`area_reports`, `area_votes`, `inbox` e `outbox` ao iniciar.

## Padrão das mensagens

Todas as mensagens, recebidas ou publicadas, usam o envelope padrão do
EcoPulse:

``` json
{
  "id": "uuid",
  "source": "SERVICE.ORIGEM",
  "contentType": "application/json",
  "data": { }
}
```

O tipo do evento não vai no corpo da mensagem: ele é a **routing key**.
Por isso, quem publica para este serviço precisa usar exatamente as
routing keys listadas abaixo.

**Exchange:** `exchange.ecopulse`

**Tipo da exchange:** `topic` (durável)

## Eventos recebidos

O serviço declara a fila durável `queue.service.affected-area` e a
vincula à exchange com as routing keys abaixo.

  Evento                                  `routing_key`                Origem
  --------------------------------------- ---------------------------- -------------------------------------------
  Relato validado                         `EVENTS.REPORT.VALIDATED`    report-service
  Resposta à pergunta "ainda ocorre?"     `EVENTS.AFFECTED_AREA.VOTE`  quem coleta as respostas dos usuários

### `EVENTS.REPORT.VALIDATED`

Um relato que já passou pela validação. É a partir dele que as áreas
são criadas e crescem.

``` json
{
  "id": "3a8d1e05-6f2c-4b91-8840-d9e2358b1c47",
  "source": "SERVICE.REPORT",
  "contentType": "application/json",
  "data": {
    "id": "8f62b19e-5c40-42ae-bb73-936d5e1a7b80",
    "user_id": "c19b43d7-e218-450f-a396-745812fe69b2",
    "type": "Flooding",
    "latitude": -21.1775,
    "longitude": -47.8103,
    "datetime": "2026-10-31T15:30:00Z"
  }
}
```

Campos de `data`:

-   `id`: UUID do relato.
-   `user_id`: UUID do usuário que fez o relato.
-   `type`: um destes valores: `Landslide`, `Flooding`, `Thunderstorm`,
    `Fire Disaster`.
-   `latitude` e `longitude`: coordenadas do relato, em graus.
-   `datetime`: data e hora da ocorrência em ISO 8601. Sem fuso horário,
    é tratada como UTC.

Um relato que já entrou em alguma área, mesmo que ela já tenha sido
encerrada, é ignorado se chegar de novo.

### `EVENTS.AFFECTED_AREA.VOTE`

A resposta de um usuário a um evento
`EVENTS.AFFECTED_AREA.CONFIRMATION_REQUESTED`.

``` json
{
  "id": "b7c1f0a2-4e3d-4c6b-9a8e-1f2d3c4b5a69",
  "source": "SERVICE.ORIGEM",
  "contentType": "application/json",
  "data": {
    "area_id": "45f229ff-1127-4dbc-9b5b-318f8eccc9eb",
    "user_id": "c19b43d7-e218-450f-a396-745812fe69b2",
    "still_happening": false
  }
}
```

Campos de `data`:

-   `area_id`: UUID da área, recebido em `data.area_id` dos eventos
    publicados por este serviço.
-   `user_id`: UUID do usuário que respondeu.
-   `still_happening`: `true` se a ocorrência continua, `false` se
    acabou.

Vale só a resposta mais recente de cada usuário em cada área. Votos para
uma área inexistente ou que não está `active` são ignorados.

### O que acontece com cada mensagem recebida

-   O Event Consumer grava a mensagem na tabela `inbox` e só então
    confirma (ack) para o RabbitMQ. O processamento acontece depois, no
    Inbox Consumer.
-   Mensagens repetidas são descartadas pelo campo `id` do envelope.
    Para que isso funcione, cada evento precisa ter um `id` único e o
    mesmo `id` em todas as tentativas de envio.
-   Uma mensagem que não é JSON, ou que não tem `id` ou `data`, é
    rejeitada e descartada sem ser gravada.
-   Se `data` estiver com campos faltando ou valores inválidos (UUID mal
    formado, `type` desconhecido, data inválida), o evento fica com
    status `REJECTED` na `inbox`, com o motivo na coluna `error`. Ele
    não é reprocessado.
-   Se o erro for de outro tipo (banco fora do ar, por exemplo), o
    evento é tentado novamente, até 5 vezes, e depois fica `REJECTED`.

## Eventos publicados

O serviço publica na mesma exchange `exchange.ecopulse`, com
`source` igual a `SERVICE.AFFECTED_AREA`.

  Ação                                              `routing_key`
  ------------------------------------------------- ----------------------------------------------
  Área confirmada pela primeira vez                 `EVENTS.AFFECTED_AREA.DETECTED`
  Novo relato entrou em uma área ativa              `EVENTS.AFFECTED_AREA.UPDATED`
  Pedido para os usuários confirmarem a ocorrência  `EVENTS.AFFECTED_AREA.CONFIRMATION_REQUESTED`
  Área desmarcada pelos votos dos usuários          `EVENTS.AFFECTED_AREA.RESOLVED`
  Área encerrada por falta de atividade             `EVENTS.AFFECTED_AREA.EXPIRED`

### Estrutura da mensagem

Todos os eventos levam o estado atual completo da área em `data`:

``` json
{
  "id": "2f76aaed-109a-4542-9734-2eb4a20c8cd3",
  "source": "SERVICE.AFFECTED_AREA",
  "contentType": "application/json",
  "data": {
    "area_id": "45f229ff-1127-4dbc-9b5b-318f8eccc9eb",
    "type": "Flooding",
    "status": "active",
    "area_m2": 58598.47,
    "center": {
      "latitude": -21.177542928654507,
      "longitude": -47.81039045186131
    },
    "polygon": {
      "type": "Polygon",
      "coordinates": [
        [
          [-47.80915977255845, -21.1775609594799],
          [-47.809118275766075, -21.17763871413161],
          [-47.80908501840699, -21.177719844912232],
          [-47.80915977255845, -21.1775609594799]
        ]
      ]
    },
    "reports": 3,
    "users": 3
  }
}
```

O polígono do exemplo foi encurtado; o real tem dezenas de pontos.

Campos de `data`:

-   `area_id`: UUID da área. É o mesmo em todos os eventos da mesma
    área e é o valor a ser enviado no voto.
-   `type`: tipo da ocorrência, igual ao `type` dos relatos.
-   `status`: `active`, `resolved` ou `expired`.
-   `area_m2`: área do polígono em metros quadrados, com duas casas
    decimais.
-   `center`: centro do polígono.
-   `polygon`: polígono no formato GeoJSON. **As coordenadas seguem a
    ordem do GeoJSON: `[longitude, latitude]`.**
-   `reports`: quantidade de relatos na área.
-   `users`: quantidade de usuários diferentes que relataram.

Além do corpo, cada mensagem tem estas propriedades AMQP: `message_id`
igual ao `id` do envelope, `type` igual à routing key,
`content_type` igual a `application/json` e entrega persistente.

## Regras de cálculo da área

Os valores ficam em `src/core/rules.py`.

  Regra                                                         Valor
  ------------------------------------------------------------- --------------------------
  Usuários diferentes para a área ser detectada                 3
  Distância máxima até algum relato da área                     500 m
  Diferença máxima de tempo para o último relato da área        6 horas
  Margem ao redor dos relatos no polígono                       100 m
  Tempo sem atividade para a área expirar                       24 horas
  Intervalo entre pedidos de confirmação                        10 segundos (testes)
  Votos "não" para desmarcar a área                             3

Como o ciclo funciona:

1.  Um relato entra na área aberta mais próxima do mesmo `type` que
    esteja a até 500 m e a até 6 horas do último relato dela. Se não
    houver nenhuma, ele começa uma área nova com status `candidate`.
2.  Áreas `candidate` não são publicadas. Quando 3 usuários diferentes
    relatam na mesma área, ela vira `active` e o serviço publica
    `DETECTED`. Cada relato novo depois disso gera `UPDATED`.
3.  O polígono é o contorno que envolve todos os relatos da área, com
    100 m de margem ao redor, calculado na projeção SIRGAS 2000 / Brazil
    Polyconic (EPSG:5880).
4.  Enquanto a área está `active`, o serviço publica
    `CONFIRMATION_REQUESTED` a cada intervalo de confirmação. O
    intervalo está em 10 segundos para testes; em produção a ideia é
    usar 10 minutos.
5.  A área vira `resolved` e gera `RESOLVED` quando há pelo menos 3
    votos "não" e eles superam os votos "sim". Um voto "sim" conta como
    atividade e adia a expiração.
6.  Uma área sem relatos novos nem votos "sim" por 24 horas vira
    `expired`. Se ela estava `active`, o serviço publica `EXPIRED`; se
    ainda era `candidate`, ela expira sem gerar evento.

**Importante para os consumidores:**

-   O serviço grava o evento na tabela `outbox` dentro da mesma
    transação que altera a área.
-   O Outbox Dispatcher procura eventos pendentes a cada
    `POLL_INTERVAL_SECONDS` (1 segundo por padrão) e os envia em lotes
    de até 50.
-   O serviço não cria as filas dos consumidores. Cada microsserviço
    consumidor deve declarar sua própria fila durável e vinculá-la à
    exchange `exchange.ecopulse` com as routing keys que lhe interessam
    (por exemplo, `EVENTS.AFFECTED_AREA.*` para receber todos). A fila
    precisa existir antes da publicação; o RabbitMQ não guarda mensagens
    para filas criadas depois.
-   A entrega pode ocorrer mais de uma vez. Consumidores devem
    implementar idempotência usando o `id` do envelope.
-   Como cada evento traz o estado completo da área, o consumidor pode
    simplesmente substituir o que tem guardado para aquele `area_id`.

## Observações de integração

-   O ID da área é disponibilizado nos eventos como `data.area_id`.
-   Para remover uma área do mapa, trate `RESOLVED` e `EXPIRED`.
-   Áreas `candidate` existem só dentro deste serviço; nenhum evento é
    publicado para elas.
-   Os relatos individuais e os IDs dos usuários não são incluídos nos
    eventos publicados, apenas as contagens `reports` e `users`.
