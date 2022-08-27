Benthos with tuya inut plugin
======================

## Build

```sh
go build
```

Alternatively build it as a Docker image with:

```sh
go mod vendor
docker build . -t benthos_tuya_input
```

## Usage

```yaml
input:
  tuya:
    # required
    accessId: abcd
    # required
    accessSecert: yueyqueqwyiyuyuiy
    # (optional) default value is false
    debug: false
```

And you can run it like this:

```sh
./benthos_tuya_input -c ./yourconfig.yaml
```
