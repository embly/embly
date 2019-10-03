
function "hello" {
  path    = "./examples/project/hello"
  context = "./../.."
  runtime = "rust"
}

function "echo" {
  path    = "./examples/project/echo"
  context = "./../.."
  runtime = "rust"
}
