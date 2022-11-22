#!/bin/bash

echo "[MCIS-Agent: Start to prepare a VM evaluation]"

echo "[MCIS-Agent: Initial Setting]"
echo 'debconf debconf/frontend select Noninteractive' | sudo debconf-set-selections
sudo apt-get install -y -q

sudo apt-get update
sudo apt-get install -y locales locales-all
sudo locale-gen ko_KR.UTF-8

sudo apt-get install -y --no-install-recommends apt-utils
sudo apt-get -y install dialog

echo "[MCIS-Agent: Install sysbench]"
sudo apt-get -y install sysbench

echo "[MCIS-Agent: Install Ping]"
sudo apt-get -y install iputils-ping

echo "[MCIS-Agent: Set debconf]"
sudo debconf-set-selections <<< 'mariadb-server mysql-server/root_password password psetri1234ak'
sudo debconf-set-selections <<< 'mariadb-server mysql-server/root_password_again password psetri1234ak'

echo "[MCIS-Agent: Install MySQL]"
sudo apt-get -y install mariadb-server

echo "[MCIS-Agent: Generate dump tables for evaluation]"
sudo mysql -u root -ppsetri1234ak -e "CREATE DATABASE sysbench;"
sudo mysql -u root -ppsetri1234ak -e "CREATE USER 'sysbench'@'localhost' IDENTIFIED BY 'psetri1234ak';"
sudo mysql -u root -ppsetri1234ak -e "GRANT ALL PRIVILEGES ON *.* TO 'sysbench'@'localhost' IDENTIFIED  BY 'psetri1234ak';"

echo "[MCIS-Agent: Install tcpdump]"
sudo apt-get -y install tcpdump

echo "[MCIS-Agent: Preparation is done]"
