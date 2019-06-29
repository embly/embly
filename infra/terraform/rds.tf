
# resource "aws_db_subnet_group" "embly" {
#   name       = "embly"
#   subnet_ids = "${module.vpc.private_subnets}"
# }
# resource "aws_db_instance" "db" {
#   allocated_storage      = 20
#   db_subnet_group_name   = "${aws_db_subnet_group.embly.name}"
#   engine                 = "postgres"
#   engine_version         = "9.6.10"
#   identifier             = "embly"
#   instance_class         = "db.t3.medium"
#   name                   = "embly"
#   password               = "changeme"
#   publicly_accessible    = false
#   skip_final_snapshot    = true
#   storage_type           = "gp2"
#   username               = "embly"
#   vpc_security_group_ids = ["${module.vpc.default_security_group_id}"]
# }
