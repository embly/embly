resource "aws_acm_certificate" "embly_run" {
  domain_name               = "embly.run"
  subject_alternative_names = ["*.embly.run"]
  validation_method         = "EMAIL"
}
