package appstate

var ShouldExit = false
var NoInternetConnection = false

var ExitChan = make(chan bool)
