function "hello" {
  runtime = "rust"
  path    = "./hello"
}

gateway {
  type = "http"
  port = 8765
  route "/" {
    function = "${function.hello}"
  }
}
