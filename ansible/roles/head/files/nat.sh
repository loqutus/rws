#!/usr/bin/env bash
/sbin/iptables -t nat -A POSTROUTING -o enxdc9b9cee12cf -j MASQUERADE
/sbin/iptables -A FORWARD -i enxdc9b9cee12cf -o eth0 -m state --state RELATED,ESTABLISHED -j ACCEPT
/sbin/iptables -A FORWARD -i eth0 -o enxdc9b9cee12cf -j ACCEPT
