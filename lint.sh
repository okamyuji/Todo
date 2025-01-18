#!/usr/bin/env bash

if [[ $# != 0 ]]; then
    echo '[用法]
./lint.sh

[オプション]
-h/--help    #このヘルプを表示する

[終了ステータス]
0: コマンド全体が成功したとき
1: そうでないとき

[仕様]
- `goimports`で検出されたエラーは自動修正されますが、`git add`は手動で必要です
  (終了ステータスの判定上は、自動修正が成功すれば「成功」と見なされます)

- そのほかのエラーは自動修正されないため、手動で修正が必要です

- `typos`でfalse positiveがあった場合、`_typos.toml`を編集してください

- `codespell`でfalse positiveがあった場合、`.codespellrc`を編集してください'

    exit 0
fi

function check_command_existence() {
    local commands=(
        goimports
        staticcheck
        golangci-lint
        typos
        codespell
    )
    local exit_status=0
    for command in "${commands[@]}"; do
        local does_command_exist=1
        command -v "${command}" > /dev/null || does_command_exist=0
        if [[ "${does_command_exist}" == 0 ]]; then
            echo "コマンド\`${command}\`がインストールされていません"
            exit_status=1
        fi
    done
    if [[ "${exit_status}" != 0 ]]; then
        exit "${exit_status}"
    fi
}
check_command_existence

function print_header() {
    echo "========== $1 =========="
}

exit_status=0

function tear_down() {
    if [[ $? != 0 ]]; then
        exit_status=1
    fi
    echo
}

print_header '[formatter] goimports -l -w .'
goimports -l -w .
tear_down

print_header '[linter] go vet ./...'
go vet ./...
tear_down

print_header '[linter] staticcheck ./...'
staticcheck ./...
tear_down

print_header '[linter] golangci-lint ./...'
golangci-lint run
tear_down

print_header '[spell checker] typos --format brief'
typos --format brief
tear_down

print_header '[spell checker] codespell'
codespell
tear_down

exit "${exit_status}"