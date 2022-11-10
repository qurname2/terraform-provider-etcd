provider "etcd" {
  username  = var.username
  password  = var.password
  endpoints = "https://etcd-server:2379"
  tls       = true
  ca_cert   = var.ca_cert
}
