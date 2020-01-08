resource "aws_instance" "docker_machine" {
  ami           = "ami-927185ef"
  instance_type = "t3.medium"

  tags = {
    Name = "embly-run"
  }
}

resource "aws_security_group" "docker_machine" {
  name        = "docker-machine"
  description = "Docker Machine"
  vpc_id      = "${module.vpc.vpc_id}"

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "TCP"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "TCP"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 2376
    to_port     = 2376
    protocol    = "TCP"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port = 0
    to_port   = 0

    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
