FROM ubuntu:18.04

MAINTAINER crycherd

RUN	apt-get -y update \
	&& apt-get -y install mysql-server mysql-client curl git

RUN	curl -O https://dl.google.com/go/go1.15.linux-amd64.tar.gz \
	&& tar xvf go1.15.linux-amd64.tar.gz \
	&& rm go1.15.linux-amd64.tar.gz \
	&& chown -R root:root ./go \
	&& mv go /usr/local 

RUN	git clone https://github.com/Kot2Na/Go_server.git \
	&& /etc/init.d/mysql start \
	&& cat /Go_server/db/db_creation.sql | mysql -u root mysql \
	&& /usr/local/go/bin/go get github.com/go-sql-driver/mysql 

ENTRYPOINT service mysql start \ 
	&& /usr/local/go/bin/go run Go_server/src/Server.go
