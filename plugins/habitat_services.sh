#!/bin/bash

for service_group in $(curl --silent localhost:9631/services | jq -r '.[] | .service_group')
do
  service=$(echo -n ${service_group} | cut -d. -f1)
  group=$(echo -n ${service_group} | cut -d. -f2)
  status=$(curl --silent localhost:9631/services/${service}/${group}/health | jq -r '.status')
  echo -e "${service}\ts\t$status"
done
