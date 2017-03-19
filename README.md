## README
Imageserver is a simple [Go](https://golang.org/) http server for basic image processing based on [fasthttp](https://github.com/valyala/fasthttp), and [bimg](https://github.com/h2non/bimg) (extending [libvips](https://github.com/jcupitt/libvips)).

## REQUIREMENTS
On OSX:
```
brew update
brew install vips --with-webp
```

Tested on:
```
$ uname -a
Darwin smx 15.6.0 Darwin Kernel Version 15.6.0: Mon Jan  9 23:07:29 PST 2017; root:xnu-3248.60.11.2.1~1/RELEASE_X86_64 x86_64

$ go version
go version go1.8 darwin/amd64

$ vips -v
vips-8.4.5-Sun Mar 19 07:06:21 WIB 2017

---
Machine: MacBook Pro (Retina, 15-inch, Mid 2014)
Processor: 2,5 GHz Intel Core i7
Memory: 16 GB 1600 MHz DDR3
```

## INSTALL
```
$ go get -u -v github.com/simukti/imageserver
```

## USAGE
```
$ imageserver
[ERROR] Flag param -s is required
Usage:
  -h string
    	Host port to serve this app. (default "127.0.0.1:8080")
  -m string
    	Directory for master image storage. (default "/tmp/imgsrv_master")
  -r string
    	Directory for result image storage. (default "/tmp/imgsrv_result")
  -s string
    	Source server base URL. (Example: https://kadalkesit.storage.googleapis.com)
  -t int
    	Process timeout per image processing request. (default 30)
```

## SUPPORTED IMAGE PARAMETERS

- **w**

    Output width (in pixel)

- **h**

    Output height (in pixel)

- **q**
    
    Output quality (JPG default 75)

- **fmt**

    Output format (default: same as source)

- **blur**

    Output image blur level (1-50)

- **c**

    Output colour space (srgb, bw)

- **flip**

    Output image flip (h : flip horizontally, v : flip vertically)


## SAMPLE REQUEST
I use [my photo from flickr](https://www.flickr.com/photos/simukti/8045877062/).

```
$ imageserver -s https://c1.staticflickr.com
```

- Original flickr: `https://c1.staticflickr.com/9/8173/8045877062_481f4e80b4_b.jpg`

Size: 313.95 KB (321485 bytes)

- Parsed as jpg (default): `http://127.0.0.1:8080/9/8173/8045877062_481f4e80b4_b.jpg`

Size: 137.58 KB (140882 bytes)

- Parsed as webp: `http://127.0.0.1:8080/9/8173/8045877062_481f4e80b4_b.jpg?fmt=webp`

Size: 131.63 KB (134786 bytes)

- Parsed as png: `http://127.0.0.1:8080/9/8173/8045877062_481f4e80b4_b.jpg?fmt=png`

Size: 1,587.41 KB (1,625,510 bytes) !!!

## NOTE
Source image file will be downloaded once and saved to master source folder. 
Result file will be saved to result folder using filename based on hashed request params. 
If requested params already exists that file will be returned without any further processing.

Is it fast ?? hmmm... I think it depends on hardware configuration ;) try it yourself. 

## LICENSE
This project is released under the MIT licence. See the LICENSE.md file for more.