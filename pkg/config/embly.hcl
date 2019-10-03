
dependencies = [
  "embly/http",
  "embly/tcp",
  "embly/vinyl",
  "embly/ssh"
]

function "encoder" {
  path    = "./examples/mjpeg/encoder"
  context = "../../.."
  runtime = "rust"

  # service_definition = "${schema.proto.EncoderService}"
}


gateway {
  type = "http"
  port = 8080

  function = "${function.encoder}"
  route "/" {
    function = "${function.encoder}"
  }
  route "/foo" {
    function = "${function.encoder}"
  }
  route "/bundle" {
    files = "${files.assets}"
  }
}

# include files with the build to be made available through http
files "assets" {
  path = "./"
}

database "vinyl" "main" {
  definition = "things.proto"

  record "User" {
    primary_key = "id"
    index "email" {
      unique = true
    }
  }

}

