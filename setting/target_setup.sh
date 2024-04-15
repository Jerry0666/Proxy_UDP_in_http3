#!/usr/bin/env bash
ifconfig target-eth0 200.0.0.1
ip route add default dev target-eth0