# Documentação Validadção de Relatos

## Configurações

### PostgreSQL

As tabelas de relatos, inbox e outbox foram criadas conforme o arquivo /src/db.sql

### Docker


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
    "longitude": 0.0
}

- validacaorelato.deletar: Recebe apenas o id do relato e apaga do banco de dados local. Estrutura do payload que deve receber:

{
    "id": "padrao_uuid"
}

---

Eventos de outbox definidos no RelatoService.go. A fila com as mensagem de resposta é validacao_notificacoes_queue.

- validacaonotificacao.criado: disparado por CriarRelato

- validacaonotificacao.votado: disparado por VotarRelato

- validacaonotificacao.atualizado: disparado por AtualizarRelato

- validacaonotificacao.deletado: disparado por DeletarRelato

- validacaonotificacao.erro: 
