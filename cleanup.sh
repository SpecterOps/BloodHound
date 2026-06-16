#!/bin/bash
just bh-dev down
just bh-debug down
mapfile -t containers < <(docker container ls -q | grep "^bloodhound-dev")
if [ ${#containers[@]} -gt 0 ]; then
	docker container stop "${containers[@]}"
fi

mapfile -t volumes < <(docker volume ls -q | grep "^bloodhound-dev")
if [ ${#volumes[@]} -gt 0 ]; then
	docker volume rm "${volumes[@]}"
fi