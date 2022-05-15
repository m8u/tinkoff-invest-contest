# tinkoff-invest-contest
Торговый робот для Tinkoff Invest Robot Contest (https://github.com/Tinkoff/invest-robot-contest)

# Сборка
Вам поднадобится компилятор Go версии 1.18 и новее
```
$ go build .
```

# Запуск
Перед запуском нужно указать свой токен(ы) Tinkoff Invest API через переменные окружения. Можно использовать файл с переменными окружения ```.env```:
```
SANDBOX_TOKEN=<ваш_токен_песочницы>
COMBAT_TOKEN=<ваш_токен_для_реальной_биржи>
```
Получить описание параметров и режимов работы бота:
```
$ ./tinkoff-invest-contest --help
```

# Описание торгового алгоритма
coming soon
