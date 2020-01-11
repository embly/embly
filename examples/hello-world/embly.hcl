function "hello" {
  runtime = "rust"
  path    = "./hello"
}

gateway {
  type = "http"
  port = 8080
  route "/" {
    function = "${function.hello}"
  }
}
