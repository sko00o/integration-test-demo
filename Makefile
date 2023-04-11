.PHONY: test-%
test-%:
	cd $* && go test ./...

.PHONY: gitlab-runner
gitlab-runner:
	docker run --rm \
  --entrypoint bash \
  -w $(PWD) \
  -v $(PWD):$(PWD) \
  -v "/var/run/docker.sock:/var/run/docker.sock" \
  gitlab/gitlab-runner:latest \
  -c 'git config --global --add safe.directory "*"; gitlab-runner exec docker --docker-privileged test_app'
