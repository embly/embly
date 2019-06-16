

resource "aws_elasticache_subnet_group" "embly" {
  name       = "embly"
  subnet_ids = "${module.vpc.private_subnets}"

}
resource "aws_elasticache_cluster" "build_cache" {
  subnet_group_name    = "${aws_elasticache_subnet_group.embly.name}"
  cluster_id           = "embly-build-cache"
  engine               = "redis"
  node_type            = "cache.t2.medium"
  num_cache_nodes      = 1
  engine_version       = "5.0.4"
  port                 = 6379
  az_mode              = "single-az"
  availability_zone    = "us-east-1c"
  parameter_group_name = "${aws_elasticache_parameter_group.build_cache.name}"
}

resource "aws_elasticache_parameter_group" "build_cache" {
  name   = "embly-build-cache-params"
  family = "redis5.0"

  parameter {
    name  = "maxmemory-policy"
    value = "allkeys-lfu"
  }

}

resource "aws_elasticache_cluster" "function_store" {
  subnet_group_name    = "${aws_elasticache_subnet_group.embly.name}"
  cluster_id           = "embly-function-store"
  engine               = "redis"
  node_type            = "cache.t2.medium"
  num_cache_nodes      = 1
  engine_version       = "5.0.4"
  port                 = 6379
  az_mode              = "single-az"
  availability_zone    = "us-east-1c"
  parameter_group_name = "default.redis5.0"
}
