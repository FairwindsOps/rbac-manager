FROM python:3.6.4-alpine3.7

WORKDIR /rbac-manager

COPY . .

RUN pip install -r requirements.txt

CMD python manage-rbac.py
