#!/bin/sh

# CB-Dragonfly Image
sudo docker build -t cloudbaristaorg/cb-dragonfly:0.7.0 . --no-cache
sudo docker push cloudbaristaorg/cb-dragonfly:0.7.0

# MCIS Collector Image
sudo docker build -t cloudbaristaorg/cb-dragonfly:0.7.0-mcis-collector -f Collector.MCIS.Dockerfile . --no-cache
sudo docker push cloudbaristaorg/cb-dragonfly:0.7.0-mcis-collector

# MCKS Collector Image
sudo docker build -t cloudbaristaorg/cb-dragonfly:0.7.0-mck8s-collector -f Collector.MCKS.Dockerfile . --no-cache
sudo docker push cloudbaristaorg/cb-dragonfly:0.7.0-mck8s-collector