FROM python:3.13-alpine

# Não gera arquivos .pyc e mostra os logs na hora (sem buffer)
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

WORKDIR /app

# Instala as dependências primeiro para aproveitar o cache do Docker
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copia o código do serviço
COPY src/ ./src/

# Roda como usuário sem privilégios de root
RUN adduser -D appuser
USER appuser

CMD ["python", "src/main.py"]
