example:
	go build -mod vendor
	(cd example && \
	INPUT_CREDS=.google_application_credentials.json \
	GITHUB_RUN_NUMBER=201 \
	GITHUB_SHA=13e82dd30df4e87118faa98712a5aebb0ab05c45 \
	../gce-deploy-action)

test:
	go test -mod vendor -v .


.PHONY: test example
