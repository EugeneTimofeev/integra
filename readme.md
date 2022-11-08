# Пример интеграционного сервиса.

Сервис представляет из себя 1 исполняемый файл и файл конфигурации config.json.

Пример конфигурации:

{
  "Mode": "prod", 
  "Server": "",
  "Database": "",
  "InputPort": "8000",
  "UseUsernameAndPassword": false,
  "Username": "",
  "Password": ""
}
Параметр Mode - задает уровень реакции на критические ошибки (нет связи с БД, нет возможности выполнить хранимку и тд). Если параметр равен dev - приложение при критической ошибке логирует ошибку и закрывается, в противном случае только логирует ошибку.

Параметр Server - сервер БД

Параметр Database - БД

Параметр InputPort - порт, который слушает приложение (в данном случае 8000)

Параметр UseUsernameAndPassword - если для авторизации в БД нужны имя пользователя и пароль, устанавливается в true, username и password указываются в параметрах Username и Password соответственно. При запуске приложения из локальной сети предприятия можно указать false для windows-авторизации.

Приложение при старте начинает слушать порт, указанный в параметре InputPort. При получении POST-запроса с JSON-объектом типа:

{
  "CorrelationId": "00000000-0000-0000-0000-00000000000555",
  "PaymentDocumentID": "Строка 1 PaymentDocumentID",
  "PaymentOrderNumber": "Строка 2 PaymentOrderNumber"
}

 на точку вызова goexec (например, localhost:8000/goexec) приложение парсит его, создает XML-объект типа 
 
<test1C>
   <corID>00000000-0000-0000-0000-00000000000555</corID>
   <paymentDoc>Строка 1 PaymentDocumentID</paymentDoc>
   <paymentOrder>Строка 2 PaymentOrderNumber</paymentOrder>
</test1C>

и выполняет хранимку _tee_testProc

на сервере из параметра Server в БД из параметра Database, передавая ей в качестве входного параметра созданный XML-объект. Если все прошло без ошибок, приложение в респонсе отвечает строкой "ok", в противном случае строкой "error". Ошибки логируются в файле вида "20221019.log" и имеют вид:

2022/10/19 18:26:55 ERROR  mssql: login error: Login failed. The login is from an untrusted domain and cannot be used with Windows authentication.

ХП _tee_testProc_ на вход принимает XML-объект из примера выше, парсит его и вставляет в таблицу _tee_testTable_, не возвращая никаких данных.

# Запуск приложения в качестве Windows сервиса.

В командной строке из под администратора (или в командной строке с правами администратора) можно запустить приложение с параметрами.

go_exec_sp.exe -service-install - устанавливает приложение как сервис. Имя сервиса - golang-test-winsvc, описание сервиса - golang-test windows service;

go_exec_sp.exe -service-remove - удаляет сервис;

go_exec_sp.exe -service-start - запускает сервис (аналогично команде в оснастке Службы - Запустить службу)

go_exec_sp.exe -service-stop - останавливает работу сервиса (аналогично команде в оснастке Службы - Остановить службу).

Запуск приложения с параметром сопровождается ответом Done в случае исполнения команды корректно, или сообщением с текстом ошибки.

go_exec_sp.exe без параметров аналогичен запуску не как сервис, а как обычное приложение.



