#!/bin/bash

ip link set dev v-net-0 down

ip netns del t1
ip netns del t2

ip link del v-net-0