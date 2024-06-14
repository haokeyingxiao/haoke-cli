#!/usr/bin/env bash

rm -rf completions
mkdir completions
go run . completion bash > completions/haoke-cli.bash
go run . completion zsh > completions/haoke-cli.zsh
go run . completion fish > completions/haoke-cli.fish