go-tdtool-api
=============
A simple API in front of the TellStick tdtool. (Written in [Go](http://golang.org/))

## tdtool

The tdtool binary is installed when installing the [TelldusCenter](http://www.telldus.se/products/nativesoftware).

You can also find it in the [telldus-core](https://github.com/telldus/telldus/tree/master/telldus-core) package.
_(You probably don’t need to install the GUI)_

## TellStick

You will also need one of these:

[![TellStick](http://www.telldus.se/img/img_start_product_tellstick.jpg)](http://www.telldus.se/products/tellstick)

And probably one or two controllable devices. (I’ve got the [Nexa PB-3](http://www.nexa.se/PB3Ny3packsjalvlarande.htm))

## Running in the background

    nohup ./tdtool-api &

This will try to start a web server on port **8080**

## Using the API

### Listing available devices
```ruby
curl http://localhost:8080/
```
#### This should output something along these lines:

    Number of devices: 4
    1    Lights          ON
    3    Hörnlampa       ON
    2    Skrivbordslampa OFF
    4    Sovrumslampa    ON

### Turning a device _ON_
```ruby
curl -X PUT http://localhost:8080/2/on
```

This should output:

    Turning on device 2, Skrivbordslampa - Success

### Turning a device _OFF_
```ruby
curl -X PUT http://localhost:8080/4/off
```

This should output:

    Turning off device 4, Sovrumslampa - Success

### Sync calls

You can also make synchronous requests by adding `/sync` to the end of the path:

```
curl -X PUT http://localhost:8080/3/off/sync
```
