## Config

`make config` скопирует примеры конфигов в директорию data/config

## Logs

Graylog: http://127.0.0.1:7555/ (admin/admin)

System->Inputs, добавляем инпут типа GELF tcp, все значения по-умолчанию

## Metrics

Prometheus: http://127.0.0.1:9090/

Grafana: http://127.0.0.1:3000/ (admin/admin)

При первом логине в Графану она попросит установить новый пароль, ставим.

Заходим в шестеренку слева, выбираем Data sources, добавляем Prometheus, адрес `http://prometheus:9090`

## Tracing

Jaeger: http://127.0.0.1:16686/
