FROM ubuntu
RUN echo "Europe/Moscow" > /etc/timezone && dpkg-reconfigure -f noninteractive tzdata
EXPOSE 80
EXPOSE 443
EXPOSE 8099
WORKDIR /app
COPY shodan /app/
COPY config.yaml /app/
