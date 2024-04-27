#!/bin/bash


../go-autotests/bin/gophermarttest \
            -test.v -test.run=^TestGophermart/TestEndToEnd$ \
            -gophermart-binary-path=cmd/gophermart/gophermart \
            -gophermart-host=localhost \
            -gophermart-port=8080 \
            -gophermart-database-uri="postgres://postgres:postgres@localhost:5432/gophermarket" \
            -accrual-binary-path=cmd/accrual/accrual_linux_amd64 \
            -accrual-host=localhost \
            -accrual-port=8081 \
            -accrual-database-uri="postgres://postgres:postgres@localhost:5432/gophermarket"



