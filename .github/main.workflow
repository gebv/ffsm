workflow "New workflow" {
  on = "push"
  resolves = ["./.github/tests-go-1.12"]
}

action "./.github/tests-go-1.12" {
  uses = "./.github/tests-go-1.12"
  env = {
    HOME = "/usr/local/go/github.com/gebv/ffsm"
  }
}
