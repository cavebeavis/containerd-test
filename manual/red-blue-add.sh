#!/bin/bash

ip netns add red
ip netns add blue

ip link add veth-red type veth peer name veth-blue

ip link set veth-red netns red
ip link set veth-blue netns blue

ip -n red addr add 192.168.14.1 dev veth-red
ip -n blue addr add 192.168.14.2 dev veth-blue

ip -n red link set veth-red up
ip -n blue link set veth-blue up