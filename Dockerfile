FROM python:2.7.14-alpine3.7

WORKDIR /rbac-manager

COPY . .

RUN pip install -r requirements.txt

CMD python manage_rbac.py
