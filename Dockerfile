FROM ubuntu
EXPOSE 80
EXPOSE 8099
WORKDIR /app
COPY shodan /app/
COPY config.yaml /app/
