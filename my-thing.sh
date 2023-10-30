#! /usr/bin/env bash

mysql -t -e "SELECT SUBSTRING_INDEX(host, ':', 1) AS host_short, GROUP_CONCAT(DISTINCT USER) AS users, COUNT(*) FROM information_schema.processlist GROUP BY host_short UNION ALL SELECT 'total', '', count(*) as TOTAL FROM information_schema.processlist ORDER BY 3, 2"

#mysql -t -e "select @@read_only"

#mysqladmin extended-status | grep -wi 'threads_connected\|threads_running' | awk '{ print $2,$4}'

#pt-show-grants

#mysql -t -e "select user,host,password from mysql.user"

#mysql -vv -e "UPDATE mysql.user SET password=concat(substr(password,1,1),reverse(substr(password,2))) WHERE user NOT IN ('root','replication','monitor','dba_util','pt_heartbeat','gds_mha'); FLUSH PRIVILEGES"

#mysql -t -e "select user,host,password from mysql.user"

#mysql -NBe "SELECT CONCAT('KILL ', id, ';') FROM information_schema.processlist WHERE user not in ('replication','root','System user','tungsten','pt_heartbeat')" | mysql -vv