#!/usr/bin/env bash
#set -xv

# The query to monitor
query_to_monitor="SELECT t.table_schema,
       t.table_name,
       COLUMN_NAME,
       AUTO_INCREMENT,
       pow(2, CASE data_type
                  WHEN 'tinyint' THEN 7
                  WHEN 'smallint' THEN 15
                  WHEN 'mediumint' THEN 23
                  WHEN 'int' THEN 31
                  WHEN 'bigint' THEN 63
              END+(column_type like '% unsigned'))-1 AS max_int
FROM information_schema.columns c
STRAIGHT_JOIN information_schema.tables t ON BINARY t.table_schema = c.table_schema
AND BINARY t.table_name = c.table_name
WHERE c.extra = 'auto_increment'
  AND t.auto_increment IS NOT NULL"


# The time period to monitor (in seconds)
monitoring_period=$((20 * 60))  # 20 minutes
#interval=300  # Check every 5 minutes
#interval=240  # Check every 4 minutes
# The time from when the query starts to when it complete.
interval=215  # Check every 3 minutes and 35 seconds

# The start time
start_time=$(date +%s)

# The count of the query
count=0

while true; do
  # Get the current time
  current_time=$(date +%s)

  # Break the loop if the monitoring period has passed
  if (( current_time - start_time >= monitoring_period )); then
    break
  fi

  # Get the current queries
  current_queries=$(mysql -e "SHOW FULL PROCESSLIST" | grep "$query_to_monitor")

  # Increment the count if the query is found
  if [[ -n "$current_queries" ]]; then
    ((count++))
  fi

  # Wait for the interval
  sleep "$interval"
done

echo "The query '$query_to_monitor' was found $count times in the last $((monitoring_period / 60)) minutes."