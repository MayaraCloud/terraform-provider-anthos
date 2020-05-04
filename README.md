
Terraform Provider for Google Anthos
==================

- Website: [mayara.io](https://www.mayara.io)
- Documentation: [mayara.io/terraform-provider-anthos](https://mayara.io/terraform-provider-anthos)
<img src="https://mayara.io/images/mayara_logo.svg" width="350">

Maintainers
-----------

This provider plugin is maintained by:

- The [Mayara Cloud Team](https://mayara.io)

Requirements
------------

- [Terraform](https://www.terraform.io/downloads.html) 0.12+

Using the provider
----------------------

See the [Anthos Provider documentation](https://www.mayara.io/terraform-provider-anthos/docs/index.html) to get started using the
Anthos provider.

**The provider is still in Alpha stage.**

Building the provider
---------------------

Clone repository to: `$GOPATH/src/github.com/MayaraCloud/terraform-provider-anthos`

```sh
mkdir -p $GOPATH/github.com/MayaraCloud; cd $GOPATH/src/github.com/MayaraCloud
git clone git@github.com:MayaraCloud/terraform-provider-anthos.git
```

Enter the provider directory and build the provider

```sh
cd $GOPATH/src/github.com/MayaraCloud/terraform-provider-anthos
make
```

Developing the provider
---------------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org)
installed on your machine (version 1.13.0+ is *required*). You can use [goenv](https://github.com/syndbg/goenv)
to manage your Go version. You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH),
as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`.
This will build the provider and put the provider binary in the `$GOPATH/bin`
directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-google
...
```
