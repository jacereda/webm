package main

import (
	"bufio"
	"code.google.com/p/ebml-go/webm"
	"code.google.com/p/ffvp8-go/ffvp8"
	"flag"
	"fmt"
	gl "github.com/chsc/gogl/gl21"
	"github.com/jteeuwen/glfw"
	"image"
	"io"
	"log"
	"os"
)

var in = flag.String("i", "", "Input file")
var nf = flag.Int("n", 0x7fffffff, "Number of frames")

const vss = `
void main() {
  gl_TexCoord[0] = gl_MultiTexCoord0;
  gl_Position = ftransform();
}
`

const fss = `
uniform sampler2D ytex;
uniform sampler2D cbtex;
uniform sampler2D crtex;

const mat3 ycbcr2rgb = mat3(
                          1.164, 0, 1.596,
                          1.164, -0.392, -0.813,
                          1.164, 2.017, 0.0
                          );
const float ysub = 0.0625;
void main() {
   float y = texture2D(ytex, gl_TexCoord[0].st).r;
   float cb = texture2D(cbtex, gl_TexCoord[0].st).r;
   float cr = texture2D(crtex, gl_TexCoord[0].st).r;
   vec3 ycbcr = vec3(y - ysub, cb - 0.5, cr - 0.5);
   vec3 rgb = ycbcr * ycbcr2rgb;
   gl_FragColor = vec4(rgb,1.0);
}
`

func decode(ch chan []byte, wch chan *image.YCbCr) {
	dec := ffvp8.NewDecoder()
	for data := <-ch; data != nil; data = <-ch {
		wch <- dec.Decode(data)
	}
	wch <- nil
}

func setupvp(w, h int) {
	gl.Viewport(0, 0, gl.Sizei(w), gl.Sizei(h))
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0, 1, 1, 0, -1, 1)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
}

func texinit() (yid gl.Uint) {
	gl.GenTextures(1, &yid)
	gl.BindTexture(gl.TEXTURE_2D, yid)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	return
}

func shinit(yid, cbid, crid gl.Uint) gl.Uint {
	vs := loadShader(gl.VERTEX_SHADER, vss)
	fs := loadShader(gl.FRAGMENT_SHADER, fss)
	prg := gl.CreateProgram()
	gl.AttachShader(prg, vs)
	gl.AttachShader(prg, fs)
	gl.LinkProgram(prg)
	fmt.Println(gl.GetError(), prg)
	gl.UseProgram(prg)
	fmt.Println(gl.GetError())
	return prg
}

func upload(id gl.Uint, data []byte, stride int, w int, h int) {
	gl.BindTexture(gl.TEXTURE_2D, id)
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, gl.Int(stride))
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.LUMINANCE, gl.Sizei(w), gl.Sizei(h), 0,
		gl.LUMINANCE, gl.UNSIGNED_BYTE, gl.Pointer(&data[0]))
}

func initquad() {
	ver := []gl.Float{0, 0, 1, 0, 0, 1, 1, 1}
	gl.BindBuffer(gl.ARRAY_BUFFER, 1)
	gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(4*len(ver)),
		gl.Pointer(&ver[0]), gl.STATIC_DRAW)
	gl.VertexPointer(2, gl.FLOAT, 0, nil)
	gl.BindBuffer(gl.ARRAY_BUFFER, 2)
	gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(4*len(ver)),
		gl.Pointer(&ver[0]), gl.STATIC_DRAW)
	gl.TexCoordPointer(2, gl.FLOAT, 0, nil)
	gl.EnableClientState(gl.VERTEX_ARRAY)
	gl.EnableClientState(gl.TEXTURE_COORD_ARRAY)
}

func loadShader(shtype gl.Enum, src string) gl.Uint {
	sh := gl.CreateShader(shtype)
	gsrc := gl.GLString(src)
	gl.ShaderSource(sh, 1, &gsrc, nil)
	gl.CompileShader(sh)
	return sh
}

func write(ch chan *image.YCbCr, ech chan int) {
	img := <-ch
	w := img.Rect.Dx()
	h := img.Rect.Dy()
	glfw.Init()
	defer glfw.Terminate()
	glfw.OpenWindow(w, h, 0, 0, 0, 0, 0, 0, glfw.Windowed)
	defer glfw.CloseWindow()
	glfw.SetSwapInterval(1)
	glfw.SetWindowTitle("webmplay")
	gl.Init()
	setupvp(w, h)
	yid := texinit()
	cbid := texinit()
	crid := texinit()
	shinit(yid, cbid, crid)
	initquad()
	for i := 0; img != nil; i, img = i+1, <-ch {
		gl.ActiveTexture(gl.TEXTURE0)
		gl.Enable(gl.TEXTURE_2D)
		upload(yid, img.Y, img.YStride, w, h)
		gl.Uniform1i(0, 0)

		gl.ActiveTexture(gl.TEXTURE1)
		gl.Enable(gl.TEXTURE_2D)
		upload(cbid, img.Cb, img.CStride, w/2, h/2)
		gl.Uniform1i(1, 1)

		gl.ActiveTexture(gl.TEXTURE2)
		gl.Enable(gl.TEXTURE_2D)
		upload(crid, img.Cr, img.CStride, w/2, h/2)
		gl.Uniform1i(2, 2)
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
		gl.Flush()
		glfw.SwapBuffers()
		glfw.Sleep(0.001)
	}
	ech <- 1
}

func main() {
	flag.Parse()
	r, err := os.Open(*in)
	if err != nil {
		log.Panic("unable to open file " + *in)
	}
	br := bufio.NewReader(r)
	var wm webm.WebM
	e, rest, err := webm.Parse(br, &wm)
	track := wm.FindFirstVideoTrack()
	dchan := make(chan []byte, 16)
	wchan := make(chan *image.YCbCr, 16)
	echan := make(chan int, 1)
	go decode(dchan, wchan)
	go write(wchan, echan)
	for i := 0; err == nil && i < *nf; {
		t := make([]byte, 4)
		io.ReadFull(e.R, t)
		if uint(t[0])&0x7f == track.TrackNumber {
			data := make([]byte, e.Size())
			io.ReadFull(e.R, data)
			dchan <- data
			i++
		}
		_, err = e.ReadData()
		e, err = rest.Next()
	}
	dchan <- nil
	<-echan
}
