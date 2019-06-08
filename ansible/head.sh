#!/usr/bin/env bash
ansible-playbook ./site.yml -i ./hosts --limit head
