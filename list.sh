#! /usr/bin/env zsh
# set -xv
################################################################################################################################################
# ssh -i /Users/klarsen/projects/data-sync/.vagrant/machines/primary/virtualbox/private_key klarsen@192.168.50.152
# ssh -i /Users/klarsen/projects/data-sync/.vagrant/machines/replica/virtualbox/private_key klarsen@192.168.50.153
# ssh -i /Users/klarsen/projects/data-sync/.vagrant/machines/etlreplica/virtualbox/private_key klarsen@192.168.50.154
# ssh -i /Users/klarsen/projects/data-sync/.vagrant/machines/proxysql/virtualbox/private_key klarsen@192.168.50.150
# ssh -i /Users/klarsen/projects/data-sync/.vagrant/machines/orchestrator/virtualbox/private_key klarsen@192.168.50.151
################################################################################################################################################

# declare an array of hosts and their corresponding private key and IP address
declare -A L1=( ["primary"]="/Users/klarsen/projects/data-sync/.vagrant/machines/primary/virtualbox/private_key" ["replica"]="/Users/klarsen/projects/data-sync/.vagrant/machines/replica/virtualbox/private_key" ["etlreplica"]="/Users/klarsen/projects/data-sync/.vagrant/machines/etlreplica/virtualbox/private_key" ["proxysql"]="/Users/klarsen/projects/data-sync/.vagrant/machines/proxysql/virtualbox/private_key" ["orchestrator"]="/Users/klarsen/projects/data-sync/.vagrant/machines/orchestrator/virtualbox/private_key" )
declare -A L2=( ["primary"]="192.168.50.152" ["replica"]="192.168.50.153" ["etlreplica"]="192.168.50.154" ["proxysql"]="192.168.50.150" ["orchestrator"]="192.168.50.151")


#for i in ${(k)L1}
#   do
#        echo "key  : $i"
#        echo "value: ${L1[$i]}"
#    done


#for i in ${(k)L2}
#   do
#        echo "key  : $i"
#        echo "value: ${L2[$i]}"
#    done


# iterate over the keys of L1 and use them to access the values in L1 and L2
#for host in ${(k)L1}
#do
#  echo "ssh -i ${L1[$host]} klarsen@${L2[$host]}"
#done


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
