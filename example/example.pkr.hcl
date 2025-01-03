packer {
  required_plugins {
    letscloud = {
      version = ">= 0.0.1"
      source  = "github.com/letscloud-community/letscloud"
    }
  }
}

source "letscloud" "hello" {
  token  = "seu_token_aqui"
  region = "br-sao"
}

build {
  sources = ["source.letscloud.hello"]
}