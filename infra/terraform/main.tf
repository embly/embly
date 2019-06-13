terraform {
  backend "s3" {
    bucket = "embly-tfstate"
    key    = "embly-terraform.tfstate"
    profile = "max"
    region = "us-east-1"
  }
}

provider "aws" {
  region                  = "us-east-1"
  shared_credentials_file = "~/.aws/credentials"
  profile                 = "max"
  version                 = "~> 2.7.0"
}

resource "aws_s3_bucket" "embly_tfstate" {
  bucket = "embly-tfstate"
  acl    = "private"

  versioning {
    enabled = true
  }
}


resource "aws_acm_certificate" "embly_run" {
  domain_name               = "embly.run"
  subject_alternative_names = ["*.embly.run"]
  validation_method         = "EMAIL"
}

resource "aws_acm_certificate" "embly_org" {
  domain_name               = "embly.org"
  subject_alternative_names = ["*.embly.org"]
  validation_method         = "EMAIL"
}

