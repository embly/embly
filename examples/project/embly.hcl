
function "hello" {
  path    = "./hello"
  sources = ["../../embly-rs"]
  runtime = "rust"
}

function "echo" {
  path    = "./echo"
  sources = ["../../embly-rs"]
  runtime = "rust"
}

function "listener" {
  path    = "./listener"
  sources = ["../../embly-rs"]
  runtime = "rust"
}

gateway {
  type     = "http"
  port     = 8082
  function = "${function.listener}"
}
