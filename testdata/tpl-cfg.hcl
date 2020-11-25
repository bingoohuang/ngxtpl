# https://learn.hashicorp.com/tutorials/consul/load-balancing-nginx
# https://github.com/hashicorp/hcl

consul {
  address = "localhost:8500"

  retry {
    enabled = true
    attempts = 12
    backoff = "250ms"
  }
}

template {
  source = "/etc/nginx/conf.d/load-balancer.conf.tpl"
  destination = "/etc/nginx/conf.d/load-balancer.conf"
  perms = 0600
  command = "service nginx reload"
}

