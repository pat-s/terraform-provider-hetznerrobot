resource "hetznerrobot_vswitch" "example" {
  name    = "example-vswitch"
  vlan_id = 100

  server {
    server_ip = "1.1.1.1"
  }
}
