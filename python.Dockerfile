# syntax=docker/dockerfile:1

FROM python:3.10-slim-bullseye

WORKDIR /app

# Устанавливаем необходимые зависимости для сборки
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    ffmpeg \
    build-essential \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# Обновляем pip
RUN pip install --no-cache-dir --upgrade pip

# Устанавливаем Python-зависимости
RUN pip install --no-cache-dir numpy==1.22.0 \
                               scipy==1.11.2 \
                               torch \
                               grpcio \
                               grpcio-tools \
                               git+https://github.com/coqui-ai/TTS.git@dev

# Копируем серверный код
COPY main.py .

# Открываем порт gRPC
EXPOSE 50051

# Запускаем сервер
CMD ["python", "main.py"]
