// This example is the webserver example using the heapless httphi router
// from github.com/soypat/lneto instead of "net/http".
//
// $ curl http://10.0.0.2

//go:build ninafw || wioterminal

package main

import (
	"log"
	"machine"
	"net"
	"strconv"
	"time"

	"github.com/soypat/lneto/http/httphi"
	"github.com/soypat/lneto/http/httpraw"

	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

var (
	ssid string
	pass string
	port string = ":80"
)

const (
	numWorkers    = 2
	memoryPerConn = 1024
)

var led = machine.LED

func main() {

	// wait a bit for serial
	time.Sleep(2 * time.Second)

	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	link, _ := probe.Probe()

	err := link.NetConnect(&netlink.ConnectParams{
		Ssid:       ssid,
		Passphrase: pass,
	})
	if err != nil {
		log.Fatal(err)
	}

	var mux httphi.MuxSlice
	mux.Handle("GET /{$}", root)
	mux.Handle("GET /hello", hello)
	mux.Handle("/cnt", cnt)
	mux.Handle("GET /6", sixlines)
	mux.Handle("GET /off", LED_OFF)
	mux.Handle("GET /on", LED_ON)

	var router httphi.Router
	cfg := httphi.DefaultRouterConfig(numWorkers, memoryPerConn, mux.MaxPathValues())
	err = router.Configure(&mux, cfg)
	if err != nil {
		log.Fatal(err)
	}

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			println("accept:", err.Error())
			time.Sleep(time.Second)
			continue
		}
		if err := router.Handle(conn); err != nil {
			conn.Close()
		}
	}
}

func root(ex *httphi.Exchange) {
	access := int64(1)

	var cookie httpraw.Cookie
	var cookieBuf [128]byte
	cookie.Reset(cookieBuf[:0], 4)
	if ex.RequestParseCookie(&cookie, "Cookie") == nil {
		if v := cookie.Get("access"); v != nil {
			n, err := strconv.ParseInt(string(v), 10, 0)
			if err != nil {
				ex.RespondString(httphi.StatusBadRequest, "text/plain", "invalid cookie value")
				return
			}
			access = n + 1
		}
	}

	var setCookie [32]byte
	ex.StageHeaderBytes("Set-Cookie", strconv.AppendInt(append(setCookie[:0], "access="...), access, 10))
	ex.StageHeader("Content-Type", "text/html; charset=UTF-8")
	ex.StageHeader("Connection", "close")

	var num [20]byte
	ex.WriteBodyString(rootHead)
	ex.WriteBody(strconv.AppendInt(num[:0], access, 10))
	ex.WriteBodyString(rootTail)
}

const rootHead = `
<html>
<head>
    <title>TinyGo HTTP Server</title>
    <script language="javascript" type="text/javascript">
        var counter = 0
        function ledOn() { fetch("/on"); }
        function ledOff() { fetch("/off"); }
        function fetchCnt() { fetch("/cnt").then(response => response.json()).then(json => { counter = json.cnt; cnt.innerHTML = counter; }); }
        function incrCnt() { counter = counter + 1; fetch("/cnt?cnt=" + counter, { method: 'POST' }).then(response => response.json()).then(json => { counter = json.cnt; cnt.innerHTML = counter; }); }
        function setCnt() { fetch("/cnt", {
            method: "POST",
            body: "cnt=" + document.getElementsByName("cnt")[0].value,
            headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        }).then(response => response.json()).then(json => { counter = json.cnt; cnt.innerHTML = counter; }); return false; }
        function onLoad() { fetchCnt(); }
    </script>
</head>
<body onLoad="onLoad()">
    <h5>TinyGo HTTP Server</h5>

    <p>
        access: `

const rootTail = `
    </p>

    <a href="/hello">/hello</a><br>
    <a href="/6">/6</a><br>

    <p>
        LED<br>
        <a href="javascript:ledOn();">/on</a><br>
        <a href="javascript:ledOff();">/off</a><br>
    </p>


    <p>
        <a href="/cnt">/cnt</a><br>
        cnt: <span id="cnt"></span><br>
        <a href="javascript:incrCnt()">incrCnt()</a><br>
        <form id="form1" style="display: inline" onSubmit="return setCnt()">
        <input type="text" name="cnt">
        <input type="button" value="set cnt", onClick="setCnt()">
        </form>
    </p>
</body>
</html>
`

func sixlines(ex *httphi.Exchange) {
	// https://fukuno.jig.jp/3267
	ex.RespondString(httphi.StatusOK, "text/html; charset=UTF-8", `<body onload='onkeydown=e=>K=parseInt(e.key[5]||6,28)/3-8;Z=X=[B=A=12];Y=_=>`+
		`{for(C=[q=c=i=4];f=i--*K;c-=!Z[h+(K+6?p+K:C[i]=p*A-(p/9|0)*145)])p=B[i];for(c?0:K+6?h+=K:B=C;`+
		`i=K=q--;f+=Z[A+p])X[p=h+B[q]]=1;h+=A;if(f|B)for(Z=X,X=[l=228],B=[[-7,-20,6,h=17,-9,3,3][t=++t%7]-4,0,1,t-6?-A:2];l--;)`+
		`for(l%A?l-=l%A*!Z[l]:(P++,c=l+=A);--c>A;)Z[c]=Z[c-A];for(S="";i<240;S+=X[i]|(X[i]=Z[i]|=++i%A<2|i>228)?i%A?"■":"■<br>":"　");`+
		`D.innerHTML=S+P;setTimeout(Y,i-P)};Y(h=K=t=P=0)'id=D>`)
}

func LED_ON(ex *httphi.Exchange) {
	led.High()
	ex.RespondString(httphi.StatusOK, "text/plain; charset=UTF-8", "led.High()")
}

func LED_OFF(ex *httphi.Exchange) {
	led.Low()
	ex.RespondString(httphi.StatusOK, "text/plain; charset=UTF-8", "led.Low()")
}

func hello(ex *httphi.Exchange) {
	ex.RespondString(httphi.StatusOK, "text/plain; charset=UTF-8", "hello")
}

var counter int64

func cnt(ex *httphi.Exchange) {
	if ex.RequestMethod() == httphi.MethPost {
		var form httpraw.Form
		var formBuf [128]byte
		form.Reset(formBuf[:0], 4)
		form.EnableBufferGrowth(false)
		if err := ex.RequestParseForm(&form, true, false); err != nil {
			ex.RespondString(httphi.StatusBadRequest, "text/plain", err.Error())
			return
		}
		if c := form.Get("cnt"); len(c) > 0 {
			counter, _ = strconv.ParseInt(string(c), 0, 0)
		}
	}

	var body [32]byte
	b := append(body[:0], `{"cnt": `...)
	b = strconv.AppendInt(b, counter, 10)
	b = append(b, '}')
	ex.Respond(httphi.StatusOK, "application/json", b)
}
