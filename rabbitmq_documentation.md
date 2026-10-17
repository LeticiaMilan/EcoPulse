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