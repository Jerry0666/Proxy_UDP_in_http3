#!/usr/bin/env bash
ip link add mp-bridge type bridge
ip link add client-eth0 type veth peer name br-veth0
ip link add client-eth1 type veth peer name br-veth1
ip link add server-eth0 type veth peer name br-veth2
ip link add target-eth0 type veth peer name br-veth3

ip netns add client
ip link set dev client-eth0 netns client
ip link set dev client-eth1 netns client
ip netns exec client ip link set lo up
ip netns exec client ip link set client-eth0 up
ip netns exec client ip link set client-eth1 up

ip netns add server
ip link set dev server-eth0 netns server
ip netns exec server ip link set lo up
ip netns exec server ip link set server-eth0 up

ip netns add target
ip link set dev target-eth0 netns target
ip netns exec target ip link set lo up
ip netns exec target ip link set target-eth0 up

ip link set dev br-veth0 master mp-bridge
ip link set dev br-veth1 master mp-bridge
ip link set dev br-veth2 master mp-bridge
ip link set br-veth0 up
ip link set br-veth1 up
ip link set br-veth2 up
ip link set mp-bridge up

ip link set dev br-veth3 master mp-bridge
ip link set br-veth3 up