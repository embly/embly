workflow "New workflow" {
  on       = "push"
  resolves = ["GitHub Action for Docker"]
}

action "GitHub Action for Docker" {
  uses = "actions/docker/cli@86ab5e854a74b50b7ed798a94d9b8ce175d8ba19"
  args = "build -f ./tools/go-test.Dockerfile ."
}
