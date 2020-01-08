resource "aws_lb" "lb" {
  name = "embly-run"

  internal           = false
  load_balancer_type = "application"

  enable_deletion_protection = false
  subnets                    = "${module.vpc.public_subnets}"
  security_groups            = ["${aws_security_group.lb.id}"]
}

resource "aws_security_group" "lb" {
  name        = "external-elb-embly-run"
  description = "Allow traffic to external elb"
  vpc_id      = "${module.vpc.vpc_id}"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "TCP"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 443
    to_port     = 443
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

resource "aws_lb_listener" "http_listener" {
  load_balancer_arn = "${aws_lb.lb.arn}"
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "redirect"

    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

resource "aws_lb_listener" "lb" {
  load_balancer_arn = "${aws_lb.lb.arn}"
  port              = "443"
  protocol          = "HTTPS"
  ssl_policy        = "ELBSecurityPolicy-2016-08"
  certificate_arn   = "${aws_acm_certificate.embly_run.arn}"

  default_action {
    type             = "forward"
    target_group_arn = "${aws_lb_target_group.embly_run.arn}"
  }
}

resource "aws_lb_target_group" "embly_run" {
  name = "embly-run-tg"

  port     = 8080
  protocol = "HTTP"
  vpc_id   = "${module.vpc.vpc_id}"
}

resource "aws_lb_target_group_attachment" "collector_prod" {
  target_group_arn = "${aws_lb_target_group.embly_run.arn}"
  target_id        = "${aws_instance.docker_machine.id}"
  port             = 8080
}
