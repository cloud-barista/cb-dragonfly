#!/bin/bash
./wait-for-it.sh $DRAGONFLY_INFLUXDB_URL -t 10 -- ./runMyapp
