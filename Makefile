example:
	docker build -t mattes/gce-deploy-action .
	docker run --rm -it \
		-v $(PWD)/example:/github/workspace -w /github/workspace \
		-e INPUT_GOOGLE_APPLICATION_CREDENTIALS=.google_application_credentials.json \
		mattes/gce-deploy-action

test:
	go test -mod vendor -v .


.PHONY: test example
