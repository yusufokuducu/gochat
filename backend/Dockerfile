FROM python:3.9-slim

WORKDIR /app

COPY requirements.txt .

RUN pip install --no-cache-dir -r requirements.txt && \
    apt-get update && \
    apt-get install -y netcat-openbsd && \
    apt-get clean

COPY . .

RUN chmod +x start.sh

CMD ["./start.sh"]
