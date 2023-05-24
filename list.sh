#! /usr/bin/env bash

ssh -i /Users/klarsen/projects/data-sync/.vagrant/machines/primary/virtualbox/private_key klarsen@10.8.0.152
ssh -i /Users/klarsen/projects/data-sync/.vagrant/machines/replica/virtualbox/private_key klarsen@10.8.0.153
ssh -i /Users/klarsen/projects/data-sync/.vagrant/machines/etlreplica/virtualbox/private_key klarsen@10.8.0.154
ssh -i /Users/klarsen/projects/data-sync/.vagrant/machines/proxysql/virtualbox/private_key klarsen@10.8.0.150
ssh -i /Users/klarsen/projects/data-sync/.vagrant/machines/orchestrator/virtualbox/private_key klarsen@10.8.0.151
