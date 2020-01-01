function "main" {
  runtime = "rust"
  path    = "./main"
  sources = [
    "../../embly-rs",
  ]
}

gateway {
  type     = "http"
  port     = 8082
  function = "${function.main}"
}
