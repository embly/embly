
function "auth" {
  runtime = "rust"
  path    = "./auth"
  sources = [
    "../embly-rs",
  ]
}

gateway {
  type = "http"
  port = 8082

  route "/" {
    files = "${files.index}"
  }

  route "/app/" {
    files = "${files.frontend}"
  }
  route "/api/auth/" {
    function = "${function.auth}"
  }
  route "/blog/" {
    files = "${files.blog}"
  }

}

files "frontend" {
  path              = "./frontend/build/"
  local_file_server = "http://localhost:3000"
}


files "blog" {
  path = "./blog/dist/"
}

files "index" {
  path = "./index/"
}
