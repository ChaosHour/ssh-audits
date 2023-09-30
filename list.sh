#! /usr/bin/env zsh
# set -xv


# declare an array of hosts and their corresponding private key and IP address
declare -A L1=( ["primary"]="/Users/klarsen/projects/data-sync/.vagrant/machines/primary/virtualbox/private_key" ["replica"]="/Users/klarsen/projects/data-sync/.vagrant/machines/replica/virtualbox/private_key" ["etlreplica"]="/Users/klarsen/projects/data-sync/.vagrant/machines/etlreplica/virtualbox/private_key" ["proxysql"]="/Users/klarsen/projects/data-sync/.vagrant/machines/proxysql/virtualbox/private_key" ["orchestrator"]="/Users/klarsen/projects/data-sync/.vagrant/machines/orchestrator/virtualbox/private_key" )
declare -A L2=( ["primary"]="192.168.50.152" ["replica"]="192.168.50.153" ["etlreplica"]="192.168.50.154" ["proxysql"]="192.168.50.150" ["orchestrator"]="192.168.50.151")


# Usage
if [[ $# -gt 1 ]]; then
  echo "Usage: $0 [primary|replica|etlreplica|proxysql|orchestrator]"
  exit 1
fi

# check if an argument was passed to the script
if [[ -n $1 ]]; then
  # use the argument to access the corresponding values in L1 and L2
  ssh -i ${L1[$1]} klarsen@${L2[$1]}
else
  # iterate over the keys of L1 and use them to access the values in L1 and L2
  for host in ${(k)L1}
  do
    # connect to the host
    ssh -i ${L1[$host]} klarsen@${L2[$host]}
  done
fi
