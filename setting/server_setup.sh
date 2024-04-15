#!/usr/bin/env bash
ifconfig server-eth0 100.0.0.1
ip route add default dev server-eth0