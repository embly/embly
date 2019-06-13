
# resource "aws_elasticache_cluster" "embly_cache" {
#   cluster_id           = "embly-cache"
#   engine               = "memcached"
#   node_type            = "cache.t2.medium"
#   num_cache_nodes      = 1
#   parameter_group_name = "default.memcached1.5"
#   port                 = 11211
#   az_mode = "single-az"
#   availability_zone = "us-east-1a"
# }
