# So You Want To Containerd??

**FYI**: `/plugins` comes from https://github.com/containernetworking/plugins and you will need to build the plugins and put them in a bin folder for use during interacting with the CNI tooling.

https://www.youtube.com/watch?v=6v_BDHIgOY8

https://www.youtube.com/watch?v=hDIcS66HpSk


https://jvns.ca/blog/2016/12/22/container-networking/

## TL;DR

All main packages need to be built and then run with `sudo` permissions...

Namespaces for network(s) and pid(s) are interwoven to provide virtual private networks for container networking to occur. See: https://unix.stackexchange.com/questions/396175/how-do-i-connect-a-veth-device-inside-an-anonymous-network-namespace-to-one-ou 

https://pkg.go.dev/github.com/containerd/containerd@v1.5.7#:~:text=the%20task%20is%20now%20running%20and%20has%20a%20pid%20that%20can%20be%20used%20to%20setup%20networking

```bash
# To see all control groups 'cgroups'
$ systemd-cgls
```

## CTR Commands If The Auto-Magic Approach Dies...

```bash
$ sudo ctr ns list
NAME    LABELS 
example        
moby           
t1

$ sudo ctr -n t1 containers ls
CONTAINER       IMAGE                             RUNTIME                  
redis-server    docker.io/library/redis:alpine    io.containerd.runc.v2

$ sudo ctr -n t1 tasks ls
TASK            PID      STATUS    
redis-server    56837    RUNNING

$ sudo kill -9 56837
# or just:
$ sudo ctr -n t1 tasks kill redis-server

$ sudo ctr -n t1 tasks rm redis-server
WARN[0000] task redis-server exit with non-zero exit code 137

$ sudo ctr -n t1 containers rm redis-server

$ sudo ctr -n t1 snapshots ls
KEY                                                                     PARENT                                                                  KIND      
sha256:00e1c01121e23eaa1f7224878570fdecc0d979c3994fbd8679e93f1c9c2291f9 sha256:400124b56a6a57fafcade7c358e691d340f280a5c0825d6daa5a7c9ae2be21c4 Committed 
sha256:1a058d5342cc722ad5439cacae4b2b4eedde51d8fe8800fcf28444302355c16d                                                                         Committed 
sha256:25dfa921bc4367acc24d1dfae1be4aad0b2774ea00f6caea4f14b67884c14622 sha256:2b2bb89d2b40e05e25f420b031ed2c3de43d80b880be58021244ce2d1d357cef Committed 
sha256:2b2bb89d2b40e05e25f420b031ed2c3de43d80b880be58021244ce2d1d357cef sha256:1a058d5342cc722ad5439cacae4b2b4eedde51d8fe8800fcf28444302355c16d Committed 
sha256:400124b56a6a57fafcade7c358e691d340f280a5c0825d6daa5a7c9ae2be21c4 sha256:25dfa921bc4367acc24d1dfae1be4aad0b2774ea00f6caea4f14b67884c14622 Committed 
sha256:c48287157d36be2a63b0c6a526287d68ce7a7c819c6af1c3382601def3b381f4 sha256:00e1c01121e23eaa1f7224878570fdecc0d979c3994fbd8679e93f1c9c2291f9 Committed

$ sudo ctr -n t1 snapshots rm sha256:00e1c01121e23eaa1f7224878570fdecc0d979c3994fbd8679e93f1c9c2291f9
ctr: failed to remove "sha256:00e1c01121e23eaa1f7224878570fdecc0d979c3994fbd8679e93f1c9c2291f9": cannot remove snapshot with child: failed precondition

$ sudo ctr -n t1 snapshots rm sha256:c48287157d36be2a63b0c6a526287d68ce7a7c819c6af1c3382601def3b381f4
$ sudo ctr -n t1 snapshots rm sha256:00e1c01121e23eaa1f7224878570fdecc0d979c3994fbd8679e93f1c9c2291f9
ctr: failed to remove "sha256:00e1c01121e23eaa1f7224878570fdecc0d979c3994fbd8679e93f1c9c2291f9": snapshot sha256:00e1c01121e23eaa1f7224878570fdecc0d979c3994fbd8679e93f1c9c2291f9 does not exist: not found
$ sudo ctr -n t1 snapshots ls
KEY PARENT KIND 
```

## Networking Commands

```bash
# To check for what namespaces exist
$ ip netns

# To add a namespace "testing"
$ sudo ip netns add testing

# To check the ip link for the main host
$ ip link

# To check the ip link within a container
$ sudo ip -n testing link
# - or -
$ sudo ip netns exec testing ip link

# To see ARP tables: ARP (Address Resolution Protocol) is the protocol
# that bridges Layer 2 and Layer 3 of the OSI model, which in the typical
# TCP/IP stack is effectively gluing together the Ethernet and Internet
# Protocol layers. This critical function allows for the discovery of a
# devices’ MAC (media access control) address based on its known IP address.
# On the Host:
$ arp

# Inside container network namespace:
$ sudo ip netns exec testing arp

# To see routing table on host:
$ route

# Inside container network namespace:
$ sudo ip netns exec testing route

# To delete namespace
$ sudo ip netns del testing



# To create a virtual cable with interfaces on each end (or pipe):
# First create namespaces
$ sudo ip netns add testing1
$ sudo ip netns add testing2
$ sudo ip link add veth-testing1 type veth peer name veth-testing2
$ sudo ip link set veth-testing1 netns testing1
$ sudo ip link set veth-testing2 netns testing2

# Assign ip addresses
$ sudo ip -n testing1 addr add 192.168.15.1 dev veth-testing1
$ sudo ip -n testing2 addr add 192.168.15.2 dev veth-testing2

# Set the links "UP"
$ sudo ip -n testing1 link set veth-testing1 up
$ sudo ip -n testing2 link set veth-testing2 up

# Ping to verify
$ sudo ip netns exec testing1 ping 192.168.15.2



### TO ADD VIRTUAL SWITCH
$ sudo ip link add v-net-0 type bridge
$ ip link
$ sudo ip link set dev v-net-0 up

# Create networks
$ sudo ip netns add t1
$ sudo ip netns add t2

# Wire up the virtual switch -- naming convention -br "bridge"
$ sudo ip link add veth-t1 type veth peer name veth-t1-br
$ sudo ip link set veth-t1 netns t1
$ sudo ip link set veth-t1-br master v-net-0

$ sudo ip link add veth-t2 type veth peer name veth-t2-br
$ sudo ip link set veth-t2 netns t2
$ sudo ip link set veth-t2-br master v-net-0

$ sudo ip -n t1 addr add 192.168.15.1/24 dev veth-t1
$ sudo ip -n t2 addr add 192.168.15.2/24 dev veth-t2

$ sudo ip -n t1 link set veth-t1 up
$ sudo ip -n t2 link set veth-t2 up

# To reach the containers from the host:
$ sudo ip addr add 192.168.15.5/24 dev v-net-0

# If we need to reach outside say 192.168.1.3:
$ sudo ip netns exec t2 ip route add 192.168.1.0/24 via 192.168.15.5

# For internal network traffic
$ sudo iptables -t nat -A POSTROUTING -s 192.168.15.0/24 -j MASQUERADE

# For internet traffic
$ sudo ip netns exec t2 ip route add default via 192.168.15.5

# Test
$ sudo ip netns exec t2 ping 8.8.8.8

# Port forwarding for outside network
$ sudo iptables -t nat -A PREROUTING --dport 80 --to-destination 192.168.15.2:80 -j DNAT

# TODO: add delete everything lol!
```

