workflow "New workflow" {
  on = "push"
  resolves = ["./.github/tests-go-1.13"]
}

action "./.github/tests-go-1.13" {
  uses = "./.github/tests-go-1.13"
}
