
resource "aws_s3_bucket" "embly_static_files" {
  bucket = "embly-static"
  acl    = "public-read"
}
