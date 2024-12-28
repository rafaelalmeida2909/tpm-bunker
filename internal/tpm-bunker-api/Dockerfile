FROM python:3.12.0

RUN apt-get update && \
    apt-get install -y curl vim

ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1

WORKDIR /django-app/

COPY ./src ./src
COPY run.py .
COPY requirements.txt .

RUN pip install --upgrade pip
RUN pip install -r requirements.txt

CMD ["python", "run.py"]
