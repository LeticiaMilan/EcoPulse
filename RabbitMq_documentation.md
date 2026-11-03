
# Especificação de Mensageria: Microsserviço Report

## 1. Visão Geral
Este documento descreve a interface assíncrona orientada a eventos para o **Microsserviço de Relatos (Report)** utilizando RabbitMQ. O microsserviço consome eventos de criação de relatórios de ocorrências e publica eventos resultantes de sucesso ou erros decorrentes de validações de regras de negócio.

---

## 2. Topologia RabbitMQ

| Parâmetro | Detalhes |
| :--- | :--- |
| **Exchange** | `exchange.ecopulse` |
| **Exchange Type** | `topic` |
| **Padrão de Envelope** | Baseado em especificações de envelope JSON (compatível com CloudEvents/Inbox Pattern) |
| **Codificação** | UTF-8 (`application/json`) |

---

## 3. Padrão do Envelope de Mensagem

Todas as mensagens transitadas pela exchange `exchange.ecopulse` seguem a estrutura padrão abaixo:

```go
type EventEnvelope struct {
    ID          string          `json:"id"`          // UUID v4 único da mensagem
    Source      string          `json:"source"`      // Identificador do serviço emissor
    ContentType string          `json:"contentType"` // MIME type (ex: application/json)
    Data        json.RawMessage `json:"data"`        // Payload contextual do evento
}
```

---

## 4. Dicionário de Eventos

### 4.1. Evento de Entrada: Criação de Relato Solicitada

Consumido pelo Microsserviço Report para validar, registrar e dar início ao ciclo de vida de uma ocorrência.

* **Routing Key:** `EVENTS.REPORT.GENERATED`

#### Estrutura do Payload (`data`)

| Campo | Tipo | Obrigatório | Descrição |
| :--- | :--- | :--- | :--- |
| `id` | `UUID` | Sim | Identificador primário da ocorrência gerada pelo cliente/emissor |
| `latitude` | `float64` | Sim | Coordenada de latitude do evento |
| `longitude` | `float64` | Sim | Coordenada de longitude do evento |
| `type` | `string` | Sim | Categoria da ocorrência (ex.: `"Flooding"`, `"Landslide"`) |
| `user_id` | `UUID` | Sim | Identificador do usuário que registrou a ocorrência |
| `area` | `string` | Não | Região, bairro ou zona da ocorrência |
| `datetime` | `string (RFC 3339)` | Sim | Data e hora em que a ocorrência foi reportada |

#### Exemplo de Mensagem

```json
{
  "id": "3a8d1e05-6f2c-4b91-8840-d9e2358b1c47",
  "source": "SERVICE.FRONTEND",
  "contentType": "application/json",
  "data": {
    "id": "8f62b19e-5c40-42ae-bb73-936d5e1a7b80",
    "latitude": -21.1775,
    "longitude": -47.8103,
    "type": "Flooding",
    "user_id": "c19b43d7-e218-450f-a396-745812fe69b2",
    "area": "Centro",
    "datetime": "2026-10-12T15:30:00Z"
  }
}
```

---

### 4.2. Evento de Saída: Sucesso (`EVENTS.REPORT.CREATED`)

Publicado pelo Microsserviço Report após a persistência e validação dos dados da ocorrência no banco de dados.

* **Routing Key:** `EVENTS.REPORT.CREATED`
* **Origem (`source`):** `SERVICE.REPORT`

#### Exemplo de Mensagem

```json
{
  "id": "45d9e2a4-d3cd-455e-a2ac-c13df4cdd7d5",
  "source": "SERVICE.REPORT",
  "contentType": "application/json",
  "data": {
    "id": "9c14f8d2-19e0-47cb-b461-26ea7849102c",
    "area": "Centro",
    "type": "Flooding",
    "status": "inProgress",
    "user_id": "f583e71d-3b7c-4029-9231-15b69a23c894",
    "datetime": "2026-10-12T15:30:00Z",
    "latitude": -21.1775,
    "longitude": -47.8103
  }
}
```

---

### 4.3. Evento de Saída: Falha de Negócio (`EVENTS.REPORT.GENERATED_ERROR`)

Publicado quando a mensagem recebida possui formato válido, mas viola regras de negócio do domínio (ex.: tipo de ocorrência desconhecido, coordenadas fora dos limites suportados, usuário invalido).

* **Routing Key:** `EVENTS.REPORT.GENERATED_ERROR`
* **Origem (`source`):** `SERVICE.REPORT`


#### Estrutura do Payload de Erro (`data`)

| Campo | Tipo | Descrição |
| :--- | :--- | :--- |
| `id` | `UUID` | ID da mensagem de origem que gerou o erro |
| `error` | `string` | Motivo do erro amigável/descritivo |
| `content` | `object` | Cópia exata dos dados recebidos que falharam na validação |
| `eventType` | `string` | Nome da chave de roteamento de entrada que originou a requisição |

#### Exemplo de Mensagem

```json
{
  "id": "cc261f99-7965-4561-9128-9b0f45c0ea40",
  "source": "SERVICE.REPORT",
  "contentType": "application/json",
  "data": {
    "id": "3a8d1e05-6f2c-4b91-8840-d9e2358b1c47",
    "error": "Tipo de relato inválido",
    "eventType": "EVENTS.REPORT.GENERATED",
    "content": {
      "id": "8f62b19e-5c40-42ae-bb73-936d5e1a7b80",
      "area": "Centro",
      "type": "Floodig",
      "status": "PENDING",
      "user_id": "c19b43d7-e218-450f-a396-745812fe69b2",
      "datetime": "2026-10-12T15:30:00Z",
      "latitude": -21.1775,
      "longitude": -47.8103
    }
  }
}
```

---

# Especificação de Mensageria: Microsserviço Report Validation

## 1. Visão Geral

Este documento descreve a para o Microsserviço Report Validation usando o RabbitMQ. O microsserviço consome eventos de criação de relatos, votação para validação e busca de relato com seu estado de validação e publica eventos de resposta.

Quando um relato é criado ele apresenta datas de criação, de próxima avaliação e a última vez que foi atualizado. Inicialmente o relato é publicado com estado "POSSIVEL_RELATO" e com a próxima avaliação para o horário 15 minutos à frente da data de criação. Se o número de votos positivos for maior que o negativo o estado é atualizado para "RELATO_VALIDADO" e a hora de avaliação é colocada para 1 hora depois da atualização, caso contrário será atualizado para "REJEITADO". Após a 1 hora que o relato foi validado ele retorna ao estado de "POSSIVEL_RELATO" e seus votos são reiniciados.

---

## 2. Topologia RabbitMQ

| Parâmetro | Detalhes |
| :--- | :--- |
| **Exchange** | `exchange.ecopulse` |
| **Exchange Type** | `topic` |
| **Codificação** | UTF-8 (`application/json`) |

---

## 3 Dicionário de Eventos

### 3.1 Evento de entrada: Criação de cópia local do Relato

Consumido para criar uma cópia do relato original, com seu id e outras informações, para registrar os votos.

* **Routing Key:** `validacaorelato.criar`

#### Estrutura do Payload

| Campo | Tipo | Obrigatório | Descrição |
| :--- | :--- | :--- | :--- |
| `id` | `UUID` | Sim | Identificador primário da ocorrência gerada pelo cliente/emissor |
| `tipo` | `string` | Sim | Categoria da ocorrência (ex.: `"Flooding"`, `"Landslide"`) |
| `descricao` | `string` | Sim | Descrição do relato |
| `usuario_id` | `UUID` | Sim | Identificador do usuário que registrou a ocorrência |
| `latitude` | `float64` | Sim | Coordenada de latitude do evento |
| `longitude` | `float64` | Sim | Coordenada de longitude do evento |


#### Exemplo de Mensagem

```
{
    "id": "661211f1-921e-4100-9280-de1578f4dc0f",
    "usuario_id": "d3b07384-d113-4ec2-a499-e82fa7ceef88",
    "tipo": "Flooding",
    "descricao": "Rua do centro está alagada devido a chuva",
    "latitude": -20.8956,
    "longitude": -48.2867
}
```

### 3.2 Evento de Saída: Relato Criado com Sucesso

Recebido como resposta para confirmar que a cópia do relato foi criada

* **Routing Key:** `validacaoresposta.criado`

### 3.3 Evento de Entrada: Votar no Relato para Validação

Consumido para realizar a votação para confirmar o relato. Este evento ira receber votos positivos e votos negativos. O primeiro são votos de confirmação do relato e, o segundo de rejeição.

* **Routing Key:** `validacaorelato.votar`

#### Estrutura do Payload

| Campo | Tipo | Obrigatório | Descrição |
| :--- | :--- | :--- | :--- |
| `id` | `UUID` | Sim | Identificador primário do relato que deseja vota |
| `tipo` | `string` | Sim | Tipo do voto (deve receber "positivo" ou "negativo") |

#### Exemplo de Mensagem

```
{
    "id": "661211f1-921e-4100-9280-de1578f4dc0f",
    "tipo": "positivo"
}
```

### 3.4 Evento de Saída: Voto confirmado

Resposta gerada pelo evento de votação para confirmar que o voto foi realizado

* **Routing Key:** `validacaoresposta.votado`

### 3.5 Evento de Entrada: Busca de relato por id

Consumido para buscar relato e confirmar seu estado.

* **Routing Key:** `validacaorelato.buscar`

#### Estrutura do Payload (`data`)

| Campo | Tipo | Obrigatório | Descrição |
| :--- | :--- | :--- | :--- |
| `id` | `UUID` | Sim | Identificador primário do relato que deseja buscar |

```
{
    "id": "661211f1-921e-4100-9280-de1578f4dc0f"
}
```

### 3.6 Evento de Saída: Dados do Relato Requisitado

Resposta com dados do relato requisitado. Os estados que um relato pode apresentar é:

- POSSIVEL_RELATO: Um relato ainda não confirmado;
- RELATO_VALIDADO: Relato que já foi confirmado pelos usuários como verdadeiro;
- REJEITADO: Marca que o relato é falso ou o acontecimento reportado não está mais acontecendo.

* **Routing Key:** `validacaorelato.relato`

#### Estrutura do Payload recebido

| Campo | Tipo | Descrição |
| :--- | :--- | :--- |
| `id` | `UUID` | Identificador primário da ocorrência gerada pelo cliente/emissor |
| `tipo` | `string` | Categoria da ocorrência (ex.: `"Flooding"`, `"Landslide"`) |
| `descricao` | `string` | Descrição do relato |
| `usuario_id` | `UUID` | Identificador do usuário que registrou a ocorrência |
| `latitude` | `float64`| Coordenada de latitude do evento |
| `longitude` | `float64` | Coordenada de longitude do evento |
| `votos_positivos` | `int` | Votos de confirmação do relato |
| `votos_negativos` | `int` | Votos de rejeição do relato |
| `estado` | `string` | Estado de validação do relato (POSSIVEL_RELATO, RELATO_VALIDADO e REJEITADO) |


#### Exemplo de Mensagem

```
{
  "id": "661211f1-921e-4100-9280-de1578f4dc0f",
  "usuario_id": "d3b07384-d113-4ec2-a499-e82fa7ceef88",
  "tipo": "Flooding",
  "descricao": "Rua do centro está alagada devido a chuva",
  "latitude": -20.8956,
  "longitude": -48.2867
  "estado": "RELATO_VALIDADO",
  "votos_positivos": 10,
  "votos_negativos": 1,
  "criado_em": "2026-11-02T11:08:54.059867-03:00",
  "proxima_avaliacao_em": "2026-11-02T12:20:23.159327-03:00",
  "atualizado_em": "2026-11-02T11:20:23.159327-03:00"
}
```
