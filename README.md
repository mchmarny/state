[![](https://github.com/mchmarny/state/actions/workflows/qualify.yaml/badge.svg?branch=main)](https://github.com/mchmarny/state/actions/workflows/qualify.yaml)
[![codecov](https://codecov.io/gh/mchmarny/state/graph/badge.svg?token=LVXW9OXHYZ)](https://codecov.io/gh/mchmarny/state)
[![version](https://img.shields.io/github/release/mchmarny/state.svg?label=version)](https://github.com/mchmarny/state/releases/latest)
[![](https://img.shields.io/github/go-mod/go-version/mchmarny/state.svg?label=go)](https://github.com/mchmarny/state)
[![](https://goreportcard.com/badge/github.com/mchmarny/state)](https://goreportcard.com/report/github.com/mchmarny/state)
[![](https://img.shields.io/badge/License-Apache%202.0-blue.svg?label=license)](https://github.com/mchmarny/state/blob/main/LICENSE)

# state

Simple local state persistence of Go structs. 

Support: 
* Configurable serialization (JSON, YAML, Binary)
* Custom opt-in persistence annotation (`state`) 
* Thread safe

## usage example

* [simple](examples/simple/main.go)
* [annotations](examples/annotations/main.go)

## disclaimer

This is my personal project and it does not represent my employer. While I do my best to ensure that everything works, I take no responsibility for issues caused by this code.