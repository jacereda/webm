package main

import (
	"code.google.com/p/ebml-go/common"
	gl "github.com/chsc/gogl/gl21"
	"github.com/jteeuwen/glfw"
	"image"
	"runtime"
)

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

func texinit(id gl.Uint) {
	gl.BindTexture(gl.TEXTURE_2D, id)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	return
}

func shinit() {
	vs := loadShader(gl.VERTEX_SHADER, vss)
	fs := loadShader(gl.FRAGMENT_SHADER, fss)
	prg := gl.CreateProgram()
	gl.AttachShader(prg, vs)
	gl.AttachShader(prg, fs)
	gl.LinkProgram(prg)
	gl.UseProgram(prg)
	gl.Uniform1i(0, 0)
	gl.Uniform1i(1, 1)
	gl.Uniform1i(2, 2)
}

func upload(id gl.Uint, data []byte, stride int, w int, h int) {
	gl.BindTexture(gl.TEXTURE_2D, id)
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, gl.Int(stride))
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.LUMINANCE, gl.Sizei(w), gl.Sizei(h), 0,
		gl.LUMINANCE, gl.UNSIGNED_BYTE, gl.Pointer(&data[0]))
}

func initquad() {
	ver := []gl.Float{-1, 1, 1, 1, -1, -1, 1, -1}
	gl.BindBuffer(gl.ARRAY_BUFFER, 1)
	gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(4*len(ver)),
		gl.Pointer(&ver[0]), gl.STATIC_DRAW)
	gl.VertexPointer(2, gl.FLOAT, 0, nil)
	tex := []gl.Float{0, 0, 1, 0, 0, 1, 1, 1}
	gl.BindBuffer(gl.ARRAY_BUFFER, 2)
	gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(4*len(tex)),
		gl.Pointer(&tex[0]), gl.STATIC_DRAW)
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

func setupvp(w, h int) {
	gl.Viewport(0, 0, gl.Sizei(w), gl.Sizei(h))
}

func write(wchan <-chan *image.YCbCr) {
	img := <-wchan
	w := img.Rect.Dx()
	h := img.Rect.Dy()
	gl.Init()
	glfw.Init()
	defer glfw.Terminate()
	glfw.OpenWindow(w, h, 0, 0, 0, 0, 0, 0, glfw.Windowed)
	defer glfw.CloseWindow()
	glfw.SetWindowSizeCallback(setupvp)
	glfw.SetSwapInterval(1)
	glfw.SetWindowTitle("webmplay")
	texinit(1)
	texinit(2)
	texinit(3)
	shinit()
	initquad()
	gl.Enable(gl.TEXTURE_2D)
	for ; glfw.WindowParam(glfw.Opened) == 1 && img != nil; img = <-wchan {
		gl.ActiveTexture(gl.TEXTURE0)
		upload(1, img.Y, img.YStride, w, h)
		gl.ActiveTexture(gl.TEXTURE1)
		upload(2, img.Cb, img.CStride, w/2, h/2)
		gl.ActiveTexture(gl.TEXTURE2)
		upload(3, img.Cr, img.CStride, w/2, h/2)
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
		runtime.GC()
		glfw.SwapBuffers()
	}
}

func main() {
	common.Main(write)
}
