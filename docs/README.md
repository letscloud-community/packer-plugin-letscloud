### Installation

To install this plugin, copy and paste this code into your Packer configuration, then run [`packer init`](https://www.packer.io/docs/commands/init).

```hcl
packer {
  required_plugins {
    letscloud = {
      source  = "github.com/letscloud-community/letscloud"
      version = ">=0.1.0"
    }
  }
}
```

Alternatively, you can use `packer plugins install` to manage installation of this plugin.

```sh
$ packer plugins install github.com/letscloud-community/letscloud
```

### Components

#### Builders

- [lestcloud](/packer/integrations/hashicorp/letscloud/latest/components/builder/letscloud) - The letscloud builder is used to create custom snapshot to be reusable image.
