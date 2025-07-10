.PHONY: manifest
manifest:
	helm template \
  		mysql oci://registry-1.docker.io/bitnamicharts/mysql --version 12.3.1 \
  		--namespace list-images-for-pods \
  		--create-namespace > manifest.yaml
