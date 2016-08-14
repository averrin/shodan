FROM ubuntu
EXPOSE 80
WORKDIR /app
COPY shodan /app/
COPY config.yaml /app/
