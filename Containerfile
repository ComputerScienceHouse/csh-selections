FROM docker.io/python:3.13-bookworm

WORKDIR /src/
COPY ./ .
RUN pip install -r requirements.txt
RUN git config --system --add safe.directory /src

ENTRYPOINT [ "gunicorn", "selections:app", "--bind=0.0.0.0:8080"]
