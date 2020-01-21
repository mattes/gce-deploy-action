example:
	go build
	INPUT_DIR=example \
	INPUT_GOOGLE_APPLICATION_CREDENTIALS=.google_application_credentials.json \
	GITHUB_RUN_NUMBER=48 \
	GITHUB_SHA=13e82dd30df4e87118faa98712a5aebb0ab05c45 \
	./gce-deploy-action

github-action:
	docker build -t mattes/gce-deploy-action .
	docker run --rm -it \
		-v $(PWD)/example:/github/workspace -w /github/workspace \
		-e INPUT_GOOGLE_APPLICATION_CREDENTIALS=.google_application_credentials.json \
		-e GITHUB_RUN_NUMBER=1 \
		-e GITHUB_SHA=13e82dd30df4e87118faa98712a5aebb0ab05c45 \
		mattes/gce-deploy-action

test:
	go test -mod vendor -v .


.PHONY: test example github-action
