provider "etcd" {
  username      = var.username    # optionally use ETCD_USERNAME env var
  password      = var.password    # optionally use ETCD_PASSWORD env var
  etcd_endpoint = var.etcd_ep     # optionally use ETCD_ENDPOINT env var

  # The provider will connect using a tls session. But for some weird reason 
  # you decide skip that you can set tls to false
  # tls           = var.tls         # optionally use ETCD_TLS env var
  # ca_cert       = var.ca_cert     # optionally use ETCD_CACERT env var
}
