terraform {
  required_version = "~> 1.3.3"
  required_providers {
    etcd = {
      source  = "github.com/qurname2/terraform-provider-etcd"
      version = "0.1.2"
    }
  }
  backend "s3" {
    bucket         = ""
    key            = ""
    region         = ""
    dynamodb_table = ""
    profile        = ""
  }
}
