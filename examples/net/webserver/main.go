// This example listens on port :80 serving a web page.  Multiple clients
// may connect and be serviced at the same time.  IPv4 only.  HTTP only.
//
// $ curl http://10.0.0.2
//
// Note: It may be necessary to increase the stack size when using "net/http".
// Use the -stack-size=4KB command line option.

//go:build ninafw || wioterminal

package main

import (
	"fmt"
	"log"
	"machine"
	"net/http"
	"strconv"
	"time"

	"tinygo.org/x/drivers/netlink"
	"tinygo.org/x/drivers/netlink/probe"
)

var (
	ssid string
	pass string
	port string = ":80"
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

	http.HandleFunc("/", root)
	http.HandleFunc("/hello", hello)
	http.HandleFunc("/cnt", cnt)
	http.HandleFunc("/6", sixlines)
	http.HandleFunc("/off", LED_OFF)
	http.HandleFunc("/on", LED_ON)

	err = http.ListenAndServe(port, nil)
	for err != nil {
		fmt.Printf("error: %s\r\n", err.Error())
		time.Sleep(5 * time.Second)
	}
}

func root(w http.ResponseWriter, r *http.Request) {
	access := 1

	cookie, err := r.Cookie("access")
	if err != nil {
		if err == http.ErrNoCookie {
			cookie = &http.Cookie{
				Name:  "access",
				Value: "1",
			}
		} else {
			http.Error(w, fmt.Sprintf("%s", err.Error()), http.StatusBadRequest)
			return
		}
	} else {
		v, err := strconv.ParseInt(cookie.Value, 10, 0)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid cookie.Value : %s", cookie.Value), http.StatusBadRequest)
			return
		}
		cookie.Value = fmt.Sprintf("%d", v+1)
		access = int(v) + 1
	}
	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, `
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
        access: %d
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
`, access)
}

func sixlines(w http.ResponseWriter, r *http.Request) {
	// https://fukuno.jig.jp/3267
	fmt.Fprint(w, `<body onload='onkeydown=e=>K=parseInt(e.key[5]||6,28)/3-8;Z=X=[B=A=12];Y=_=>`+
		`{for(C=[q=c=i=4];f=i--*K;c-=!Z[h+(K+6?p+K:C[i]=p*A-(p/9|0)*145)])p=B[i];for(c?0:K+6?h+=K:B=C;`+
		`i=K=q--;f+=Z[A+p])X[p=h+B[q]]=1;h+=A;if(f|B)for(Z=X,X=[l=228],B=[[-7,-20,6,h=17,-9,3,3][t=++t%7]-4,0,1,t-6?-A:2];l--;)`+
		`for(l%A?l-=l%A*!Z[l]:(P++,c=l+=A);--c>A;)Z[c]=Z[c-A];for(S="";i<240;S+=X[i]|(X[i]=Z[i]|=++i%A<2|i>228)?i%A?"■":"■<br>":"　");`+
		`D.innerHTML=S+P;setTimeout(Y,i-P)};Y(h=K=t=P=0)'id=D>`)
}

func LED_ON(w http.ResponseWriter, r *http.Request) {
	led.High()
	w.Header().Set(`Content-Type`, `text/plain; charset=UTF-8`)
	fmt.Fprintf(w, "led.High()")
}

func LED_OFF(w http.ResponseWriter, r *http.Request) {
	led.Low()
	w.Header().Set(`Content-Type`, `text/plain; charset=UTF-8`)
	fmt.Fprintf(w, "led.Low()")
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(`Content-Type`, `text/plain; charset=UTF-8`)
	fmt.Fprintf(w, "hello")
}

var counter int

func cnt(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	if r.Method == "POST" {
		c := r.Form.Get("cnt")
		if c != "" {
			i64, _ := strconv.ParseInt(c, 0, 0)
			counter = int(i64)
		}
	}

	w.Header().Set(`Content-Type`, `application/json`)
	fmt.Fprintf(w, `{"cnt": %d}`, counter)
}
