#!/bin/bash
./wait-for-it.sh cb-dragonfly-influxdb:8086 -t 10 -- ./runMyapp
