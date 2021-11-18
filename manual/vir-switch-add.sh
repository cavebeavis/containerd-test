#!/bin/bash

ip netns add t1
ip netns add t2

ip link add dev v-net-0 type bridge
ip link set dev v-net-0 up

ip link add dev veth-t1 type veth peer name veth-t1-br
ip link set veth-t1 netns t1
ip link set veth-t1-br master v-net-0
ip link set dev veth-t1-br up

ip link add dev veth-t2 type veth peer name veth-t2-br
ip link set veth-t2 netns t2
ip link set veth-t2-br master v-net-0
ip link set dev veth-t2-br up

ip -n t1 addr add 192.168.15.1/24 dev veth-t1
ip -n t2 addr add 192.168.15.2/24 dev veth-t2

ip netns exec t1 ip link set dev lo up
ip netns exec t2 ip link set dev lo up

ip -n t1 link set dev veth-t1 up
ip -n t2 link set dev veth-t2 up