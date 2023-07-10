# Base image for NGINX and PHP-FPM
FROM nginx:1.25

# Install required packages for PHP
RUN apt-get update && apt-get install -y \
    php-fpm \
    php-mysql \
    php-mbstring \
    php-xml \
    php-gd \
    php-zip \
    wget \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Install Go
ENV GOLANG_VERSION 1.17
RUN wget -O go.tgz "https://golang.org/dl/go$GOLANG_VERSION.linux-amd64.tar.gz" \
    && tar -C /usr/local -xzf go.tgz \
    && rm go.tgz

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

# Copy Go microservice source code
COPY go-microservice /go/src/go-microservice

# Build Go microservice
RUN cd /go/src/go-microservice \
    && go build -o /go/bin/go-microservice

# Copy NGINX configuration file
COPY nginx.conf /etc/nginx/nginx.conf

# Copy PHP configuration file
COPY php-fpm.conf /etc/php/8.2/fpm/php-fpm.conf
COPY www.conf /etc/php/8.2/fpm/pool.d/www.conf

# Copy index.php
COPY index.php /usr/share/nginx/html/index.php
#CMD chown -R www-data:www-data  /usr/share/nginx/html/
# Expose ports
EXPOSE 80

ENV REQ_PER_SEC=2

# Start NGINX and PHP-FPM
CMD service php8.2-fpm start && service nginx start && ./go/bin/go-microservice