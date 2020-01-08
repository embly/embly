function "encoder" {
  runtime = "rust"
  path    = "./encoder"
  sources = [
    "../../embly-rs",
  ]
}
gateway {
  type     = "http"
  port     = 8084
  function = "${function.encoder}"
}
