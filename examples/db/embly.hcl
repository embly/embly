
function "listener" {
  runtime = "rust"
  context = "./../../.."
  path    = "./embly/examples/db/listener"
}

gateway {
  type = "http"
  port = 8082
  route "/fortest/" {
    files = "${files.index}"
  }
  route "/api/" {
    function = "${function.listener}"
  }
}

files "index" {
  path = "./assets"
}


database "vinyl" "main" {
  definition = "data.proto"
  record "User" {
    primary_key = "id"
    index "email" {
      unique = "true"
    }
    index "username" {
      unique = "true"
    }
  }
}
