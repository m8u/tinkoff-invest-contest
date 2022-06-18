FROM postgres:14.3-alpine3.16

ENV POSTGRES_USER postgres
ENV POSTGRES_PASSWORD postgres

COPY ./build/db/init.sql /docker-entrypoint-initdb.d/
