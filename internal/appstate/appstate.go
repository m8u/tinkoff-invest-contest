package appstate

import "sync"

var ShouldExit = false
var NoInternetConnection = false

var ExitActionsWG sync.WaitGroup
var PostExitActionsWG sync.WaitGroup
