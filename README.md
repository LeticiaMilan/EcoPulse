# EcoPulse
Plataforma Distribuída Mobile para Monitoramento Colaborativo de Riscos Ambientais Urbanos e Qualidade do Ar.

## Como rodar o projeto

Guia para rodar tudo localmente no Windows (PowerShell): o backend no Docker e o app no celular pelo Expo Go.

### 1. Pré-requisitos

No PC:

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (aberto e rodando)
- [Node.js](https://nodejs.org/) (para o app)
- [Git](https://git-scm.com/)

No celular:

- App **Expo Go** ([Play Store](https://play.google.com/store/apps/details?id=host.exp.exponent) / [App Store](https://apps.apple.com/app/expo-go/id982107779))

PC e celular precisam estar na **mesma rede Wi-Fi**.

### 2. Subir o backend com Docker

Na raiz do repositório:

```powershell
docker compose up --build
```

Use `docker compose up --build -d` para rodar em segundo plano. Para parar:

```powershell
docker compose down
```

Isso sobe o RabbitMQ, os bancos Postgres e os serviços `login-service`, `report-service`, `validacao-relato`, `calculate-area-afetada-service` e `bff`.

Portas expostas no PC:

| Serviço | Porta no PC |
| --- | --- |
| login-service | 81 |
| report-service | 82 |
| validacao-relato | 83 |
| BFF (entrada do app) | 8090 |
| RabbitMQ (AMQP) | 5672 |
| RabbitMQ (painel) | 15672 |
| Postgres do login (`ecopulse_auth`) | 5432 |
| Postgres do report (`banco`) | 5433 |
| Postgres da área afetada (`calculate_area`) | 5434 |
| Postgres da validação (`aps_validacao`) | 5435 |

O `calculate-area-afetada-service` não expõe porta: ele só conversa pelo RabbitMQ.

Painel do RabbitMQ: http://localhost:15672 (usuário `ecopulse`, senha `ecopulse_dev`).

**Erros de "connection refused" no RabbitMQ** nos primeiros segundos da primeira subida são normais. Os serviços reiniciam e conectam sozinhos.

**Banco do validacao-relato sem tabelas.** Os scripts de `services/validacao-relato/init-db` só rodam quando o volume do banco está vazio. Se aparecer o erro `relation "outbox" does not exist`, apague só o volume desse banco:

```powershell
docker compose down
docker volume rm ecopulse_postgres_data_validacao
docker compose up --build
```

O prefixo `ecopulse_` vem do nome da pasta do repositório. Confira o nome exato com:

```powershell
docker volume ls
```

> **Não use `docker compose down -v`**: isso apaga os volumes de **todos** os bancos.

### 3. Configurar o app (antes do `npx expo start`)

1. Descubra o IP do PC na rede:

   ```powershell
   ipconfig
   ```

   Pegue o **Endereço IPv4** do adaptador Wi-Fi ou Ethernet. Não use o do WSL/vEthernet, nem `127.0.0.1`/`localhost`.

2. Abra `frontend/src/services/config.ts` e coloque esse IP na constante `HOST`:

   ```ts
   export const HOST = '192.168.0.10'; // troque pelo IP do seu PC
   ```

   O app usa o login na porta 81 e o BFF na porta 8090 desse IP.

3. Libere o firewall do Windows (necessário para o celular alcançar o PC). Faça isso **só em rede de casa/confiável**. Abra o PowerShell **como administrador** e veja o nome do adaptador:

   ```powershell
   Get-NetConnectionProfile
   ```

   Depois rode os comandos abaixo, trocando `"Ethernet"` pelo valor de `InterfaceAlias` que apareceu (pode ser `"Wi-Fi"`):

   ```powershell
   Set-NetConnectionProfile -InterfaceAlias "Ethernet" -NetworkCategory Private
   New-NetFirewallRule -DisplayName "EcoPulse dev" -Direction Inbound -Protocol TCP -LocalPort 8081,81,8090 -Action Allow -Profile Private
   ```

   A porta 8081 é a do Expo (Metro); 81 e 8090 são o login e o BFF.

   > Não marque como **Privada** uma rede pública (café, faculdade, aeroporto).

### 4. Rodar o app

```powershell
cd frontend
npm install
npx expo start
```

Leia o QR code com o Expo Go (Android) ou com a câmera (iPhone).

Para recarregar o app, aperte `r` no terminal do Expo.

**O QR code mostra o IP errado?** Force o IP certo e limpe o cache:

```powershell
$env:REACT_NATIVE_PACKAGER_HOSTNAME = "SEU_IP"
npx expo start -c
```

**Sem acesso de administrador?** Use o túnel:

```powershell
npx expo start --tunnel
```

O app carrega pelo túnel, mas o login e os relatos ainda precisam alcançar as portas 81 e 8090 do PC. Para usar de verdade, o firewall continua necessário.

### 5. Problemas comuns

| Sintoma | Causa provável |
| --- | --- |
| Expo Go fica carregando para sempre | O celular não alcança o PC: redes diferentes, firewall bloqueando ou IP errado |
| "Sem conexão com o servidor" no login | `HOST` errado em `config.ts` ou porta 81 bloqueada no firewall |
| `relation "outbox" does not exist` | Banco do validacao-relato sem tabelas (veja o passo 2) |

### 6. Regras do sistema (para quem for testar)

- **Validação do relato:** o relato é validado na hora quando chega a **5 confirmações** e tem mais confirmações que negações. Se não chegar a 5, depois de **15 minutos** vale a maioria (empate, inclusive 0 a 0, = rejeitado).
- **Área afetada:** aparece quando há **3 relatos validados do mesmo tipo**, de **3 contas diferentes**, a até **500 m** entre si, dentro de **6 horas**.
- **Encerramento da área:** precisa de pelo menos **3 votos de "já resolvido"** e mais votos de "já resolvido" do que de "ainda ativo".
- Quem criou o relato **não pode confirmar o próprio relato**.
- As atualizações levam de alguns segundos a **~30 s** para aparecer no app: os workers verificam as filas periodicamente e o app atualiza a cada 15 s.

### 7. Dica de teste: adiantar a avaliação de 15 minutos

Para não esperar os 15 minutos, marque todos os relatos pendentes para avaliação imediata:

```powershell
docker exec postgres_db psql -U postgres -d aps_validacao -c "UPDATE relatos SET proxima_avaliacao_em = NOW() WHERE estado = 'POSSIVEL_RELATO';"
```

A avaliação periódica roda a cada 1 minuto, então o resultado aparece em até ~1 minuto.
