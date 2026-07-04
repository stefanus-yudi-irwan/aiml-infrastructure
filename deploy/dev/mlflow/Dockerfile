FROM python:3.10-slim

RUN apt update && apt upgrade -y && apt install -y curl
RUN pip install mlflow[extras] psycopg2-binary boto3 cryptography pymysql

EXPOSE 5000