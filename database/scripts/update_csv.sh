#!/bin/bash

# gets script dir
SRCDIR=$(dirname $(readlink -f "$0"))
DBDIR="$(dirname "$SRCDIR")"

## create file csv
CSVFILE=${DBDIR}/summary-file.csv
echo '"id","name","category","tags","resources","web"' > $CSVFILE
jq -s '[.[][]]' ${DBDIR}/file/*.json | jq -s '.[][]' \
	| jq -r '{ id: .id, name: .name, category: .category, tags: (.tags // empty) |join(";") , resources: .resources|join(";"), web: .web}' \
	| jq -r '[.id, .name, .category, .tags, .resources, .web] | @csv' >> $CSVFILE


## create dnsxl csv
CSVFILE=${DBDIR}/summary-dnsxl.csv
echo '"id","name","category","tags","resources","web"' > $CSVFILE
jq -s '[.[][]]' ${DBDIR}/dnsxl/*.json | jq -s '.[][]' \
	| jq -r '{ id: .id, name: .name, category: .category, tags: (.tags // empty) |join(";") , resources: .resources|join(";"), web: .web}' \
	| jq -r '[.id, .name, .category, .tags, .resources, .web] | @csv' >> $CSVFILE

