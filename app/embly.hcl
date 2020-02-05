
function "auth" {
  runtime = "rust"
  path    = "./auth"
  sources = [
    "../embly-rs",
  ]
}

function "hello" {
  runtime = "rust"
  path    = "./hello"
  sources = [
    "../embly-rs",
  ]
}


gateway {
  type = "http"
  port = 8082

  route "/" {
    files = "${files.blog}"
  }

  route "/app/" {
    files = "${files.frontend}"
  }
  route "/api/auth/" {
    function = "${function.auth}"
  }
  route "/hello/" {
    function = "${function.hello}"
  }

  route "/star-playground" {
    files = "${files.star-playground}"
  }
}

files "star-playground" {
  path = "./star-playground"
}

files "frontend" {
  path              = "./frontend/build/"
  local_file_server = "http://localhost:3000"
}


files "blog" {
  path = "./blog/dist/"
}
