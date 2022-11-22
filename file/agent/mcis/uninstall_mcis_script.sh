#!/bin/bash

echo "[MCIS-Agent: Start to Delete Milkyway]"

echo "[MCIS-Agent: UnInstall tcpdump]"
sudo apt-get -y purge tcpdump

echo "[MCIS-Agent: Drop dump tables for evaluation]"
sudo mysql -u root -ppsetri1234ak -e "DROP USER 'sysbench'@'localhost';"
sudo mysql -u root -ppsetri1234ak -e "DROP DATABASE sysbench;"

echo "[MCIS-Agent: UnInstall MySQL]"
sudo apt-get -y purge mariadb-server

echo "[MCIS-Agent: UnInstall Ping]"
sudo apt-get purge -y iputils-ping

echo "[MCIS-Agent: UnInstall sysbench]"
sudo apt-get purge -y sysbench

echo "[MCIS-Agent: Remove Initial Setting]"
sudo apt-get purge -y dialog

echo "[MCIS-Agent: Deletion is done]"
