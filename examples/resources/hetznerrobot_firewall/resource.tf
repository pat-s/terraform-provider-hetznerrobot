resource "hetznerrobot_firewall" "example" {
  server_ip     = "1.1.1.1"
  active        = true
  whitelist_hos = true

  rule {
    name       = "Allow SSH"
    src_ip     = "0.0.0.0/0"
    src_port   = "0-65535"
    dst_ip     = "0.0.0.0/0"
    dst_port   = "22"
    protocol   = "tcp"
    tcp_flags  = "syn"
    action     = "accept"
    ip_version = "ipv4"
  }

  rule {
    name       = "Allow ICMP"
    src_ip     = "0.0.0.0/0"
    dst_ip     = "0.0.0.0/0"
    protocol   = "icmp"
    action     = "accept"
    ip_version = "ipv4"
  }
}
